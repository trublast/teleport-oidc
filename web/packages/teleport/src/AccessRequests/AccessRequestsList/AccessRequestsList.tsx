/**
 * Copyright 2026 Flant JSC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useCallback, useEffect, useState } from 'react';
import styled from 'styled-components';
import {
  Alert,
  Box,
  ButtonPrimary,
  ButtonSecondary,
  ButtonWarning,
  Flex,
  Indicator,
  Text,
} from 'design';
import Table, { Cell } from 'design/DataTable';
import { displayDateTime } from 'shared/services/loc';

import {
  FeatureBox,
  FeatureHeader,
  FeatureHeaderTitle,
} from 'teleport/components/Layout';

import { AccessRequest } from 'teleport/AccessRequests/types';
import {
  deleteAccessRequest,
  fetchAccessRequests,
  reviewAccessRequest,
} from 'teleport/AccessRequests/service';
import { NewAccessRequestDialog } from 'teleport/AccessRequests/NewAccessRequestDialog/NewAccessRequestDialog';
import useTeleport from 'teleport/useTeleport';
import session from 'teleport/services/websession';

type ViewFilter = 'all' | 'PENDING' | 'APPROVED' | 'DENIED';

export function AccessRequestsList() {
  const ctx = useTeleport();
  const requestableRoles = ctx.storeUser.getRequestableRoles();
  const suggestedReviewers = ctx.storeUser.getSuggestedReviewers();
  const accessStrategy = ctx.storeUser.getAccessStrategy();
  const currentUser = ctx.storeUser.state.username;
  const activeRequestId = ctx.storeUser.getAccessRequestId();
  const canDeleteRequests = ctx.storeUser.getAccessRequestAccess().remove;

  const [requests, setRequests] = useState<AccessRequest[]>([]);
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>(
    'loading'
  );
  const [error, setError] = useState('');
  const [filter, setFilter] = useState<ViewFilter>('all');
  const [actionError, setActionError] = useState('');
  const [showNewDialog, setShowNewDialog] = useState(false);

  const load = useCallback(async () => {
    setStatus('loading');
    setError('');
    try {
      const items = await fetchAccessRequests();
      setRequests(items);
      setStatus('success');
    } catch (err) {
      setError(err.message || 'Failed to load access requests');
      setStatus('error');
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  async function handleReview(id: string, state: 'APPROVED' | 'DENIED') {
    setActionError('');
    try {
      const updated = await reviewAccessRequest(id, state, '');
      setRequests(prev => prev.map(r => (r.id === id ? updated : r)));
    } catch (err) {
      setActionError(err.message || `Failed to ${state.toLowerCase()} request`);
    }
  }

  async function handleDelete(id: string) {
    setActionError('');
    try {
      await deleteAccessRequest(id);
      setRequests(prev => prev.filter(r => r.id !== id));
    } catch (err) {
      setActionError(err.message || 'Failed to delete request');
    }
  }

  async function handleAssume(id: string) {
    setActionError('');
    try {
      await session.renewSession({ requestId: id });
      window.location.reload();
    } catch (err) {
      setActionError(err.message || 'Failed to assume roles');
    }
  }

  async function handleSwitchback() {
    setActionError('');
    try {
      await session.renewSession({ switchback: true });
      window.location.reload();
    } catch (err) {
      setActionError(err.message || 'Failed to switch back');
    }
  }

  const filtered =
    filter === 'all' ? requests : requests.filter(r => r.state === filter);

  return (
    <FeatureBox>
      <FeatureHeader alignItems="center" justifyContent="space-between">
        <FeatureHeaderTitle>Access Requests</FeatureHeaderTitle>
        <Flex gap={2}>
          {activeRequestId && (
            <ButtonWarning size="small" onClick={handleSwitchback}>
              Switch Back
            </ButtonWarning>
          )}
          <ButtonPrimary size="small" onClick={() => setShowNewDialog(true)}>
            New Request
          </ButtonPrimary>
          <ButtonSecondary size="small" onClick={load}>
            Refresh
          </ButtonSecondary>
        </Flex>
      </FeatureHeader>

      {actionError && <Alert mb={3}>{actionError}</Alert>}

      {showNewDialog && (
        <NewAccessRequestDialog
          requestableRoles={requestableRoles}
          suggestedReviewers={suggestedReviewers}
          accessStrategy={accessStrategy}
          onClose={() => setShowNewDialog(false)}
          onCreated={req => setRequests(prev => [req, ...prev])}
        />
      )}

      <Flex gap={2} mb={3}>
        {(['all', 'PENDING', 'APPROVED', 'DENIED'] as ViewFilter[]).map(f => (
          <FilterButton
            key={f}
            size="small"
            $active={filter === f}
            onClick={() => setFilter(f)}
          >
            {f === 'all' ? 'All' : capitalize(f)}
          </FilterButton>
        ))}
      </Flex>

      {status === 'loading' && (
        <Box textAlign="center" p={4}>
          <Indicator />
        </Box>
      )}

      {status === 'error' && <Alert>{error}</Alert>}

      {status === 'success' && (
        <Table
          data={filtered}
          columns={[
            {
              key: 'id',
              headerText: 'ID',
              render: (r: AccessRequest) => (
                <Cell>
                  <Text
                    css={`
                      white-space: nowrap;
                    `}
                  >
                    {r.id}
                  </Text>
                </Cell>
              ),
            },
            {
              key: 'user',
              headerText: 'User',
            },
            {
              key: 'state',
              headerText: 'State',
              render: (r: AccessRequest) => (
                <Cell>
                  <StateBadge state={r.state} />
                </Cell>
              ),
            },
            {
              altKey: 'roles',
              headerText: 'Roles',
              render: (r: AccessRequest) => (
                <Cell>
                  <Text>{r.roles.join(', ') || '—'}</Text>
                </Cell>
              ),
            },
            {
              altKey: 'reason',
              headerText: 'Reason',
              render: (r: AccessRequest) => (
                <Cell>
                  <Text
                    css={`
                      max-width: 200px;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: nowrap;
                    `}
                    title={r.requestReason}
                  >
                    {r.requestReason || '—'}
                  </Text>
                </Cell>
              ),
            },
            {
              altKey: 'created',
              headerText: 'Created',
              render: (r: AccessRequest) => (
                <Cell>{displayDateTime(r.created)}</Cell>
              ),
            },
            {
              altKey: 'expires',
              headerText: 'Expires',
              render: (r: AccessRequest) => (
                <Cell>{displayDateTime(r.expires)}</Cell>
              ),
            },
            {
              altKey: 'actions',
              headerText: 'Actions',
              render: (r: AccessRequest) => (
                <Cell>
                  <Flex gap={1}>
                    {r.state === 'PENDING' && r.user !== currentUser && (
                      <>
                        <ButtonPrimary
                          size="small"
                          onClick={() => handleReview(r.id, 'APPROVED')}
                        >
                          Approve
                        </ButtonPrimary>
                        <ButtonWarning
                          size="small"
                          onClick={() => handleReview(r.id, 'DENIED')}
                        >
                          Deny
                        </ButtonWarning>
                      </>
                    )}
                    {r.state === 'APPROVED' &&
                      r.user === currentUser &&
                      r.id !== activeRequestId && (
                        <ButtonPrimary
                          size="small"
                          onClick={() => handleAssume(r.id)}
                        >
                          Assume Roles
                        </ButtonPrimary>
                      )}
                    {r.id === activeRequestId && (
                      <Text
                        fontSize={1}
                        color="success.main"
                        alignSelf="center"
                      >
                        Active
                      </Text>
                    )}
                    {canDeleteRequests && (
                      <ButtonSecondary
                        size="small"
                        onClick={() => handleDelete(r.id)}
                      >
                        Delete
                      </ButtonSecondary>
                    )}
                  </Flex>
                </Cell>
              ),
            },
          ]}
          emptyText="No Access Requests Found"
          isSearchable
        />
      )}
    </FeatureBox>
  );
}

function StateBadge({ state }: { state: string }) {
  let color = 'text.main';
  let bg = 'spotBackground.0';

  switch (state) {
    case 'PENDING':
      color = 'text.primaryInverse';
      bg = 'warning.main';
      break;
    case 'APPROVED':
      color = 'text.primaryInverse';
      bg = 'success.main';
      break;
    case 'DENIED':
      color = 'text.primaryInverse';
      bg = 'error.main';
      break;
    case 'APPLIED':
      color = 'text.primaryInverse';
      bg = 'brand';
      break;
  }

  return (
    <Badge bg={bg} color={color}>
      {capitalize(state)}
    </Badge>
  );
}

const Badge = styled.span<{ bg: string; color: string }>`
  display: inline-block;
  font-size: 11px;
  font-weight: 500;
  padding: 2px 10px;
  border-radius: 10px;
  white-space: nowrap;
  background: ${p => resolveThemeColor(p.theme, p.bg)};
  color: ${p => resolveThemeColor(p.theme, p.color)};
`;

function resolveThemeColor(theme: any, path: string): string {
  return path.split('.').reduce((obj, key) => obj?.[key], theme.colors) || path;
}

const FilterButton = styled(ButtonSecondary)<{ $active: boolean }>`
  ${p =>
    p.$active &&
    `
    background: ${p.theme.colors.brand};
    color: ${p.theme.colors.text.primaryInverse};
    &:hover {
      background: ${p.theme.colors.brand};
    }
  `}
`;

function capitalize(s: string): string {
  if (!s) return s;
  return s.charAt(0).toUpperCase() + s.slice(1).toLowerCase();
}
