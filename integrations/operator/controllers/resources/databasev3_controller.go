/*
Copyright 2026 Flant JSC

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

package resources

import (
	"context"

	"github.com/gravitational/trace"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gravitational/teleport/api/types"
	resourcesv1 "github.com/gravitational/teleport/integrations/operator/apis/resources/v1"
	"github.com/gravitational/teleport/integrations/operator/sidecar"
)

// databaseClient implements TeleportResourceClient and offers CRUD methods needed to reconcile databases
type databaseClient struct {
	TeleportClientAccessor sidecar.ClientAccessor
}

// Get gets the Teleport database of a given name
func (r databaseClient) Get(ctx context.Context, name string) (types.Database, error) {
	teleportClient, release, err := r.TeleportClientAccessor(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer release()

	database, err := teleportClient.GetDatabase(ctx, name)
	return database, trace.Wrap(err)
}

// Create creates a Teleport database
func (r databaseClient) Create(ctx context.Context, database types.Database) error {
	teleportClient, release, err := r.TeleportClientAccessor(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	defer release()

	return trace.Wrap(teleportClient.CreateDatabase(ctx, database))
}

// Update updates a Teleport database
func (r databaseClient) Update(ctx context.Context, database types.Database) error {
	teleportClient, release, err := r.TeleportClientAccessor(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	defer release()

	return trace.Wrap(teleportClient.UpdateDatabase(ctx, database))
}

// Delete deletes a Teleport database
func (r databaseClient) Delete(ctx context.Context, name string) error {
	teleportClient, release, err := r.TeleportClientAccessor(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	defer release()

	return trace.Wrap(teleportClient.DeleteDatabase(ctx, name))
}

// NewDatabaseV3Reconciler instantiates a new Kubernetes controller reconciling database v3 resources
func NewDatabaseV3Reconciler(client kclient.Client, accessor sidecar.ClientAccessor) *TeleportResourceReconciler[types.Database, *resourcesv1.TeleportDatabaseV3] {
	databaseClient := &databaseClient{
		TeleportClientAccessor: accessor,
	}

	resourceReconciler := NewTeleportResourceReconciler[types.Database, *resourcesv1.TeleportDatabaseV3](
		client,
		databaseClient,
	)

	return resourceReconciler
}
