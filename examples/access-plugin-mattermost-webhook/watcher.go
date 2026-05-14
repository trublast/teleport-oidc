// Copyright 2026 Gravitational, Inc
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

package main

import (
	"context"
	"log"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/types"
)

// accessRequestPlugin wires a Teleport gRPC watcher to a notification bot.
type accessRequestPlugin struct {
	teleport *client.Client
	bot      *mattermostBot
}

func (p *accessRequestPlugin) run(ctx context.Context) error {
	watch, err := p.teleport.NewWatcher(ctx, types.Watch{
		Kinds: []types.WatchKind{
			{Kind: types.KindAccessRequest},
		},
	})
	if err != nil {
		return trace.Wrap(err)
	}
	defer watch.Close()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-watch.Done():
			if err := watch.Error(); err != nil {
				return trace.Wrap(err)
			}
			return nil

		case event := <-watch.Events():
			if err := p.handleEvent(ctx, event); err != nil {
				// Don't fail the whole watcher loop on per-event errors.
				log.Printf("event handler error: %v", err)
			}
		}
	}
}

func (p *accessRequestPlugin) handleEvent(ctx context.Context, event types.Event) error {
	if event.Resource == nil {
		return nil
	}

	// The first event after starting the watcher is an OpInit / WatchStatus
	// confirmation rather than an actual resource. Surface it once.
	if _, ok := event.Resource.(*types.WatchStatusV1); ok {
		log.Println("watcher ready, listening for access request events")
		return nil
	}

	switch event.Type {
	case types.OpPut:
		req, ok := event.Resource.(types.AccessRequest)
		if !ok {
			log.Printf("unexpected resource type %T for OpPut", event.Resource)
			return nil
		}
		return p.bot.notify(ctx, req)

	case types.OpDelete:
		// On delete the auth server only sends a resource header; we still
		// want to leave a trace in the channel.
		return p.bot.notifyDeleted(ctx, event.Resource.GetName())

	default:
		return nil
	}
}
