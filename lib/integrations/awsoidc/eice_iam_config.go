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

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"

	awslib "github.com/gravitational/teleport/lib/cloud/aws"
)

const (
	// defaultPolicyNameForEICE is the default name for the Inline Policy added to the IntegrationRole.
	defaultPolicyNameForEICE = "EC2InstanceConnectEndpoint"
)

// EICEIAMConfigureRequest is a request to configure the required Policies to use the EC2 Instance Connect Endpoint feature.
type EICEIAMConfigureRequest struct {
	// Region is the AWS Region.
	// Used to set up the AWS SDK Client.
	Region string

	// IntegrationRole is the Integration's AWS Role used to set up Teleport as an OIDC IdP.
	IntegrationRole string

	// IntegrationRoleEICEPolicy is the Policy Name that is created to allow access to call AWS APIs.
	// Defaults to EC2InstanceConnectEndpoint
	IntegrationRoleEICEPolicy string
}

// CheckAndSetDefaults ensures the required fields are present.
func (r *EICEIAMConfigureRequest) CheckAndSetDefaults() error {
	if r.Region == "" {
		return trace.BadParameter("region is required")
	}

	if r.IntegrationRole == "" {
		return trace.BadParameter("integration role is required")
	}

	if r.IntegrationRoleEICEPolicy == "" {
		r.IntegrationRoleEICEPolicy = defaultPolicyNameForEICE
	}

	return nil
}

// EICEIAMConfigureClient describes the required methods to create the IAM Policies required for accessing EC2 instances usine EICE.
type EICEIAMConfigureClient interface {
	// PutRolePolicy creates or replaces a Policy by its name in a IAM Role.
	PutRolePolicy(ctx context.Context, params *iam.PutRolePolicyInput, optFns ...func(*iam.Options)) (*iam.PutRolePolicyOutput, error)
}

type defaultEICEIAMConfigureClient struct {
	*iam.Client
}

// NewEICEIAMConfigureClient creates a new EICEIAMConfigureClient.
func NewEICEIAMConfigureClient(ctx context.Context, region string) (EICEIAMConfigureClient, error) {
	if region == "" {
		return nil, trace.BadParameter("region is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &defaultEICEIAMConfigureClient{
		Client: iam.NewFromConfig(cfg),
	}, nil
}

// ConfigureEICEIAM set ups the roles required for accessing an EC2 Instance using EICE.
// It creates an embedded policy with the following permissions:
//
// Action: List EC2 instances to add them as Teleport Nodes
//   - ec2:DescribeInstances
//
// Action: List EC2 Instance Connect Endpoints so that knows if they must create one Endpoint.
//   - ec2:DescribeInstanceConnectEndpoints
//
// Action: Select one or more SecurityGroups to apply to the EC2 Instance Connect Endpoints (the VPC's default SG is applied if no SG is provided).
//   - ec2:DescribeSecurityGroups
//
// Action: Create EC2 Instance Connect Endpoint so the user can open a tunnel to the EC2 instance.
// More info: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/permissions-for-ec2-instance-connect-endpoint.html
//   - ec2:CreateInstanceConnectEndpoint
//   - ec2:CreateTags
//   - ec2:CreateNetworkInterface
//   - iam:CreateServiceLinkedRole
//
// Action: Send a temporary SSH Key to the target host.
//   - ec2-instance-connect:SendSSHPublicKey
//
// Action: Open a Tunnel to the EC2 using the Endpoint
//   - ec2-instance-connect:OpenTunnel
//
// The following actions must be allowed by the IAM Role assigned in the Client.
//   - iam:PutRolePolicy
func ConfigureEICEIAM(ctx context.Context, clt EICEIAMConfigureClient, req EICEIAMConfigureRequest) error {
	if err := req.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	ec2ICEPolicyDocument, err := awslib.NewPolicyDocument(
		awslib.StatementForEC2InstanceConnectEndpoint(),
	).Marshal()
	if err != nil {
		return trace.Wrap(err)
	}

	_, err = clt.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		PolicyName:     &req.IntegrationRoleEICEPolicy,
		RoleName:       &req.IntegrationRole,
		PolicyDocument: &ec2ICEPolicyDocument,
	})
	if err != nil {
		if trace.IsNotFound(awslib.ConvertIAMv2Error(err)) {
			return trace.NotFound("role %q not found.", req.IntegrationRole)
		}
		return trace.Wrap(err)
	}

	logrus.Infof("IntegrationRole: IAM Policy %q added to Role %q\n", req.IntegrationRoleEICEPolicy, req.IntegrationRole)
	return nil
}
