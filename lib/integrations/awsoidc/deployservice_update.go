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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/lib/automaticupgrades"
	awslib "github.com/gravitational/teleport/lib/cloud/aws"
)

// waitDuration specifies the amount of time to wait for a service to become healthy after an update.
const waitDuration = time.Minute * 5

// UpdateServiceRequest contains the required fields to update a Teleport Service.
type UpdateServiceRequest struct {
	// TeleportClusterName specifies the teleport cluster name
	TeleportClusterName string
	// TeleportVersionTag specifies the desired teleport version in the format "13.4.0"
	TeleportVersionTag string
	// OwnershipTags specifies ownership tags
	OwnershipTags AWSTags
}

// CheckAndSetDefaults checks and sets default config values.
func (req *UpdateServiceRequest) CheckAndSetDefaults() error {
	if req.TeleportClusterName == "" {
		return trace.BadParameter("teleport cluster name required")
	}

	if req.TeleportVersionTag == "" {
		return trace.BadParameter("teleport version tag required")
	}

	if req.OwnershipTags == nil {
		return trace.BadParameter("ownership tags required")
	}

	return nil
}

// UpdateDeployService updates all the AWS OIDC deployed services with the specified version tag.
func UpdateDeployService(ctx context.Context, clt DeployServiceClient, log *logrus.Entry, req UpdateServiceRequest) error {
	if err := req.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	teleportImage := getDistrolessTeleportImage(req.TeleportVersionTag)
	services, err := getManagedServices(ctx, clt, log, req.TeleportClusterName, req.OwnershipTags)
	if err != nil {
		return trace.Wrap(err)
	}

	for _, ecsService := range services {
		log := log.WithFields(logrus.Fields{
			"ecs-service-arn": aws.ToString(ecsService.ServiceArn),
			"teleport-image":  teleportImage,
		})
		if err := updateServiceContainerImage(ctx, clt, log, &ecsService, teleportImage, req.OwnershipTags); err != nil {
			log.WithError(err).Warn("Failed to upgrade ECS Service.")
			continue
		}
	}

	return nil
}

func updateServiceContainerImage(ctx context.Context, clt DeployServiceClient, log *logrus.Entry, service *ecsTypes.Service, teleportImage string, ownershipTags AWSTags) error {
	taskDefinition, err := getManagedTaskDefinition(ctx, clt, aws.ToString(service.TaskDefinition), ownershipTags)
	if err != nil {
		return trace.Wrap(err)
	}

	currentTeleportImage, err := getTaskDefinitionTeleportImage(taskDefinition)
	if err != nil {
		return trace.Wrap(err)
	}

	// There is no need to update the ecs service if the ecs service is already
	// running the latest stable version of teleport.
	if currentTeleportImage == teleportImage {
		return nil
	}

	registerTaskDefinitionIn, err := generateTaskDefinitionWithImage(taskDefinition, teleportImage, ownershipTags.ToECSTags())
	if err != nil {
		return trace.Wrap(err)
	}

	// Ensure that the upgrader variables are set.
	// These will ensure that the instance reports Teleport upgrader metrics.
	if err := ensureUpgraderEnvironmentVariables(registerTaskDefinitionIn); err != nil {
		return trace.Wrap(err)
	}

	registerTaskDefinitionOut, err := clt.RegisterTaskDefinition(ctx, registerTaskDefinitionIn)
	if err != nil {
		return trace.Wrap(err)
	}
	newTaskDefinitionARN := registerTaskDefinitionOut.TaskDefinition.TaskDefinitionArn
	oldTaskDefinitionARN := aws.ToString(service.TaskDefinition)

	// Update service with new task definition
	serviceNewVersion, err := clt.UpdateService(ctx, generateServiceWithTaskDefinition(service, aws.ToString(newTaskDefinitionARN)))
	if err != nil {
		return trace.Wrap(err)
	}

	// Wait for Service to become stable, or rollback to the previous TaskDefinition.
	go waitServiceStableOrRollback(ctx, clt, log, serviceNewVersion.Service, oldTaskDefinitionARN)

	log.Info("Successfully upgraded ECS Service.")

	return nil
}

