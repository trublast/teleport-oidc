/**
 * Copyright 2026 Flant JSC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package web

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/httplib"
	"github.com/gravitational/teleport/lib/reversetunnelclient"
	"github.com/gravitational/teleport/lib/web/ui"
)

// getAccessRequests returns the list of access requests visible to the
// authenticated user on the requested cluster.
func (h *Handler) getAccessRequests(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
	sessionCtx *SessionContext,
	site reversetunnelclient.RemoteSite,
) (interface{}, error) {
	ctx := r.Context()
	clt, err := sessionCtx.GetUserClient(ctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	requests, err := clt.GetAccessRequests(ctx, types.AccessRequestFilter{})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return ui.MakeAccessRequests(requests), nil
}

// createAccessRequestReq is the body of a POST /webapi/sites/:site/accessrequests request.
type createAccessRequestReq struct {
	// Reason is the optional human-readable reason for the request.
	Reason string `json:"reason"`
	// Roles is the list of roles being requested.
	Roles []string `json:"roles"`
	// ResourceIDs is the list of specific resources requested.
	ResourceIDs []ui.AccessRequestResourceID `json:"resourceIds"`
	// SuggestedReviewers is an optional list of suggested reviewers.
	SuggestedReviewers []string `json:"suggestedReviewers"`
	// MaxDuration is the optional desired maximum duration of the access request.
	MaxDuration time.Time `json:"maxDuration"`
	// AssumeStartTime is the optional time at which the approved roles can
	// be assumed (RFC3339, must be in the future).
	AssumeStartTime time.Time `json:"assumeStartTime"`
	// DryRun indicates this is a dry-run request that should validate but not
	// persist the access request.
	DryRun bool `json:"dryRun"`
}

// createAccessRequest creates a new access request on behalf of the authenticated user.
func (h *Handler) createAccessRequest(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
	sessionCtx *SessionContext,
	site reversetunnelclient.RemoteSite,
) (interface{}, error) {
	var req *createAccessRequestReq
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}

	ctx := r.Context()
	clt, err := sessionCtx.GetUserClient(ctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	resourceIDs := make([]types.ResourceID, 0, len(req.ResourceIDs))
	for _, rid := range req.ResourceIDs {
		resourceIDs = append(resourceIDs, types.ResourceID{
			Kind:            rid.Kind,
			Name:            rid.Name,
			ClusterName:     rid.ClusterName,
			SubResourceName: rid.SubResourceName,
		})
	}

	accessReq, err := types.NewAccessRequestWithResources(
		uuid.New().String(),
		sessionCtx.GetUser(),
		req.Roles,
		resourceIDs,
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	accessReq.SetRequestReason(req.Reason)
	if len(req.SuggestedReviewers) > 0 {
		accessReq.SetSuggestedReviewers(req.SuggestedReviewers)
	}
	if !req.MaxDuration.IsZero() {
		accessReq.SetMaxDuration(req.MaxDuration)
	}
	if !req.AssumeStartTime.IsZero() {
		accessReq.SetAssumeStartTime(req.AssumeStartTime)
	}
	if req.DryRun {
		accessReq.SetDryRun(true)
	}

	created, err := clt.CreateAccessRequestV2(ctx, accessReq)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return ui.MakeAccessRequest(created), nil
}

// reviewAccessRequestReq is the body of a POST /webapi/sites/:site/accessrequests/:id/review request.
type reviewAccessRequestReq struct {
	// State is the new request state (APPROVED or DENIED).
	State string `json:"state"`
	// Reason is an optional human-readable explanation for the resolution.
	Reason string `json:"reason"`
}

// reviewAccessRequest transitions an access request to either APPROVED or DENIED
// by submitting an access review on behalf of the calling user.
//
// SubmitAccessReview is used (rather than SetAccessRequestState) so that
// reviewer permissions are checked against `allow.review_requests` on the
// caller's roles, instead of requiring full `update` permission on the
// `access_request` resource. This is what enables reviewer RBAC in OSS once
// AdvancedAccessWorkflows is enabled.
func (h *Handler) reviewAccessRequest(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
	sessionCtx *SessionContext,
	site reversetunnelclient.RemoteSite,
) (interface{}, error) {
	requestID := p.ByName("id")
	if requestID == "" {
		return nil, trace.BadParameter("missing access request id")
	}

	var req *reviewAccessRequestReq
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}

	var state types.RequestState
	if err := state.Parse(req.State); err != nil {
		return nil, trace.Wrap(err)
	}
	if !state.IsApproved() && !state.IsDenied() {
		return nil, trace.BadParameter("review state must be APPROVED or DENIED, got %q", req.State)
	}

	ctx := r.Context()
	clt, err := sessionCtx.GetUserClient(ctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	submission := types.AccessReviewSubmission{
		RequestID: requestID,
		Review: types.AccessReview{
			Author:        sessionCtx.GetUser(),
			ProposedState: state,
			Reason:        req.Reason,
			Created:       h.clock.Now().UTC(),
		},
	}

	updated, err := clt.SubmitAccessReview(ctx, submission)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return ui.MakeAccessRequest(updated), nil
}

// deleteAccessRequest removes an access request by id.
func (h *Handler) deleteAccessRequest(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
	sessionCtx *SessionContext,
	site reversetunnelclient.RemoteSite,
) (interface{}, error) {
	requestID := p.ByName("id")
	if requestID == "" {
		return nil, trace.BadParameter("missing access request id")
	}

	ctx := r.Context()
	clt, err := sessionCtx.GetUserClient(ctx, site)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if err := clt.DeleteAccessRequest(ctx, requestID); err != nil {
		return nil, trace.Wrap(err)
	}

	return OK(), nil
}
