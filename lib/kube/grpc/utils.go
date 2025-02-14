// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubev1

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net"

	"github.com/gravitational/trace"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/lib/auth/authclient"
	"github.com/gravitational/teleport/lib/auth/native"
	"github.com/gravitational/teleport/lib/client"
	"github.com/gravitational/teleport/lib/tlsca"
	"github.com/gravitational/teleport/lib/utils"
)

// getWebAddrAndKubeSNI returns the address of the web server that is running on this
// proxy and the Kube SNI. They are used to dial the Kube proxy to retrieve the pods
// available to the user.
// Since this grpc server is only enabled if the proxy is listening with
// multiplexing mode, the Kube proxy is always reachable on the same address as
// the web server using the SNI.
func getWebAddrAndKubeSNI(proxyAddr string) (string, string, error) {
	addr, port, err := utils.SplitHostPort(proxyAddr)
	if err != nil {
		return "", "", trace.Wrap(err)
	}
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", "", trace.BadParameter("proxy address %q must be have address:port format", proxyAddr)
	}
	sni := client.GetKubeTLSServerName(addr)
	// if the proxy is a unspecified address (0.0.0.0, ::), use localhost.
	if ip.IsUnspecified() {
		addr = string(teleport.PrincipalLocalhost)
	}
	return sni, net.JoinHostPort(addr, port), nil
}

// requestCertificate requests a short-lived certificate for the user using the
// Kubernetes CA.
func (s *Server) requestCertificate(username string, cluster string, identity tlsca.Identity) (*rest.Config, error) {
	s.cfg.Log.Debugf("Requesting K8s cert for %v.", username)
	keyPEM, _, err := native.GenerateKeyPair()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	privateKey, err := ssh.ParseRawPrivateKey(keyPEM)
	if err != nil {
		return nil, trace.Wrap(err, "failed to parse private key")
	}

	subject, err := identity.Subject()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	csr := &x509.CertificateRequest{
		Subject: subject,
	}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csr, privateKey)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	response, err := s.cfg.Signer.ProcessKubeCSR(authclient.KubeCSR{
		Username:    username,
		ClusterName: cluster,
		CSR:         csrPEM,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &rest.Config{
		Host: s.proxyAddress,
		TLSClientConfig: rest.TLSClientConfig{
			CertData:   response.Cert,
			KeyData:    keyPEM,
			CAData:     bytes.Join(response.CertAuthorities, []byte("\n")),
			ServerName: s.kubeProxySNI,
		},
	}, nil
}

// newKubernetesClient creates a new Kubernetes client with short-lived user
// certificates that include in the roles field the available search_as_role
// roles.
func (s *Server) newKubernetesClient(cluster string, identity tlsca.Identity) (kubernetes.Interface, error) {
	cfg, err := s.requestCertificate(identity.Username, cluster, identity)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	return client, trace.Wrap(err)
}

// decideLimit returns the number of items we should request for. If respectLimit
// is true, it returns the difference between the max number of items and the
// number of items already included in the response.
// If false, returns the max number of items.
func decideLimit(limit, items int64, respectLimit bool) int64 {
	if respectLimit {
		return limit - items
	}
	return limit
}
