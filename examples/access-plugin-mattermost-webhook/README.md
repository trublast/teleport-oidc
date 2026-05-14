# Mattermost webhook Access Request plugin

Minimal Teleport Access Request plugin that posts colored cards to a
Mattermost **incoming webhook** every time an access request is created,
reviewed, approved, denied, or deleted.

Designed for Teleport OSS — uses only the public gRPC API and Mattermost's
plain HTTP webhooks, so it has no dependency on enterprise features.

## How it compares to the bundled `teleport-mattermost` plugin

| | This example | `integrations/access/mattermost` |
| --- | --- | --- |
| Transport to Mattermost | Incoming webhook (HTTP) | Bot user + Mattermost API |
| Two-way (approve from chat) | No | Yes |
| State persistence | None (one message per event) | Plugin data, message edits, threads |
| Code size | ~300 lines | several thousand lines |
| Auth required | Identity file only | Identity file + Mattermost OAuth |

If you just need "notify our team channel when someone requests access", this
plugin is enough.

## Build

```bash
cd examples/access-plugin-mattermost-webhook
go mod tidy
go build -o teleport-mattermost-webhook .
```

The `go.mod` uses a `replace` directive pointing at `../../api`, so the build
tracks the version of the Teleport `api/` module that lives in this checkout.

## Configure Teleport

1. Create a role for the plugin:

   ```yaml
   # access-plugin-role.yaml
   kind: role
   version: v6
   metadata:
     name: access-plugin
   spec:
     allow:
       rules:
         - resources: ['access_request']
           verbs: ['list', 'read']
   ```

   ```bash
   tctl create -f access-plugin-role.yaml
   ```

2. Create a user for the plugin and issue an identity file:

   ```bash
   tctl users add access-plugin --roles=access-plugin
   tctl auth sign \
     --user=access-plugin \
     --out=/etc/teleport/access-plugin.pem \
     --format=file \
     --ttl=8760h
   ```

## Configure Mattermost

In Mattermost: **System Console → Integrations → Incoming Webhooks → Add**.
Pick the target channel; you'll get a URL like:

```
https://mattermost.example.com/hooks/xxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Run

The plugin reads all configuration from environment variables:

| Variable | Required | Description |
| --- | --- | --- |
| `TELEPORT_PROXY_ADDR` | yes | `host:port` of the Teleport Proxy, e.g. `asus:3080` |
| `TELEPORT_IDENTITY_FILE` | yes | Path to the identity file from `tctl auth sign` |
| `TELEPORT_INSECURE` | no | Set to `1` or `true` to skip TLS verification for proxy discovery (`/webapi/find`, TLS routing). Same idea as `tsh --insecure`. **Do not use in production.** |
| `MATTERMOST_WEBHOOK_URL` | yes | Mattermost incoming webhook URL |
| `MATTERMOST_CHANNEL` | no | Override the channel configured on the webhook |
| `MATTERMOST_USERNAME` | no | Bot display name, default `Teleport` |
| `MATTERMOST_INSECURE` | no | Set to `1` or `true` to skip TLS verification when POSTing to the webhook URL (self-signed Mattermost). **Do not use in production.** |
| `TELEPORT_WEB_URL` | no | Public URL of the Teleport web UI, used to build a "View Request" link |

```bash
export TELEPORT_PROXY_ADDR="asus:3080"
export TELEPORT_IDENTITY_FILE="/etc/teleport/access-plugin.pem"
export TELEPORT_INSECURE=1
export MATTERMOST_WEBHOOK_URL="https://mattermost.example.com/hooks/xxxxxxxx"
export TELEPORT_WEB_URL="https://asus:3080"

./teleport-mattermost-webhook
```

You should see:

```
teleport-mattermost-webhook started, proxy=asus:3080 webhook=https://mattermost.example.com/hooks/***
watcher ready, listening for access request events
```

Create a request to test:

```bash
tsh request create --roles=dba --reason="testing the plugin"
```

A yellow card should appear in the Mattermost channel within a second.

## Run as a systemd service

```ini
# /etc/systemd/system/teleport-mattermost-webhook.service
[Unit]
Description=Teleport access request → Mattermost webhook
After=network.target

[Service]
Environment=TELEPORT_PROXY_ADDR=asus:3080
Environment=TELEPORT_IDENTITY_FILE=/etc/teleport/access-plugin.pem
Environment=TELEPORT_INSECURE=1
Environment=MATTERMOST_WEBHOOK_URL=https://mattermost.example.com/hooks/xxxxxxxx
Environment=TELEPORT_WEB_URL=https://asus:3080
ExecStart=/usr/local/bin/teleport-mattermost-webhook
Restart=on-failure
RestartSec=5s
User=teleport

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now teleport-mattermost-webhook
sudo journalctl -fu teleport-mattermost-webhook
```
