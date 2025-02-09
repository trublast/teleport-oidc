/*
Copyright 2019 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auth

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/beevik/etree"
	"github.com/google/go-cmp/cmp"
	"github.com/gravitational/trace"
	saml2 "github.com/russellhaering/gosaml2"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/constants"
	apidefaults "github.com/gravitational/teleport/api/defaults"
	"github.com/gravitational/teleport/api/types"
	apievents "github.com/gravitational/teleport/api/types/events"
	"github.com/gravitational/teleport/api/utils/keys"
	"github.com/gravitational/teleport/lib/authz"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/events"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/services/local"
	"github.com/gravitational/teleport/lib/utils"
)

// SAMLService are the methods that the auth server delegates to a plugin for
// implementing the SAML connector. The Identity service contains a few more
// methods for getting connectors and requests from the backend, but there is
// no specific logic related to SAML in those methods.
type SAMLService interface {
	// UpsertSAMLConnector updates or creates SAML connector
	UpsertSAMLConnector(ctx context.Context, connector types.SAMLConnector) error
	// DeleteSAMLConnector deletes SAML connector by ID
	DeleteSAMLConnector(ctx context.Context, connectorID string) error
	// CreateSAMLAuthRequest creates SAML AuthnRequest
	CreateSAMLAuthRequest(ctx context.Context, req types.SAMLAuthRequest) (*types.SAMLAuthRequest, error)
	// ValidateSAMLResponse validates SAML auth response
	ValidateSAMLResponse(ctx context.Context, re string, connectorID string) (*SAMLAuthResponse, error)
}

// UpsertSAMLConnector delegates the method call to the samlAuthService if present,
// or returns a NotImplemented error if not present.
func (a *Server) UpsertSAMLConnector(ctx context.Context, connector types.SAMLConnector) error {
	if a.samlAuthService == nil {
		return trace.NotImplemented("SAML is only available in enterprise subscriptions")
	}

	return trace.Wrap(a.samlAuthService.UpsertSAMLConnector(ctx, connector))
}

// DeleteSAMLConnector delegates the method call to the samlAuthService if present,
// or returns a NotImplemented error if not present.
func (a *Server) DeleteSAMLConnector(ctx context.Context, connectorID string) error {
	if a.samlAuthService == nil {
		return trace.NotImplemented("SAML is only available in enterprise subscriptions")
	}

	return trace.Wrap(a.samlAuthService.DeleteSAMLConnector(ctx, connectorID))
}

// CreateSAMLAuthRequest delegates the method call to the samlAuthService if present,
// or returns a NotImplemented error if not present.
func (a *Server) CreateSAMLAuthRequest(ctx context.Context, req types.SAMLAuthRequest) (*types.SAMLAuthRequest, error) {
	if a.samlAuthService == nil {
		return nil, trace.NotImplemented("SAML is only available in enterprise subscriptions")
	}

	rq, err := a.samlAuthService.CreateSAMLAuthRequest(ctx, req)
	return rq, trace.Wrap(err)
}

// ValidateSAMLResponse delegates the method call to the samlAuthService if present,
// or returns a NotImplemented error if not present.
func (a *Server) ValidateSAMLResponse(ctx context.Context, re string, connectorID string) (*SAMLAuthResponse, error) {
	if a.samlAuthService == nil {
		return nil, trace.NotImplemented("SAML is only available in enterprise subscriptions")
	}

	resp, err := a.samlAuthService.ValidateSAMLResponse(ctx, re, connectorID)
	return resp, trace.Wrap(err)
}

// SAMLAuthService implements the logic of the SAML connector, allowing SSO
// logins using the SAML protocol.
//
// SAMLAuthService implements the SAMLService interface.
type SAMLAuthService struct {
	auth                   *Server
	emitter                apievents.Emitter
	assertionReplayService *local.AssertionReplayService
	samlProviders          map[string]*samlProvider
	lock                   sync.Mutex
}

type SAMLAuthServiceConfig struct {
	Auth                   *Server
	Emitter                apievents.Emitter
	AssertionReplayService *local.AssertionReplayService
}

// NewSAMLAuthService returns a SAMLAuthService configured to use the
// services given in the config.
func NewSAMLAuthService(cfg *SAMLAuthServiceConfig) *SAMLAuthService {
	return &SAMLAuthService{
		auth:                   cfg.Auth,
		emitter:                cfg.Emitter,
		assertionReplayService: cfg.AssertionReplayService,

		samlProviders: make(map[string]*samlProvider),
	}
}

// samlProvider is internal structure that stores SAML client and its config
type samlProvider struct {
	provider  *saml2.SAMLServiceProvider
	connector types.SAMLConnector
}

// ErrSAMLNoRoles results from not mapping any roles from SAML claims.
var ErrSAMLNoRoles = trace.AccessDenied("No roles mapped from claims. The mappings may contain typos.")

// UpsertSAMLConnector creates or updates a SAML connector.
func (sas *SAMLAuthService) UpsertSAMLConnector(ctx context.Context, connector types.SAMLConnector) error {
	if err := services.ValidateSAMLConnector(connector, sas.auth); err != nil {
		return trace.Wrap(err)
	}
	if err := sas.auth.Services.UpsertSAMLConnector(ctx, connector); err != nil {
		return trace.Wrap(err)
	}
	if err := sas.emitter.EmitAuditEvent(ctx, &apievents.SAMLConnectorCreate{
		Metadata: apievents.Metadata{
			Type: events.SAMLConnectorCreatedEvent,
			Code: events.SAMLConnectorCreatedCode,
		},
		UserMetadata: authz.ClientUserMetadata(ctx),
		ResourceMetadata: apievents.ResourceMetadata{
			Name: connector.GetName(),
		},
	}); err != nil {
		log.WithError(err).Warn("Failed to emit SAML connector create event.")
	}

	return nil
}

// DeleteSAMLConnector deletes a SAML connector by name.
func (sas *SAMLAuthService) DeleteSAMLConnector(ctx context.Context, connectorName string) error {
	if err := sas.auth.Services.DeleteSAMLConnector(ctx, connectorName); err != nil {
		return trace.Wrap(err)
	}
	if err := sas.emitter.EmitAuditEvent(ctx, &apievents.SAMLConnectorDelete{
		Metadata: apievents.Metadata{
			Type: events.SAMLConnectorDeletedEvent,
			Code: events.SAMLConnectorDeletedCode,
		},
		UserMetadata: authz.ClientUserMetadata(ctx),
		ResourceMetadata: apievents.ResourceMetadata{
			Name: connectorName,
		},
	}); err != nil {
		log.WithError(err).Warn("Failed to emit SAML connector delete event.")
	}

	return nil
}

func (sas *SAMLAuthService) CreateSAMLAuthRequest(ctx context.Context, req types.SAMLAuthRequest) (*types.SAMLAuthRequest, error) {
	connector, provider, err := sas.getSAMLConnectorAndProvider(ctx, req)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	doc, err := provider.BuildAuthRequestDocument()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	attr := doc.Root().SelectAttr("ID")
	if attr == nil || attr.Value == "" {
		return nil, trace.BadParameter("missing auth request ID")
	}

	req.ID = attr.Value

	// Workaround for Ping: Ping expects `SigAlg` and `Signature` query
	// parameters when "Enforce Signed Authn Request" is enabled, but gosaml2
	// only provides these parameters when binding == BindingHttpRedirect.
	// Luckily, BuildAuthURLRedirect sets this and is otherwise identical to
	// the standard BuildAuthURLFromDocument.
	if connector.GetProvider() == teleport.Ping {
		req.RedirectURL, err = provider.BuildAuthURLRedirect("", doc)
	} else {
		req.RedirectURL, err = provider.BuildAuthURLFromDocument("", doc)
	}

	if err != nil {
		return nil, trace.Wrap(err)
	}

	err = sas.auth.Services.CreateSAMLAuthRequest(ctx, req, defaults.SAMLAuthRequestTTL)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &req, nil
}

func (sas *SAMLAuthService) getSAMLConnectorAndProviderByID(ctx context.Context, connectorID string) (types.SAMLConnector, *saml2.SAMLServiceProvider, error) {
	connector, err := sas.auth.Identity.GetSAMLConnector(ctx, connectorID, true)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}
	provider, err := sas.getSAMLProvider(connector)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	return connector, provider, nil
}

func (sas *SAMLAuthService) getSAMLConnectorAndProvider(ctx context.Context, req types.SAMLAuthRequest) (types.SAMLConnector, *saml2.SAMLServiceProvider, error) {
	if req.SSOTestFlow {
		if req.ConnectorSpec == nil {
			return nil, nil, trace.BadParameter("ConnectorSpec cannot be nil when SSOTestFlow is true")
		}

		if req.ConnectorID == "" {
			return nil, nil, trace.BadParameter("ConnectorID cannot be empty")
		}

		// stateless test flow
		connector, err := types.NewSAMLConnector(req.ConnectorID, *req.ConnectorSpec)
		if err != nil {
			return nil, nil, trace.Wrap(err)
		}

		// validate, set defaults for connector
		err = services.ValidateSAMLConnector(connector, sas.auth)
		if err != nil {
			return nil, nil, trace.Wrap(err)
		}

		// we don't want to cache the provider. construct it directly instead of using sas.getSAMLProvider()
		provider, err := services.GetSAMLServiceProvider(connector, sas.auth.GetClock())
		if err != nil {
			return nil, nil, trace.Wrap(err)
		}

		return connector, provider, nil
	}

	// regular execution flow
	return sas.getSAMLConnectorAndProviderByID(ctx, req.ConnectorID)
}

func (sas *SAMLAuthService) getSAMLProvider(conn types.SAMLConnector) (*saml2.SAMLServiceProvider, error) {
	sas.lock.Lock()
	defer sas.lock.Unlock()

	providerPack, ok := sas.samlProviders[conn.GetName()]
	if ok && cmp.Equal(providerPack.connector, conn) {
		return providerPack.provider, nil
	}
	delete(sas.samlProviders, conn.GetName())

	serviceProvider, err := services.GetSAMLServiceProvider(conn, sas.auth.GetClock())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	sas.samlProviders[conn.GetName()] = &samlProvider{connector: conn, provider: serviceProvider}

	return serviceProvider, nil
}

func (sas *SAMLAuthService) calculateSAMLUser(diagCtx *SSODiagContext, connector types.SAMLConnector, assertionInfo saml2.AssertionInfo, request *types.SAMLAuthRequest) (*CreateUserParams, error) {
	p := CreateUserParams{
		ConnectorName: connector.GetName(),
		Username:      assertionInfo.NameID,
	}

	p.Traits = services.SAMLAssertionsToTraits(assertionInfo)

	diagCtx.Info.SAMLTraitsFromAssertions = p.Traits
	diagCtx.Info.SAMLConnectorTraitMapping = connector.GetTraitMappings()

	var warnings []string
	warnings, p.Roles = services.TraitsToRoles(connector.GetTraitMappings(), p.Traits)
	if len(p.Roles) == 0 {
		if len(warnings) != 0 {
			log.WithField("connector", connector).Warnf("No roles mapped from claims. Warnings: %q", warnings)
			diagCtx.Info.SAMLAttributesToRolesWarnings = &types.SSOWarnings{
				Message:  "No roles mapped for the user",
				Warnings: warnings,
			}
		} else {
			log.WithField("connector", connector).Warnf("No roles mapped from claims.")
			diagCtx.Info.SAMLAttributesToRolesWarnings = &types.SSOWarnings{
				Message: "No roles mapped for the user. The mappings may contain typos.",
			}
		}
		return nil, trace.Wrap(ErrSAMLNoRoles)
	}

	// Pick smaller for role: session TTL from role or requested TTL.
	roles, err := services.FetchRoles(p.Roles, sas.auth, p.Traits)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	roleTTL := roles.AdjustSessionTTL(apidefaults.MaxCertDuration)

	if request != nil {
		p.SessionTTL = utils.MinTTL(roleTTL, request.CertTTL)
	} else {
		p.SessionTTL = roleTTL
	}

	return &p, nil
}

func (sas *SAMLAuthService) createSAMLUser(p *CreateUserParams, dryRun bool) (types.User, error) {
	expires := sas.auth.GetClock().Now().UTC().Add(p.SessionTTL)

	log.Debugf("Generating dynamic SAML identity %v/%v with roles: %v. Dry run: %v.", p.ConnectorName, p.Username, p.Roles, dryRun)

	user := &types.UserV2{
		Kind:    types.KindUser,
		Version: types.V2,
		Metadata: types.Metadata{
			Name:      p.Username,
			Namespace: apidefaults.Namespace,
			Expires:   &expires,
		},
		Spec: types.UserSpecV2{
			Roles:  p.Roles,
			Traits: p.Traits,
			SAMLIdentities: []types.ExternalIdentity{
				{
					ConnectorID: p.ConnectorName,
					Username:    p.Username,
				},
			},
			CreatedBy: types.CreatedBy{
				User: types.UserRef{
					Name: teleport.UserSystem,
				},
				Time: sas.auth.GetClock().Now().UTC(),
				Connector: &types.ConnectorRef{
					Type:     constants.SAML,
					ID:       p.ConnectorName,
					Identity: p.Username,
				},
			},
		},
	}

	if dryRun {
		return user, nil
	}

	// Get the user to check if it already exists or not.
	existingUser, err := sas.auth.Services.GetUser(p.Username, false)
	if err != nil && !trace.IsNotFound(err) {
		return nil, trace.Wrap(err)
	}

	ctx := context.TODO()

	// Overwrite exisiting user if it was created from an external identity provider.
	if existingUser != nil {
		connectorRef := existingUser.GetCreatedBy().Connector

		// If the exisiting user is a local user, fail and advise how to fix the problem.
		if connectorRef == nil {
			return nil, trace.AlreadyExists("local user with name %q already exists. Either change "+
				"NameID in assertion or remove local user and try again.", existingUser.GetName())
		}

		log.Debugf("Overwriting existing user %q created with %v connector %v.",
			existingUser.GetName(), connectorRef.Type, connectorRef.ID)

		if err := sas.auth.UpdateUser(ctx, user); err != nil {
			return nil, trace.Wrap(err)
		}
	} else {
		if err := sas.auth.CreateUser(ctx, user); err != nil {
			return nil, trace.Wrap(err)
		}
	}

	return user, nil
}

func ParseSAMLInResponseTo(response string) (string, error) {
	raw, _ := base64.StdEncoding.DecodeString(response)

	doc := etree.NewDocument()
	err := doc.ReadFromBytes(raw)
	if err != nil {
		// Attempt to inflate the response in case it happens to be compressed (as with one case at saml.oktadev.com)
		buf, err := io.ReadAll(flate.NewReader(bytes.NewReader(raw)))
		if err != nil {
			return "", trace.Wrap(err)
		}

		doc = etree.NewDocument()
		err = doc.ReadFromBytes(buf)
		if err != nil {
			return "", trace.Wrap(err)
		}
	}

	if doc.Root() == nil {
		return "", trace.BadParameter("unable to parse response")
	}

	// Try to find the InResponseTo attribute in the SAML response. If we can't find this, return
	// a predictable error message so the caller may choose interpret it as an IdP-initiated payload.
	el := doc.Root()
	responseTo := el.SelectAttr("InResponseTo")
	if responseTo == nil {
		return "", trace.NotFound("missing InResponseTo attribute")
	}
	if responseTo.Value == "" {
		return "", trace.BadParameter("InResponseTo can not be empty")
	}
	return responseTo.Value, nil
}

// SAMLAuthResponse is returned when auth server validated callback parameters
// returned from SAML identity provider
type SAMLAuthResponse struct {
	// Username is an authenticated teleport username
	Username string `json:"username"`
	// Identity contains validated SAML identity
	Identity types.ExternalIdentity `json:"identity"`
	// Web session will be generated by auth server if requested in SAMLAuthRequest
	Session types.WebSession `json:"session,omitempty"`
	// Cert will be generated by certificate authority
	Cert []byte `json:"cert,omitempty"`
	// TLSCert is a PEM encoded TLS certificate
	TLSCert []byte `json:"tls_cert,omitempty"`
	// Req is an original SAML auth request
	Req SAMLAuthRequest `json:"req"`
	// HostSigners is a list of signing host public keys
	// trusted by proxy, used in console login
	HostSigners []types.CertAuthority `json:"host_signers"`
}

// SAMLAuthRequest is a SAML auth request that supports standard json marshaling.
type SAMLAuthRequest struct {
	// ID is a unique request ID.
	ID string `json:"id"`
	// PublicKey is an optional public key, users want these
	// keys to be signed by auth servers user CA in case
	// of successful auth.
	PublicKey []byte `json:"public_key"`
	// CSRFToken is associated with user web session token.
	CSRFToken string `json:"csrf_token"`
	// CreateWebSession indicates if user wants to generate a web
	// session after successful authentication.
	CreateWebSession bool `json:"create_web_session"`
	// ClientRedirectURL is a URL client wants to be redirected
	// after successful authentication.
	ClientRedirectURL string `json:"client_redirect_url"`
}

// ValidateSAMLResponseReq is the request made by the proxy to validate
// and activate a login via SAML.
type ValidateSAMLResponseReq struct {
	Response    string `json:"response"`
	ConnectorID string `json:"connector_id,omitempty"`
}

// SAMLAuthRawResponse is returned when auth server validated callback parameters
// returned from SAML provider
type SAMLAuthRawResponse struct {
	// Username is authenticated teleport username
	Username string `json:"username"`
	// Identity contains validated OIDC identity
	Identity types.ExternalIdentity `json:"identity"`
	// Web session will be generated by auth server if requested in OIDCAuthRequest
	Session json.RawMessage `json:"session,omitempty"`
	// Cert will be generated by certificate authority
	Cert []byte `json:"cert,omitempty"`
	// Req is original oidc auth request
	Req SAMLAuthRequest `json:"req"`
	// HostSigners is a list of signing host public keys
	// trusted by proxy, used in console login
	HostSigners []json.RawMessage `json:"host_signers"`
	// TLSCert is TLS certificate authority certificate
	TLSCert []byte `json:"tls_cert,omitempty"`
}

// SAMLAuthRequestFromProto converts the types.SAMLAuthRequest to SAMLAuthRequestData.
func SAMLAuthRequestFromProto(req *types.SAMLAuthRequest) SAMLAuthRequest {
	return SAMLAuthRequest{
		ID:                req.ID,
		PublicKey:         req.PublicKey,
		CSRFToken:         req.CSRFToken,
		CreateWebSession:  req.CreateWebSession,
		ClientRedirectURL: req.ClientRedirectURL,
	}
}

// ValidateSAMLResponse consumes attribute statements from SAML identity provider
func (sas *SAMLAuthService) ValidateSAMLResponse(ctx context.Context, samlResponse string, connectorID string) (*SAMLAuthResponse, error) {
	event := &apievents.UserLogin{
		Metadata: apievents.Metadata{
			Type: events.UserLoginEvent,
		},
		Method: events.LoginMethodSAML,
	}

	diagCtx := NewSSODiagContext(types.KindSAML, sas.auth)

	auth, err := sas.validateSAMLResponse(ctx, diagCtx, samlResponse, connectorID)
	diagCtx.Info.Error = trace.UserMessage(err)

	diagCtx.WriteToBackend(ctx)

	attributeStatements := diagCtx.Info.SAMLAttributeStatements
	if attributeStatements != nil {
		attributes, err := apievents.EncodeMapStrings(attributeStatements)
		if err != nil {
			event.Status.UserMessage = fmt.Sprintf("Failed to encode identity attributes: %v", err.Error())
			log.WithError(err).Debug("Failed to encode identity attributes.")
		} else {
			event.IdentityAttributes = attributes
		}
	}

	if err != nil {
		event.Code = events.UserSSOLoginFailureCode
		if diagCtx.Info.TestFlow {
			event.Code = events.UserSSOTestFlowLoginFailureCode
		}
		event.Status.Success = false
		event.Status.Error = trace.Unwrap(err).Error()
		event.Status.UserMessage = err.Error()
		if err := sas.emitter.EmitAuditEvent(ctx, event); err != nil {
			log.WithError(err).Warn("Failed to emit SAML login failed event.")
		}
		return nil, trace.Wrap(err)
	}

	event.Status.Success = true
	event.User = auth.Username
	event.Code = events.UserSSOLoginCode
	if diagCtx.Info.TestFlow {
		event.Code = events.UserSSOTestFlowLoginCode
	}

	if err := sas.emitter.EmitAuditEvent(ctx, event); err != nil {
		log.WithError(err).Warn("Failed to emit SAML login event.")
	}

	return auth, nil
}

func (sas *SAMLAuthService) checkIDPInitiatedSAML(ctx context.Context, connector types.SAMLConnector, assertion *saml2.AssertionInfo) error {
	if !connector.GetAllowIDPInitiated() {
		return trace.AccessDenied("IdP initiated SAML is not allowed by the connector configuration")
	}

	// Not all IdP's provide these variables, replay mitigation is best effort.
	if assertion.SessionIndex != "" || assertion.SessionNotOnOrAfter == nil {
		return nil
	}

	err := sas.assertionReplayService.RecognizeSSOAssertion(ctx, connector.GetName(), assertion.SessionIndex, assertion.NameID, *assertion.SessionNotOnOrAfter)
	return trace.Wrap(err)
}

func (sas *SAMLAuthService) validateSAMLResponse(ctx context.Context, diagCtx *SSODiagContext, samlResponse string, connectorID string) (*SAMLAuthResponse, error) {
	idpInitiated := false
	var connector types.SAMLConnector
	var provider *saml2.SAMLServiceProvider
	var request *types.SAMLAuthRequest
	requestID, err := ParseSAMLInResponseTo(samlResponse)

	switch {
	case trace.IsNotFound(err):
		if connectorID == "" {
			return nil, trace.BadParameter("ACS URI did not include a valid SAML connector ID parameter")
		}

		idpInitiated = true
		connector, provider, err = sas.getSAMLConnectorAndProviderByID(ctx, connectorID)
		if err != nil {
			return nil, trace.Wrap(err, "Failed to get SAML connector and provider")
		}
	case err != nil:
		return nil, trace.Wrap(err)
	default:
		diagCtx.RequestID = requestID
		request, err = sas.auth.Identity.GetSAMLAuthRequest(ctx, requestID)
		if err != nil {
			return nil, trace.Wrap(err, "Failed to get SAML Auth Request")
		}

		diagCtx.Info.TestFlow = request.SSOTestFlow
		connector, provider, err = sas.getSAMLConnectorAndProvider(ctx, *request)
		if err != nil {
			return nil, trace.Wrap(err, "Failed to get SAML connector and provider")
		}
	}

	assertionInfo, err := provider.RetrieveAssertionInfo(samlResponse)
	if err != nil {
		return nil, trace.AccessDenied("received response with incorrect or missing attribute statements, please check the identity provider configuration to make sure that mappings for claims/attribute statements are set up correctly. <See: https://goteleport.com/teleport/docs/enterprise/sso/ssh-sso/>, failed to retrieve SAML assertion info from response: %v.", err)
	}

	if assertionInfo != nil {
		diagCtx.Info.SAMLAssertionInfo = (*types.AssertionInfo)(assertionInfo)
	}

	if idpInitiated {
		if err := sas.checkIDPInitiatedSAML(ctx, connector, assertionInfo); err != nil {
			if trace.IsAccessDenied(err) {
				log.Warnf("Failed to process IdP-initiated login request. IdP-initiated login is disabled for this connector: %v.", err)
			}

			return nil, trace.Wrap(err)
		}
	}

	if assertionInfo.WarningInfo.InvalidTime {
		return nil, trace.AccessDenied("invalid time in SAML assertion info. SAML assertion info contained warning: invalid time.")
	}

	if assertionInfo.WarningInfo.NotInAudience {
		return nil, trace.AccessDenied("no audience in SAML assertion info. SAML: not in expected audience. Check auth connector audience field and IdP configuration for typos and other errors.")
	}

	log.Debugf("Obtained SAML assertions for %q.", assertionInfo.NameID)
	log.Debugf("SAML assertion warnings: %+v.", assertionInfo.WarningInfo)

	attributeStatements := map[string][]string{}

	for key, val := range assertionInfo.Values {
		var vals []string
		for _, vv := range val.Values {
			vals = append(vals, vv.Value)
		}
		log.Debugf("SAML assertion: %q: %q.", key, vals)
		attributeStatements[key] = vals
	}

	diagCtx.Info.SAMLAttributeStatements = attributeStatements
	diagCtx.Info.SAMLAttributesToRoles = connector.GetAttributesToRoles()

	if len(connector.GetAttributesToRoles()) == 0 {
		return nil, trace.BadParameter("no attributes to roles mapping, check connector documentation. Attributes-to-roles mapping is empty, SSO user will never have any roles.")
	}

	log.Debugf("Applying %v SAML attribute to roles mappings.", len(connector.GetAttributesToRoles()))

	// Calculate (figure out name, roles, traits, session TTL) of user and
	// create the user in the backend.
	params, err := sas.calculateSAMLUser(diagCtx, connector, *assertionInfo, request)
	if err != nil {
		return nil, trace.Wrap(err, "Failed to calculate user attributes.")
	}

	diagCtx.Info.CreateUserParams = &types.CreateUserParams{
		ConnectorName: params.ConnectorName,
		Username:      params.Username,
		KubeGroups:    params.KubeGroups,
		KubeUsers:     params.KubeUsers,
		Roles:         params.Roles,
		Traits:        params.Traits,
		SessionTTL:    types.Duration(params.SessionTTL),
	}

	user, err := sas.createSAMLUser(params, request != nil && request.SSOTestFlow)
	if err != nil {
		return nil, trace.Wrap(err, "Failed to create user from provided parameters.")
	}

	// Auth was successful, return session, certificate, etc. to caller.
	auth := &SAMLAuthResponse{
		Identity: types.ExternalIdentity{
			ConnectorID: params.ConnectorName,
			Username:    params.Username,
		},
		Username: user.GetName(),
	}

	if request != nil {
		auth.Req = SAMLAuthRequestFromProto(request)
	} else {
		auth.Req = SAMLAuthRequest{
			CreateWebSession: true,
		}
	}

	// In test flow skip signing and creating web sessions.
	if request != nil && request.SSOTestFlow {
		diagCtx.Info.Success = true
		return auth, nil
	}

	// If the request is coming from a browser, create a web session.
	if request == nil || request.CreateWebSession {
		session, err := sas.auth.CreateWebSessionFromReq(ctx, NewWebSessionRequest{
			User:       user.GetName(),
			Roles:      user.GetRoles(),
			Traits:     user.GetTraits(),
			SessionTTL: params.SessionTTL,
			LoginTime:  sas.auth.GetClock().Now().UTC(),
		})
		if err != nil {
			return nil, trace.Wrap(err, "Failed to create web session.")
		}

		auth.Session = session
	}

	// If a public key was provided, sign it and return a certificate.
	if request != nil && len(request.PublicKey) != 0 {

		sshCert, tlsCert, err := sas.auth.CreateSessionCert(user, params.SessionTTL, request.PublicKey, request.Compatibility, request.RouteToCluster,
			request.KubernetesCluster, request.ClientLoginIP, keys.AttestationStatementFromProto(request.AttestationStatement))

		if err != nil {
			return nil, trace.Wrap(err, "Failed to create session certificate.")
		}
		clusterName, err := sas.auth.GetClusterName()
		if err != nil {
			return nil, trace.Wrap(err, "Failed to obtain cluster name.")
		}
		auth.Cert = sshCert
		auth.TLSCert = tlsCert

		// Return the host CA for this cluster only.
		authority, err := sas.auth.GetCertAuthority(ctx, types.CertAuthID{
			Type:       types.HostCA,
			DomainName: clusterName.GetClusterName(),
		}, false)
		if err != nil {
			return nil, trace.Wrap(err, "Failed to obtain cluster's host CA.")
		}
		auth.HostSigners = append(auth.HostSigners, authority)
	}

	diagCtx.Info.Success = true
	return auth, nil
}
