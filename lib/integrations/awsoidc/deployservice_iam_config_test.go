/*
Copyright 2023 Gravitational, Inc.

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

package awsoidc

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gravitational/trace"
	"github.com/stretchr/testify/require"
)

var badParameterCheck = func(t require.TestingT, err error, msgAndArgs ...interface{}) {
	require.True(t, trace.IsBadParameter(err), `expected "bad parameter", but got %v`, err)
}

var alreadyExistsCheck = func(t require.TestingT, err error, msgAndArgs ...interface{}) {
	require.True(t, trace.IsAlreadyExists(err), `expected "already exists", but got %v`, err)
}

var notFoundCheck = func(t require.TestingT, err error, msgAndArgs ...interface{}) {
	require.True(t, trace.IsNotFound(err), `expected "not found", but got %v`, err)
}

var baseReq = func() DeployServiceIAMConfigureRequest {
	return DeployServiceIAMConfigureRequest{
		Cluster:         "mycluster",
		IntegrationName: "myintegration",
		Region:          "us-east-1",
		IntegrationRole: "integrationrole",
		TaskRole:        "taskrole",
	}
}

func TestDeployServiceIAMConfigReqDefaults(t *testing.T) {
	for _, tt := range []struct {
		name     string
		req      func() DeployServiceIAMConfigureRequest
		errCheck require.ErrorAssertionFunc
		expected DeployServiceIAMConfigureRequest
	}{
		{
			name:     "set defaults",
			req:      baseReq,
			errCheck: require.NoError,
			expected: DeployServiceIAMConfigureRequest{
				Cluster:                            "mycluster",
				IntegrationName:                    "myintegration",
				Region:                             "us-east-1",
				IntegrationRole:                    "integrationrole",
				TaskRole:                           "taskrole",
				partitionID:                        "aws",
				IntegrationRoleDeployServicePolicy: "DeployService",
				ResourceCreationTags: AWSTags{
					"teleport.dev/cluster":     "mycluster",
					"teleport.dev/integration": "myintegration",
					"teleport.dev/origin":      "integration_awsoidc",
				},
			},
		},
		{
			name: "missing cluster",
			req: func() DeployServiceIAMConfigureRequest {
				req := baseReq()
				req.Cluster = ""
				return req
			},
			errCheck: badParameterCheck,
		},
		{
			name: "missing integration name",
			req: func() DeployServiceIAMConfigureRequest {
				req := baseReq()
				req.IntegrationName = ""
				return req
			},
			errCheck: badParameterCheck,
		},
		{
			name: "missing region",
			req: func() DeployServiceIAMConfigureRequest {
				req := baseReq()
				req.Region = ""
				return req
			},
			errCheck: badParameterCheck,
		},
		{
			name: "missing integration role",
			req: func() DeployServiceIAMConfigureRequest {
				req := baseReq()
				req.IntegrationRole = ""
				return req
			},
			errCheck: badParameterCheck,
		},
		{
			name: "missing task role",
			req: func() DeployServiceIAMConfigureRequest {
				req := baseReq()
				req.TaskRole = ""
				return req
			},
			errCheck: badParameterCheck,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.req()
			err := req.CheckAndSetDefaults()
			tt.errCheck(t, err)
			if err != nil {
				return
			}

			require.Equal(t, tt.expected, req)
		})
	}
}

func TestDeployServiceIAMConfig(t *testing.T) {
	ctx := context.Background()

	for _, tt := range []struct {
		name              string
		mockAccountID     string
		mockExistingRoles []string
		req               func() DeployServiceIAMConfigureRequest
		errCheck          require.ErrorAssertionFunc
	}{
		{
			name:              "valid",
			mockAccountID:     "123456789012",
			mockExistingRoles: []string{"integrationrole"},
			req:               baseReq,
			errCheck:          require.NoError,
		},
		{
			name:              "task role already exists",
			mockAccountID:     "123456789012",
			mockExistingRoles: []string{"integrationrole", "taskrole"},
			req:               baseReq,
			errCheck:          alreadyExistsCheck,
		},
		{
			name:              "integration role does not exist",
			mockAccountID:     "123456789012",
			mockExistingRoles: []string{},
			req:               baseReq,
			errCheck:          notFoundCheck,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			clt := mockDeployServiceIAMConfigClient{
				accountID:     tt.mockAccountID,
				existingRoles: tt.mockExistingRoles,
			}

			err := ConfigureDeployServiceIAM(ctx, &clt, tt.req())
			tt.errCheck(t, err)
		})
	}
}

type mockDeployServiceIAMConfigClient struct {
	accountID     string
	existingRoles []string
}

// GetCallerIdentity returns information about the caller identity.
func (m *mockDeployServiceIAMConfigClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return &sts.GetCallerIdentityOutput{
		Account: &m.accountID,
	}, nil
}

// CreateRole creates a new IAM Role.
func (m *mockDeployServiceIAMConfigClient) CreateRole(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	alreadyExistsMessage := fmt.Sprintf("Role %q already exists.", *params.RoleName)
	if slices.Contains(m.existingRoles, *params.RoleName) {
		return nil, &iamTypes.EntityAlreadyExistsException{
			Message: &alreadyExistsMessage,
		}
	}
	m.existingRoles = append(m.existingRoles, *params.RoleName)

	return nil, nil
}

// PutRolePolicy creates or replaces a Policy by its name in a IAM Role.
func (m *mockDeployServiceIAMConfigClient) PutRolePolicy(ctx context.Context, params *iam.PutRolePolicyInput, optFns ...func(*iam.Options)) (*iam.PutRolePolicyOutput, error) {
	if !slices.Contains(m.existingRoles, *params.RoleName) {
		noSuchEntityMessage := fmt.Sprintf("Role %q does not exist.", *params.RoleName)
		return nil, &iamTypes.NoSuchEntityException{
			Message: &noSuchEntityMessage,
		}
	}
	return nil, nil
}
