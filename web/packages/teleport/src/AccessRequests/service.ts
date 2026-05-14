/**
 * Copyright 2023 Gravitational, Inc.
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

import { formatDuration } from 'date-fns';

import { AccessRequestResource } from 'teleport/Assist/types';
import { ResourceIdKind } from 'teleport/services/agents';
import api from 'teleport/services/api';
import cfg from 'teleport/config';
import {
  AccessRequest,
  CreateAccessRequest,
  DurationOption,
} from 'teleport/AccessRequests/types';
import { middleValues } from 'teleport/AccessRequests/utils';

function makeAccessRequest(raw: any): AccessRequest {
  return {
    id: raw.id,
    created: new Date(raw.created),
    expires: new Date(raw.expires),
    maxDuration: raw.maxDuration ? new Date(raw.maxDuration) : undefined,
    sessionTTL: raw.sessionTTL ? new Date(raw.sessionTTL) : undefined,
    requestReason: raw.requestReason || '',
    resolveReason: raw.resolveReason || '',
    resources: raw.resources || [],
    reviews: raw.reviews || [],
    roles: raw.roles || [],
    state: raw.state || 'PENDING',
    suggestedReviewers: raw.suggestedReviewers || [],
    thresholdNames: raw.thresholdNames || [],
    user: raw.user || '',
  };
}

export async function fetchAccessRequests(): Promise<AccessRequest[]> {
  const items = await api.get(cfg.getAccessRequestUrl());
  if (Array.isArray(items)) {
    return items.map(makeAccessRequest);
  }
  return [];
}

export async function reviewAccessRequest(
  requestId: string,
  state: 'APPROVED' | 'DENIED',
  reason: string
): Promise<AccessRequest> {
  const resp = await api.post(`${cfg.getAccessRequestUrl(requestId)}/review`, {
    state,
    reason,
  });
  return makeAccessRequest(resp);
}

export async function deleteAccessRequest(requestId: string): Promise<void> {
  await api.delete(cfg.getAccessRequestUrl(requestId));
}

export async function createAccessRequest(
  clusterId: string,
  roles: string[],
  resources: AccessRequestResource[],
  reason: string,
  dryRun: boolean,
  maxDuration?: Date,
  assumeStartTime?: Date,
  suggestedReviewers: string[] = []
): Promise<AccessRequest> {
  const request: CreateAccessRequest = {
    reason,
    roles,
    resourceIds: resources.map(item => ({
      kind: item.type as ResourceIdKind,
      name: item.id,
      clusterName: clusterId,
    })),
    suggestedReviewers,
    maxDuration,
    assumeStartTime,
    dryRun,
  };

  const raw = await api.post(cfg.getAccessRequestUrl(), request);
  return makeAccessRequest(raw);
}

export async function getDurationOptions(
  clusterId: string,
  roles: string[],
  resources: AccessRequestResource[]
): Promise<DurationOption[]> {
  const accessRequest = await createAccessRequest(
    clusterId,
    roles,
    resources,
    '',
    true
  );

  if (!accessRequest.sessionTTL || !accessRequest.maxDuration) {
    return [];
  }

  return middleValues(
    accessRequest.created,
    accessRequest.sessionTTL,
    accessRequest.maxDuration
  ).map(duration => ({
    value: duration.timestamp,
    label: formatDuration(duration.duration),
  }));
}
