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
package e2e

import (
	"testing"

	"github.com/gravitational/teleport/integration/helpers"
	"github.com/gravitational/teleport/lib"
)

const (
	// awsRegionEnv is the environment variable that specifies the AWS region
	// where the EKS cluster is running.
	awsRegionEnv = "AWS_REGION"
	// discoveryMatcherLabelsEnv is the env variable that specifies the matcher
	// labels to use in test discovery services.
	discoveryMatcherLabelsEnv = "DISCOVERY_MATCHER_LABELS"
	// rdsAccessRoleARNEnv is the environment variable that specifies the IAM
	// role ARN that Teleport Database Service will assume to access RDS
	// databases.
	// See modules/databases-ci/ from cloud-terraform repo for more details.
	rdsAccessRoleARNEnv = "RDS_ACCESS_ROLE"
	// rdsDiscoveryRoleARNEnv is the environment variable that specifies the
	// IAM role ARN that Teleport Discovery Service will assume to discover
	// RDS databases.
	// See modules/databases-ci/ from cloud-terraform repo for more details.
	rdsDiscoveryRoleARNEnv = "RDS_DISCOVERY_ROLE"
	// rdsPostgresInstanceNameEnv is the environment variable that specifies the
	// name of the RDS Postgres DB instance that will be created by the Teleport
	// Discovery Service.
	rdsPostgresInstanceNameEnv = "RDS_POSTGRES_INSTANCE_NAME"
	// rdsMySQLInstanceNameEnv is the environment variable that specifies the
	// name of the RDS MySQL DB instance that will be created by the Teleport
	// Discovery Service.
	rdsMySQLInstanceNameEnv = "RDS_MYSQL_INSTANCE_NAME"
	// rdsMariaDBInstanceNameEnv is the environment variable that specifies the
	// name of the RDS MariaDB instance that will be created by the Teleport
	// Discovery Service.
	rdsMariaDBInstanceNameEnv = "RDS_MARIADB_INSTANCE_NAME"
	// rssAccessRoleARNEnv is the environment variable that specifies the IAM
	// role ARN that Teleport Database Service will assume to access Redshift
	// Serverless databases.
	// See modules/databases-ci/ from cloud-terraform repo for more details.
	rssAccessRoleARNEnv = "REDSHIFT_SERVERLESS_ACCESS_ROLE"
	// rssDiscoveryRoleARNEnv is the environment variable that specifies the
	// IAM role ARN that Teleport Discovery Service will assume to discover
	// Redshift Serverless databases.
	// See modules/databases-ci/ from cloud-terraform repo for more details.
	rssDiscoveryRoleARNEnv = "REDSHIFT_SERVERLESS_DISCOVERY_ROLE"
	// rssNameEnv is the environment variable that specifies the
	// name of the Redshift Serverless workgroup that will be created by the
	// Teleport Discovery Service.
	rssNameEnv = "REDSHIFT_SERVERLESS_WORKGROUP_NAME"
	// rssEndpointNameEnv is the environment variable that specifies the
	// name of the Redshift Serverless workgroup's access endpoint that
	// will be created by the Teleport Discovery Service.
	rssEndpointNameEnv = "REDSHIFT_SERVERLESS_ENDPOINT_NAME"
	// rssDBUserEnv is the name of the IAM role that tests will use as a
	// database user to connect to Redshift Serverless.
	rssDBUserEnv = "REDSHIFT_SERVERLESS_IAM_DB_USER"
	// redshiftAccessRoleARNEnv is the environment variable that specifies the
	// IAM role ARN that Teleport Database Service will assume to access Redshift
	// cluster databases.
	// See modules/databases-ci/ from cloud-terraform repo for more details.
	redshiftAccessRoleARNEnv = "REDSHIFT_ACCESS_ROLE"
	// redshiftDiscoveryRoleARNEnv is the environment variable that specifies the
	// IAM role ARN that Teleport Discovery Service will assume to discover
	// Redshift cluster databases.
	// See modules/databases-ci/ from cloud-terraform repo for more details.
	redshiftDiscoveryRoleARNEnv = "REDSHIFT_DISCOVERY_ROLE"
	// redshiftNameEnv is the environment variable that specifies the
	// name of the Redshift cluster db that will be created by the
	// Teleport Discovery Service.
	redshiftNameEnv = "REDSHIFT_CLUSTER_NAME"
	// kubeSvcRoleARNEnv is the environment variable that specifies
	// the IAM role that Teleport Kubernetes Service will assume to access the EKS cluster.
	// This role needs to have the following permissions:
	// - eks:DescribeCluster
	// But it also requires the role to be mapped to a Kubernetes group with the following RBAC permissions:
	//	apiVersion: rbac.authorization.k8s.io/v1
	//	kind: ClusterRole
	//	metadata:
	//		name: teleport-role
	//	rules:
	//	- apiGroups:
	//		- ""
	//		resources:
	//		- users
	//		- groups
	//		- serviceaccounts
	//		verbs:
	//		- impersonate
	//	- apiGroups:
	//		- ""
	//		resources:
	//		- pods
	//		verbs:
	//		- get
	//	- apiGroups:
	//		- "authorization.k8s.io"
	//		resources:
	//		- selfsubjectaccessreviews
	//		- selfsubjectrulesreviews
	//		verbs:
	//		- create
	// check modules/eks-discovery-ci/ from cloud-terraform repo for more details.
	kubeSvcRoleARNEnv = "KUBERNETES_SERVICE_ASSUME_ROLE"
	// kubeDiscoverySvcRoleARNEnv is the environment variable that specifies
	// the IAM role that Teleport Discovery Service will assume to list the EKS clusters.
	// This role needs to have the following permissions:
	// - eks:DescribeCluster
	// - eks:ListClusters
	// check modules/eks-discovery-ci/ from cloud-terraform repo for more details.
	kubeDiscoverySvcRoleARNEnv = "KUBE_DISCOVERY_SERVICE_ASSUME_ROLE"
	// eksClusterNameEnv is the environment variable that specifies the name of
	// the EKS cluster that will be created by Teleport Discovery Service.
	eksClusterNameEnv = "EKS_CLUSTER_NAME"
)

// TestMain will re-execute Teleport to run a command if "exec" is passed to
// it as an argument. Otherwise, it will run tests as normal.
func TestMain(m *testing.M) {
	// agents connect over a reverse tunnel to proxy, so we use insecure mode.
	lib.SetInsecureDevMode(true)
	helpers.TestMainImplementation(m)
}
