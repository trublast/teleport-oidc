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
import React, { useEffect, useState } from 'react';
import { formatDistanceToNow } from 'date-fns';
import styled from 'styled-components';
import { Text } from 'design';

import { Key, Notification as NotificationIcon, UserList } from 'design/Icon';
import { useRefClickOutside } from 'shared/hooks/useRefClickOutside';
import { useStore } from 'shared/libs/stores';
import { assertUnreachable } from 'shared/utils/assertUnreachable';

import {
  Dropdown,
  DropdownItem,
  DropdownItemButton,
  DropdownItemIcon,
  STARTING_TRANSITION_DELAY,
  INCREMENT_TRANSITION_DELAY,
  DropdownItemLink,
} from 'teleport/components/Dropdown';
import useTeleport from 'teleport/useTeleport';
import cfg from 'teleport/config';
import { fetchAccessRequests } from 'teleport/AccessRequests/service';
import {
  Notification,
  NotificationKind,
} from 'teleport/stores/storeNotifications';

import { ButtonIconContainer } from '../Shared';

const ACCESS_REQUEST_POLL_INTERVAL_MS = 60_000;

export function Notifications() {
  const ctx = useTeleport();
  useStore(ctx.storeNotifications);

  const notices = ctx.storeNotifications.getNotifications();

  const [open, setOpen] = useState(false);

  const ref = useRefClickOutside<HTMLDivElement>({ open, setOpen });

  const username = ctx.storeUser.state.username;
  const activeRequestId = ctx.storeUser.getAccessRequestId();

  useEffect(() => {
    // Poll unconditionally. We intentionally don't gate on
    // `getAccessRequestAccess().list` here because Teleport's preset roles
    // (editor/access) often don't declare an explicit `kind: access_request`
    // rule even though the auth server allows the underlying gRPC call. If the
    // user genuinely lacks access, the request below will fail and be silently
    // swallowed by the try/catch.
    if (!username) {
      return;
    }
    let cancelled = false;

    const refresh = async () => {
      try {
        const requests = await fetchAccessRequests();
        if (cancelled) return;

        const route = cfg.routes.accessRequest;
        const pending: Notification[] = requests
          .filter(r => r.state === 'PENDING' && r.user !== username)
          .map(r => ({
            id: `access-request-pending:${r.id}`,
            date: r.created,
            item: {
              kind: NotificationKind.AccessRequestPending,
              requestId: r.id,
              requestor: r.user,
              roles: r.roles,
              route,
            },
          }));

        const approved: Notification[] = requests
          .filter(
            r =>
              r.state === 'APPROVED' &&
              r.user === username &&
              r.id !== activeRequestId
          )
          .map(r => ({
            id: `access-request-approved:${r.id}`,
            date: r.created,
            item: {
              kind: NotificationKind.AccessRequestApproved,
              requestId: r.id,
              roles: r.roles,
              route,
            },
          }));

        ctx.storeNotifications.updateNotificationsByKind(
          pending,
          NotificationKind.AccessRequestPending
        );
        ctx.storeNotifications.updateNotificationsByKind(
          approved,
          NotificationKind.AccessRequestApproved
        );
      } catch {
        // Silently ignore. The bell stays empty rather than showing an error.
      }
    };

    refresh();
    const interval = window.setInterval(
      refresh,
      ACCESS_REQUEST_POLL_INTERVAL_MS
    );
    return () => {
      cancelled = true;
      window.clearInterval(interval);
    };
  }, [ctx.storeNotifications, username, activeRequestId]);

  let transitionDelay = STARTING_TRANSITION_DELAY;
  const items = notices.map(notice => {
    const currentTransitionDelay = transitionDelay;
    transitionDelay += INCREMENT_TRANSITION_DELAY;

    return (
      <DropdownItem
        open={open}
        $transitionDelay={currentTransitionDelay}
        key={notice.id}
        data-testid="note-item"
      >
        <NotificationItem notice={notice} close={() => setOpen(false)} />
      </DropdownItem>
    );
  });

  return (
    <NotificationButtonContainer ref={ref} data-testid="tb-note">
      <ButtonIconContainer
        onClick={() => setOpen(!open)}
        data-testid="tb-note-button"
      >
        {items.length > 0 && <AttentionDot data-testid="tb-note-attention" />}
        <NotificationIcon />
      </ButtonIconContainer>

      <Dropdown
        open={open}
        style={{
          width: '300px',
          maxHeight: '80vh',
          overflowY: 'auto',
          overflowX: 'hidden',
        }}
        data-testid="tb-note-dropdown"
      >
        {items.length ? (
          items
        ) : (
          <Text textAlign="center" p={2}>
            No notifications
          </Text>
        )}
      </Dropdown>
    </NotificationButtonContainer>
  );
}

function NotificationItem({
  notice,
  close,
}: {
  notice: Notification;
  close(): void;
}) {
  const today = new Date();
  const numDays = formatDistanceToNow(notice.date);

  let dueText;
  if (notice.date <= today) {
    dueText = `was overdue for a review ${numDays} ago`;
  } else {
    dueText = `needs your review within ${numDays}`;
  }
  switch (notice.item.kind) {
    case NotificationKind.AccessList:
      return (
        <NotificationLink to={notice.item.route} onClick={close}>
          <NotificationItemButton>
            <DropdownItemIcon>
              <UserList mt="1px" />
            </DropdownItemIcon>
            <Text>
              Access list <b>{notice.item.resourceName}</b> {dueText}.
            </Text>
          </NotificationItemButton>
        </NotificationLink>
      );
    case NotificationKind.AccessRequestPending: {
      const { requestor, roles, route } = notice.item;
      return (
        <NotificationLink to={route} onClick={close}>
          <NotificationItemButton>
            <DropdownItemIcon>
              <UserList mt="1px" />
            </DropdownItemIcon>
            <Text>
              <b>{requestor}</b> requested {formatRoles(roles)} — review needed.
            </Text>
          </NotificationItemButton>
        </NotificationLink>
      );
    }
    case NotificationKind.AccessRequestApproved: {
      const { roles, route } = notice.item;
      return (
        <NotificationLink to={route} onClick={close}>
          <NotificationItemButton>
            <DropdownItemIcon>
              <Key mt="1px" />
            </DropdownItemIcon>
            <Text>
              Your request for {formatRoles(roles)} was approved. Click to
              assume.
            </Text>
          </NotificationItemButton>
        </NotificationLink>
      );
    }
    default:
      assertUnreachable(notice.item);
  }
}

function formatRoles(roles: string[]): React.ReactNode {
  if (roles.length === 0) return <b>access</b>;
  if (roles.length === 1) return <b>{roles[0]}</b>;
  return (
    <>
      <b>{roles[0]}</b> +{roles.length - 1} more
    </>
  );
}

const NotificationButtonContainer = styled.div`
  position: relative;
`;

const AttentionDot = styled.div`
  position: absolute;
  width: 7px;
  height: 7px;
  border-radius: 100px;
  background-color: ${p => p.theme.colors.buttons.warning.default};
  top: 10px;
  right: 15px;
`;

const NotificationItemButton = styled(DropdownItemButton)`
  align-items: flex-start;
  line-height: 20px;
`;

const NotificationLink = styled(DropdownItemLink)`
  padding: 0;
`;
