/*

 Copyright 2022 Gravitational, Inc.

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
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/auth/testauthority"
	"github.com/gravitational/teleport/lib/tlsca"
)

func Test_getSnowflakeJWTParams(t *testing.T) {
	type args struct {
		accountName string
		userName    string
		publicKey   []byte
	}
	tests := []struct {
		name        string
		args        args
		wantSubject string
		wantIssuer  string
	}{
		{
			name: "only account locator",
			args: args{
				accountName: "abc123",
				userName:    "user1",
				publicKey:   []byte("fakeKey"),
			},
			wantSubject: "ABC123.USER1",
			wantIssuer:  "ABC123.USER1.SHA256:q3OCFrBX3MOuBefrAI0e2UgNh5yLGIiSSIuncvcMdGA=",
		},
		{
			name: "GCP",
			args: args{
				accountName: "abc321.us-central1.gcp",
				userName:    "user1",
				publicKey:   []byte("fakeKey"),
			},
			wantSubject: "ABC321.USER1",
			wantIssuer:  "ABC321.USER1.SHA256:q3OCFrBX3MOuBefrAI0e2UgNh5yLGIiSSIuncvcMdGA=",
		},
		{
			name: "AWS",
			args: args{
				accountName: "abc321.us-west-2.aws",
				userName:    "user2",
				publicKey:   []byte("fakeKey"),
			},
			wantSubject: "ABC321.USER2",
			wantIssuer:  "ABC321.USER2.SHA256:q3OCFrBX3MOuBefrAI0e2UgNh5yLGIiSSIuncvcMdGA=",
		},
		{
			name: "global",
			args: args{
				accountName: "testaccount-user.global",
				userName:    "user2",
				publicKey:   []byte("fakeKey"),
			},
			wantSubject: "TESTACCOUNT.USER2",
			wantIssuer:  "TESTACCOUNT.USER2.SHA256:q3OCFrBX3MOuBefrAI0e2UgNh5yLGIiSSIuncvcMdGA=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject, issuer := getSnowflakeJWTParams(tt.args.accountName, tt.args.userName, tt.args.publicKey)

			require.Equal(t, tt.wantSubject, subject)
			require.Equal(t, tt.wantIssuer, issuer)
		})
	}
}

func TestDBCertSigning(t *testing.T) {
	t.Parallel()
	authServer, err := NewTestAuthServer(TestAuthServerConfig{
		Clock:       clockwork.NewFakeClockAt(time.Now()),
		ClusterName: "local.me",
		Dir:         t.TempDir(),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, authServer.Close()) })

	ctx := context.Background()

	privateKey, err := testauthority.New().GeneratePrivateKey()
	require.NoError(t, err)

	csr, err := tlsca.GenerateCertificateRequestPEM(pkix.Name{
		CommonName: "localhost",
	}, privateKey)
	require.NoError(t, err)

	// Set rotation to init phase. New CA will be generated.
	// DB service should use active key to sign certificates.
	// tctl should use new key to sign certificates.
	err = authServer.AuthServer.RotateCertAuthority(ctx, RotateRequest{
		Type:        types.DatabaseCA,
		TargetPhase: types.RotationPhaseInit,
		Mode:        types.RotationModeManual,
	})
	require.NoError(t, err)
	err = authServer.AuthServer.RotateCertAuthority(ctx, RotateRequest{
		Type:        types.DatabaseClientCA,
		TargetPhase: types.RotationPhaseInit,
		Mode:        types.RotationModeManual,
	})
	require.NoError(t, err)

	dbCAs, err := authServer.AuthServer.GetCertAuthorities(ctx, types.DatabaseCA, false)
	require.NoError(t, err)
	require.Len(t, dbCAs, 1)
	require.Len(t, dbCAs[0].GetActiveKeys().TLS, 1)
	require.Len(t, dbCAs[0].GetAdditionalTrustedKeys().TLS, 1)
	activeDBCACert := dbCAs[0].GetActiveKeys().TLS[0].Cert
	newDBCACert := dbCAs[0].GetAdditionalTrustedKeys().TLS[0].Cert

	dbClientCAs, err := authServer.AuthServer.GetCertAuthorities(ctx, types.DatabaseClientCA, false)
	require.NoError(t, err)
	require.Len(t, dbClientCAs, 1)
	require.Len(t, dbClientCAs[0].GetActiveKeys().TLS, 1)
	require.Len(t, dbClientCAs[0].GetAdditionalTrustedKeys().TLS, 1)
	activeDBClientCACert := dbClientCAs[0].GetActiveKeys().TLS[0].Cert
	newDBClientCACert := dbClientCAs[0].GetAdditionalTrustedKeys().TLS[0].Cert

	tests := []struct {
		name           string
		requester      proto.DatabaseCertRequest_Requester
		wantCertSigner []byte
		wantCACerts    [][]byte
		wantKeyUsage   []x509.ExtKeyUsage
	}{
		{
			name:           "DB service request is signed by active db client CA and trusts db CAs",
			wantCertSigner: activeDBClientCACert,
			wantCACerts:    [][]byte{activeDBCACert, newDBCACert},
			wantKeyUsage:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		},
		{
			name:           "tctl request is signed by new db CA and trusts db client CAs",
			requester:      proto.DatabaseCertRequest_TCTL,
			wantCertSigner: newDBCACert,
			wantCACerts:    [][]byte{activeDBClientCACert, newDBClientCACert},
			wantKeyUsage:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			certResp, err := authServer.AuthServer.GenerateDatabaseCert(ctx, &proto.DatabaseCertRequest{
				CSR:           csr,
				ServerName:    "localhost",
				TTL:           proto.Duration(time.Hour),
				RequesterName: tt.requester,
			})
			require.NoError(t, err)
			require.Equal(t, tt.wantCACerts, certResp.CACerts)

			// verify that the response cert is a DB CA cert.
			mustVerifyCert(t, tt.wantCertSigner, certResp.Cert, tt.wantKeyUsage...)
		})
	}
}

// mustVerifyCert is a helper func that verifies leaf cert with root cert.
func mustVerifyCert(t *testing.T, rootPEM, leafPEM []byte, keyUsages ...x509.ExtKeyUsage) {
	t.Helper()
	leafCert, err := tlsca.ParseCertificatePEM(leafPEM)
	require.NoError(t, err)

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(rootPEM)
	require.True(t, ok)
	opts := x509.VerifyOptions{
		Roots:     certPool,
		KeyUsages: keyUsages,
	}
	// Verify if the generated certificate can be verified with the correct CA.
	_, err = leafCert.Verify(opts)
	require.NoError(t, err)
}
