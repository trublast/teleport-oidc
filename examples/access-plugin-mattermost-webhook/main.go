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

// Command teleport-mattermost-webhook is a minimal Access Request plugin
// for Teleport OSS. It subscribes to access request events over the Teleport
// gRPC API and posts notifications to a Mattermost incoming webhook.
//
// See README.md for setup steps.
package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/gravitational/teleport/api/client"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("shutdown requested")
		cancel()
	}()

	creds := client.LoadIdentityFile(cfg.IdentityFile)

	if cfg.Insecure {
		log.Println("warning: TELEPORT_INSECURE is set — TLS certificate verification is disabled for proxy discovery")
	}

	teleClient, err := client.New(rootCtx, client.Config{
		Addrs:                    []string{cfg.ProxyAddr},
		Credentials:              []client.Credentials{creds},
		InsecureAddressDiscovery: cfg.Insecure,
		DialOpts: []grpc.DialOption{
			grpc.WithReturnConnectionError(),
		},
	})
	if err != nil {
		log.Fatalf("teleport client: %v", err)
	}
	defer teleClient.Close()

	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	if envBool("MATTERMOST_INSECURE") {
		log.Println("warning: MATTERMOST_INSECURE is set — TLS verification disabled for Mattermost webhook HTTPS")
		if httpTransport.TLSClientConfig == nil {
			httpTransport.TLSClientConfig = &tls.Config{}
		}
		httpTransport.TLSClientConfig.InsecureSkipVerify = true
	}

	bot := &mattermostBot{
		webhookURL: cfg.WebhookURL,
		channel:    cfg.Channel,
		username:   cfg.Username,
		webURL:     cfg.WebURL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: httpTransport,
		},
	}

	plugin := &accessRequestPlugin{
		teleport: teleClient,
		bot:      bot,
	}

	log.Printf("teleport-mattermost-webhook started, proxy=%s webhook=%s",
		cfg.ProxyAddr, redactWebhook(cfg.WebhookURL))

	if err := plugin.run(rootCtx); err != nil && rootCtx.Err() == nil {
		log.Fatalf("plugin terminated: %v", err)
	}
	log.Println("teleport-mattermost-webhook stopped")
}
