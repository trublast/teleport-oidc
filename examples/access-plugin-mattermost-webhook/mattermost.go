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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types"
)

// mattermostBot is a thin wrapper around an incoming webhook URL.
type mattermostBot struct {
	webhookURL string
	channel    string
	username   string
	webURL     string
	httpClient *http.Client
}

// webhookPayload is the body accepted by Mattermost incoming webhooks.
// See https://developers.mattermost.com/integrate/webhooks/incoming/.
type webhookPayload struct {
	Text        string       `json:"text,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []attachment `json:"attachments,omitempty"`
}

type attachment struct {
	Color     string   `json:"color,omitempty"`
	Pretext   string   `json:"pretext,omitempty"`
	Title     string   `json:"title,omitempty"`
	TitleLink string   `json:"title_link,omitempty"`
	Text      string   `json:"text,omitempty"`
	Fields    []field  `json:"fields,omitempty"`
	Footer    string   `json:"footer,omitempty"`
	MrkdwnIn  []string `json:"mrkdwn_in,omitempty"`
}

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

const (
	colorPending  = "#FFAB00"
	colorApproved = "#36A64F"
	colorDenied   = "#D32F2F"
	colorPromoted = "#1976D2"
	colorDeleted  = "#9E9E9E"
)

// notify posts a colored attachment describing the current state of the
// access request.
func (b *mattermostBot) notify(ctx context.Context, req types.AccessRequest) error {
	state := req.GetState()

	var (
		color   string
		pretext string
		emoji   string
	)
	switch {
	case state.IsPending():
		color, pretext, emoji = colorPending, ":hourglass_flowing_sand: New access request", ":hourglass_flowing_sand:"
	case state.IsApproved():
		color, pretext, emoji = colorApproved, ":white_check_mark: Access request approved", ":white_check_mark:"
	case state.IsDenied():
		color, pretext, emoji = colorDenied, ":no_entry: Access request denied", ":no_entry:"
	case state.IsPromoted():
		color, pretext, emoji = colorPromoted, ":arrow_up: Access request promoted to access list", ":arrow_up:"
	default:
		// Unknown state — skip rather than spam the channel.
		return nil
	}

	fields := []field{
		{Title: "User", Value: req.GetUser(), Short: true},
		{Title: "Roles", Value: codeJoin(req.GetRoles()), Short: true},
	}
	if reviewers := req.GetSuggestedReviewers(); len(reviewers) > 0 {
		fields = append(fields, field{
			Title: "Suggested reviewers",
			Value: codeJoin(reviewers),
			Short: false,
		})
	}
	if reason := req.GetRequestReason(); reason != "" {
		fields = append(fields, field{
			Title: "Reason",
			Value: reason,
			Short: false,
		})
	}
	if resolveReason := req.GetResolveReason(); resolveReason != "" {
		fields = append(fields, field{
			Title: "Resolve reason",
			Value: resolveReason,
			Short: false,
		})
	}
	if reviews := req.GetReviews(); len(reviews) > 0 {
		var sb strings.Builder
		for _, r := range reviews {
			fmt.Fprintf(&sb, "- **%s** → %s", r.Author, r.ProposedState.String())
			if r.Reason != "" {
				fmt.Fprintf(&sb, " (%s)", r.Reason)
			}
			sb.WriteString("\n")
		}
		fields = append(fields, field{
			Title: "Reviews",
			Value: sb.String(),
			Short: false,
		})
	}

	payload := webhookPayload{
		Channel:   b.channel,
		Username:  b.username,
		IconEmoji: emoji,
		Attachments: []attachment{{
			Color:     color,
			Pretext:   pretext,
			Title:     fmt.Sprintf("Request %s", shortID(req.GetName())),
			TitleLink: b.requestLink(req.GetName()),
			Fields:    fields,
			Footer:    "Teleport access-request plugin",
			MrkdwnIn:  []string{"pretext", "fields", "text"},
		}},
	}

	return b.post(ctx, payload)
}

// notifyDeleted posts a small message when an access request is removed.
func (b *mattermostBot) notifyDeleted(ctx context.Context, requestID string) error {
	payload := webhookPayload{
		Channel:   b.channel,
		Username:  b.username,
		IconEmoji: ":wastebasket:",
		Attachments: []attachment{{
			Color:   colorDeleted,
			Pretext: ":wastebasket: Access request deleted",
			Title:   fmt.Sprintf("Request %s", shortID(requestID)),
			Footer:  "Teleport access-request plugin",
		}},
	}
	return b.post(ctx, payload)
}

func (b *mattermostBot) post(ctx context.Context, payload webhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return trace.Wrap(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.webhookURL, bytes.NewReader(body))
	if err != nil {
		return trace.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return trace.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<10))
	return trace.Errorf("mattermost webhook returned %s: %s", resp.Status, bytes.TrimSpace(respBody))
}

func (b *mattermostBot) requestLink(requestID string) string {
	if b.webURL == "" {
		return ""
	}
	// The OSS access-request page lists all requests; deep-linking to a
	// single request is not yet supported, so we point at the list view.
	return b.webURL + "/web/accessrequest"
}

func codeJoin(items []string) string {
	if len(items) == 0 {
		return "_(none)_"
	}
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = "`" + s + "`"
	}
	return strings.Join(quoted, ", ")
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8] + "…"
}
