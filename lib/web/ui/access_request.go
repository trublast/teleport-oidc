/**
 * Copyright 2026 Flant JSC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ui

import (
	"time"

	"github.com/gravitational/teleport/api/types"
)

// AccessRequestReview describes a review applied to an access request,
// suitable for the webapp.
type AccessRequestReview struct {
	Author        string    `json:"author"`
	Roles         []string  `json:"roles,omitempty"`
	ProposedState string    `json:"proposedState"`
	Reason        string    `json:"reason,omitempty"`
	Created       time.Time `json:"created"`
}

// AccessRequestResource describes a requested resource in webapp form.
type AccessRequestResource struct {
	ID AccessRequestResourceID `json:"id"`
}

// AccessRequestResourceID is the identifier of a requested resource.
type AccessRequestResourceID struct {
	Kind            string `json:"kind"`
	Name            string `json:"name"`
	ClusterName     string `json:"clusterName"`
	SubResourceName string `json:"subResourceName,omitempty"`
}

// AccessRequest describes a Teleport access request suitable for the webapp.
type AccessRequest struct {
	ID                 string                  `json:"id"`
	User               string                  `json:"user"`
	Roles              []string                `json:"roles"`
	State              string                  `json:"state"`
	RequestReason      string                  `json:"requestReason"`
	ResolveReason      string                  `json:"resolveReason"`
	Created            time.Time               `json:"created"`
	Expires            time.Time               `json:"expires"`
	MaxDuration        *time.Time              `json:"maxDuration,omitempty"`
	SessionTTL         *time.Time              `json:"sessionTTL,omitempty"`
	SuggestedReviewers []string                `json:"suggestedReviewers"`
	ThresholdNames     []string                `json:"thresholdNames"`
	Reviews            []AccessRequestReview   `json:"reviews"`
	Resources          []AccessRequestResource `json:"resources"`
}

// MakeAccessRequest converts a backend access request resource into
// the webapp-friendly form.
func MakeAccessRequest(req types.AccessRequest) AccessRequest {
	roles := req.GetRoles()
	if roles == nil {
		roles = []string{}
	}

	suggestedReviewers := req.GetSuggestedReviewers()
	if suggestedReviewers == nil {
		suggestedReviewers = []string{}
	}

	thresholdNames := make([]string, 0, len(req.GetThresholds()))
	for _, th := range req.GetThresholds() {
		if th.Name != "" {
			thresholdNames = append(thresholdNames, th.Name)
		}
	}

	rawReviews := req.GetReviews()
	reviews := make([]AccessRequestReview, 0, len(rawReviews))
	for _, rev := range rawReviews {
		reviews = append(reviews, AccessRequestReview{
			Author:        rev.Author,
			Roles:         rev.Roles,
			ProposedState: rev.ProposedState.String(),
			Reason:        rev.Reason,
			Created:       rev.Created,
		})
	}

	rawResources := req.GetRequestedResourceIDs()
	resources := make([]AccessRequestResource, 0, len(rawResources))
	for _, r := range rawResources {
		resources = append(resources, AccessRequestResource{
			ID: AccessRequestResourceID{
				Kind:            r.Kind,
				Name:            r.Name,
				ClusterName:     r.ClusterName,
				SubResourceName: r.SubResourceName,
			},
		})
	}

	ar := AccessRequest{
		ID:                 req.GetName(),
		User:               req.GetUser(),
		Roles:              roles,
		State:              req.GetState().String(),
		RequestReason:      req.GetRequestReason(),
		ResolveReason:      req.GetResolveReason(),
		Created:            req.GetCreationTime(),
		Expires:            req.GetAccessExpiry(),
		SuggestedReviewers: suggestedReviewers,
		ThresholdNames:     thresholdNames,
		Reviews:            reviews,
		Resources:          resources,
	}

	if maxDuration := req.GetMaxDuration(); !maxDuration.IsZero() {
		ar.MaxDuration = &maxDuration
	}
	if sessionTTL := req.GetSessionTLL(); !sessionTTL.IsZero() {
		ar.SessionTTL = &sessionTTL
	}

	return ar
}

// MakeAccessRequests converts a list of backend access request resources
// into the webapp-friendly form.
func MakeAccessRequests(requests []types.AccessRequest) []AccessRequest {
	uiRequests := make([]AccessRequest, 0, len(requests))
	for _, req := range requests {
		uiRequests = append(uiRequests, MakeAccessRequest(req))
	}
	return uiRequests
}