func getAllServiceNamesForCluster(ctx context.Context, clt DeployServiceClient, clusterName *string) ([]string, error) {
	ret := make([]string, 0)

	nextToken := ""
	for {
		resp, err := clt.ListServices(ctx, &ecs.ListServicesInput{
			Cluster:   clusterName,
			NextToken: aws.String(nextToken),
		})
		if err != nil {
			return nil, awslib.ConvertIAMv2Error(err)
		}

		ret = append(ret, resp.ServiceArns...)

		nextToken = aws.ToString(resp.NextToken)
		if nextToken == "" {
			break
		}
	}
	return ret, nil
}

func getManagedServices(ctx context.Context, clt DeployServiceClient, log *logrus.Entry, teleportClusterName string, ownershipTags AWSTags) ([]ecsTypes.Service, error) {
	// The Cluster name is created using the Teleport Cluster Name.
	// Check the DeployDatabaseServiceRequest.CheckAndSetDefaults
	// and DeployServiceRequest.CheckAndSetDefaults.
	wellKnownClusterName := aws.String(normalizeECSClusterName(teleportClusterName))

	ecsServiceNames, err := getAllServiceNamesForCluster(ctx, clt, wellKnownClusterName)
	if err != nil {
		if !trace.IsAccessDenied(err) {
			return nil, trace.Wrap(err)
		}

		// Previous versions of the DeployService only deployed a single ECS Service, based on the DatabaseServiceDeploymentMode.
		// During the Discover Wizard flow, users were asked to run a script that added the required permissions, but ecs:ListServices was not initially included.
		// For those situations, fallback to using the only ECS Service that was deployed.
		ecsServiceNameLegacy := normalizeECSServiceName(teleportClusterName, DatabaseServiceDeploymentMode)
		ecsServiceNames = []string{ecsServiceNameLegacy}
	}

	ecsServices := make([]ecsTypes.Service, 0, len(ecsServiceNames))

	// According to AWS API docs, a maximum of 10 Services can be queried at the same time when using the ecs:DescribeServices operation.
	batchSize := 10
	for batchStart := 0; batchStart < len(ecsServiceNames); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > len(ecsServiceNames) {
			batchEnd = len(ecsServiceNames)
		}

		describeServicesOut, err := clt.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  wellKnownClusterName,
			Services: ecsServiceNames[batchStart:batchEnd],
			Include:  []ecsTypes.ServiceField{ecsTypes.ServiceFieldTags},
		})
		if err != nil {
			return nil, trace.Wrap(err)
		}

		// Filter out Services without Ownership tags or an invalid LaunchType.
		for _, s := range describeServicesOut.Services {
			log := log.WithField("ecs-service", aws.ToString(s.ServiceArn))
			if !ownershipTags.MatchesECSTags(s.Tags) {
				log.Warnf("ECS Service exists but is not managed by Teleport. "+
					"Add the following tags to allow Teleport to manage this service: %s", ownershipTags)
				continue
			}
			// If the LaunchType is the required one, than we can update the current Service.
			// Otherwise we have to delete it.
			if s.LaunchType != ecsTypes.LaunchTypeFargate {
				log.Warnf("ECS Service already exists but has an invalid LaunchType %q. Delete the Service and try again.", s.LaunchType)
				continue
			}
			ecsServices = append(ecsServices, s)
		}
	}

	return ecsServices, nil
}

func getManagedTaskDefinition(ctx context.Context, clt DeployServiceClient, taskDefinitionName string, ownershipTags AWSTags) (*ecsTypes.TaskDefinition, error) {
	describeTaskDefinitionOut, err := clt.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
		Include:        []ecsTypes.TaskDefinitionField{ecsTypes.TaskDefinitionFieldTags},
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	if !ownershipTags.MatchesECSTags(describeTaskDefinitionOut.Tags) {
		return nil, trace.Errorf("ECS Task Definition %q already exists but is not managed by Teleport. "+
			"Add the following tags to allow Teleport to manage this task definition: %s", taskDefinitionName, ownershipTags)
	}
	return describeTaskDefinitionOut.TaskDefinition, nil
}

func getTaskDefinitionTeleportImage(taskDefinition *ecsTypes.TaskDefinition) (string, error) {
	if len(taskDefinition.ContainerDefinitions) != 1 {
		return "", trace.BadParameter("expected 1 task container definition, but got %d", len(taskDefinition.ContainerDefinitions))
	}
	return aws.ToString(taskDefinition.ContainerDefinitions[0].Image), nil
}

