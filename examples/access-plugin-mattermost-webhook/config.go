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
	"errors"
	"net/url"
	"os"
	"strings"
)

type config struct {
	// Insecure disables TLS verification when the plugin discovers the proxy
	// (TLS routing, /webapi/find). Set TELEPORT_INSECURE=1 for self-signed
	// clusters — same idea as tsh --insecure.
	Insecure bool
	// ProxyAddr is the host:port of the Teleport Proxy Service (e.g.
	// "asus:3080").
	ProxyAddr string
	// IdentityFile is the path to the identity file produced with
	// `tctl auth sign --format=file`.
	IdentityFile string
	// WebhookURL is the Mattermost incoming webhook URL.
	WebhookURL string
	// Channel optionally overrides the channel configured on the webhook.
	Channel string
	// Username optionally overrides the bot display name.
	Username string
	// WebURL is the public URL of the Teleport web UI used to build a
	// "View Request" link (e.g. "https://asus:3080"). Optional.
	WebURL string
}

func loadConfig() (*config, error) {
	cfg := &config{
		Insecure:     envBool("TELEPORT_INSECURE"),
		ProxyAddr:    os.Getenv("TELEPORT_PROXY_ADDR"),
		IdentityFile: os.Getenv("TELEPORT_IDENTITY_FILE"),
		WebhookURL:   os.Getenv("MATTERMOST_WEBHOOK_URL"),
		Channel:      os.Getenv("MATTERMOST_CHANNEL"),
		Username:     os.Getenv("MATTERMOST_USERNAME"),
		WebURL:       strings.TrimRight(os.Getenv("TELEPORT_WEB_URL"), "/"),
	}

	if cfg.Username == "" {
		cfg.Username = "Teleport"
	}

	switch {
	case cfg.ProxyAddr == "":
		return nil, errors.New("TELEPORT_PROXY_ADDR must be set")
	case cfg.IdentityFile == "":
		return nil, errors.New("TELEPORT_IDENTITY_FILE must be set")
	case cfg.WebhookURL == "":
		return nil, errors.New("MATTERMOST_WEBHOOK_URL must be set")
	}

	if _, err := url.ParseRequestURI(cfg.WebhookURL); err != nil {
		return nil, errors.New("MATTERMOST_WEBHOOK_URL is not a valid URL")
	}

	if _, err := os.Stat(cfg.IdentityFile); err != nil {
		return nil, errors.New("TELEPORT_IDENTITY_FILE does not exist: " + cfg.IdentityFile)
	}

	return cfg, nil
}

func envBool(name string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// redactWebhook trims the token component from a webhook URL so it can be
// safely logged. Mattermost webhooks have the shape ".../hooks/<token>".
func redactWebhook(u string) string {
	idx := strings.Index(u, "/hooks/")
	if idx < 0 {
		return "<webhook>"
	}
	return u[:idx+len("/hooks/")] + "***"
}
