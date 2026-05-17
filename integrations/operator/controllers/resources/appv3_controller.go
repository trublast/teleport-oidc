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

// appClient implements TeleportResourceClient and offers CRUD methods needed to reconcile apps
type appClient struct {
	TeleportClientAccessor sidecar.ClientAccessor
}

// Get gets the Teleport app of a given name
func (r appClient) Get(ctx context.Context, name string) (types.Application, error) {
	teleportClient, release, err := r.TeleportClientAccessor(ctx)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer release()

	app, err := teleportClient.GetApp(ctx, name)
	return app, trace.Wrap(err)
}

// Create creates a Teleport app
func (r appClient) Create(ctx context.Context, app types.Application) error {
	teleportClient, release, err := r.TeleportClientAccessor(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	defer release()

	return trace.Wrap(teleportClient.CreateApp(ctx, app))
}

// Update updates a Teleport app
func (r appClient) Update(ctx context.Context, app types.Application) error {
	teleportClient, release, err := r.TeleportClientAccessor(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	defer release()

	return trace.Wrap(teleportClient.UpdateApp(ctx, app))
}

// Delete deletes a Teleport app
func (r appClient) Delete(ctx context.Context, name string) error {
	teleportClient, release, err := r.TeleportClientAccessor(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	defer release()

	return trace.Wrap(teleportClient.DeleteApp(ctx, name))
}

// NewAppV3Reconciler instantiates a new Kubernetes controller reconciling app v3 resources
func NewAppV3Reconciler(client kclient.Client, accessor sidecar.ClientAccessor) *TeleportResourceReconciler[types.Application, *resourcesv1.TeleportAppV3] {
	appClient := &appClient{
		TeleportClientAccessor: accessor,
	}

	resourceReconciler := NewTeleportResourceReconciler[types.Application, *resourcesv1.TeleportAppV3](
		client,
		appClient,
	)

	return resourceReconciler
}