// waitServiceStableOrRollback waits for the ECS Service to be stable, and if it takes longer than 5 minutes, it restarts it with its old task definition.
func waitServiceStableOrRollback(ctx context.Context, clt DeployServiceClient, log *logrus.Entry, service *ecsTypes.Service, oldTaskDefinitionARN string) {
	log = log.WithFields(logrus.Fields{
		"ecs-service":         aws.ToString(service.ServiceArn),
		"task-definition":     aws.ToString(service.TaskDefinition),
		"old-task-definition": oldTaskDefinitionARN,
	})

	log.Debug("Waiting for ECS Service to become stable.")
	serviceStableWaiter := ecs.NewServicesStableWaiter(clt)
	waitErr := serviceStableWaiter.Wait(ctx, &ecs.DescribeServicesInput{
		Services: []string{aws.ToString(service.ServiceName)},
		Cluster:  service.ClusterArn,
	}, waitDuration)
	if waitErr == nil {
		log.Debug("ECS Service is stable.")
		return
	}

	log.WithError(waitErr).Warn("ECS Service is not stable, restarting the service with its previous TaskDefinition.")
	_, rollbackErr := clt.UpdateService(ctx, generateServiceWithTaskDefinition(service, oldTaskDefinitionARN))
	if rollbackErr != nil {
		log.WithError(rollbackErr).Warn("Failed to update ECS Service with its previous version.")
	}
}

// generateTaskDefinitionWithImage returns new register task definition input with the desired teleport image
func generateTaskDefinitionWithImage(taskDefinition *ecsTypes.TaskDefinition, teleportImage string, tags []ecsTypes.Tag) (*ecs.RegisterTaskDefinitionInput, error) {
	if len(taskDefinition.ContainerDefinitions) != 1 {
		return nil, trace.BadParameter("expected 1 task container definition, but got %d", len(taskDefinition.ContainerDefinitions))
	}

	// Copy container definition and replace the teleport image with desired version
	newContainerDefinition := new(ecsTypes.ContainerDefinition)
	awsutil.Copy(newContainerDefinition, &taskDefinition.ContainerDefinitions[0])
	newContainerDefinition.Image = aws.String(teleportImage)

	// Copy task definition and replace container definitions
	registerTaskDefinitionIn := new(ecs.RegisterTaskDefinitionInput)
	awsutil.Copy(registerTaskDefinitionIn, taskDefinition)
	registerTaskDefinitionIn.ContainerDefinitions = []ecsTypes.ContainerDefinition{*newContainerDefinition}
	registerTaskDefinitionIn.Tags = tags

	return registerTaskDefinitionIn, nil
}

// ensureUpgraderEnvironmentVariables modifies the taskDefinition and ensures that
// the upgrader specific environment variables are set.
func ensureUpgraderEnvironmentVariables(taskDefinition *ecs.RegisterTaskDefinitionInput) error {
	containerDefinitions := []ecsTypes.ContainerDefinition{}
	for _, containerDefinition := range taskDefinition.ContainerDefinitions {
		environment := []ecsTypes.KeyValuePair{}

		// Copy non-upgrader specific environemt variables as is
		for _, env := range containerDefinition.Environment {
			if aws.ToString(env.Name) == automaticupgrades.EnvUpgrader ||
				aws.ToString(env.Name) == automaticupgrades.EnvUpgraderVersion {
				continue
			}
			environment = append(environment, env)
		}

		// Ensure ugprader specific environment variables are set
		environment = append(environment,
			ecsTypes.KeyValuePair{
				Name:  aws.String(automaticupgrades.EnvUpgraderVersion),
				Value: aws.String(teleport.Version),
			},
		)
		containerDefinition.Environment = environment
		containerDefinitions = append(containerDefinitions, containerDefinition)
	}
	taskDefinition.ContainerDefinitions = containerDefinitions
	return nil
}

// generateServiceWithTaskDefinition returns new update service input with the desired task definition
func generateServiceWithTaskDefinition(service *ecsTypes.Service, taskDefinitionName string) *ecs.UpdateServiceInput {
	updateServiceIn := new(ecs.UpdateServiceInput)
	awsutil.Copy(updateServiceIn, service)
	updateServiceIn.Service = service.ServiceName
	updateServiceIn.Cluster = service.ClusterArn
	updateServiceIn.TaskDefinition = aws.String(taskDefinitionName)
	return updateServiceIn
}
