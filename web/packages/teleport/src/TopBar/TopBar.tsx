/*
Copyright 2019-2020 Gravitational, Inc.

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

import React, { lazy, Suspense, useState } from 'react';
import styled from 'styled-components';
import { Flex, Text, TopNav } from 'design';

import { matchPath, useHistory } from 'react-router';

import { BrainIcon } from 'design/SVGIcon';

import { ArrowLeft } from 'design/Icon';

import useTeleport from 'teleport/useTeleport';
import useStickyClusterId from 'teleport/useStickyClusterId';
import { UserMenuNav } from 'teleport/components/UserMenuNav';
import { useFeatures } from 'teleport/FeaturesContext';

import cfg from 'teleport/config';

import { useLayout } from 'teleport/Main/LayoutContext';
import { KeysEnum } from 'teleport/services/storageService';
import { getFirstRouteForCategory } from 'teleport/Navigation/Navigation';

import ClusterSelector from './ClusterSelector';
import { Notifications } from './Notifications';
import { ButtonIconContainer } from './Shared';

const Assist = lazy(() => import('teleport/Assist'));

export function TopBar() {
  const ctx = useTeleport();
  const history = useHistory();
  const features = useFeatures();

  const assistEnabled = ctx.getFeatureFlags().assist && ctx.assistEnabled;

  const [showAssist, setShowAssist] = useState(false);

  const { clusterId, hasClusterUrl } = useStickyClusterId();

  const { hasDockedElement } = useLayout();

  function loadClusters() {
    return ctx.clusterService.fetchClusters();
  }

  function changeCluster(value: string) {
    const newPrefix = cfg.getClusterRoute(value);

    const oldPrefix = cfg.getClusterRoute(clusterId);

    const newPath = history.location.pathname.replace(oldPrefix, newPrefix);

    // TODO (avatus) DELETE IN 15 (LEGACY RESOURCES SUPPORT)
    // this is a temporary hack to support leaf clusters _maybe_ not having access
    // to unified resources yet. When unified resources are loaded in fetchUnifiedResources,
    // if the response is a 404 (the endpoint doesnt exist), we:
    // 1. push them to the servers page (old default)
    // 2. set this variable conditionally render the "legacy" navigation
    // When we switch clusters (to leaf or root), we remove the item and perform the check again by pushing
    // to the resource (new default view).
    window.localStorage.removeItem(KeysEnum.UNIFIED_RESOURCES_NOT_SUPPORTED);
    // we also need to reset the pinned resources flag when we switch clusters to try again
    window.localStorage.removeItem(KeysEnum.PINNED_RESOURCES_NOT_SUPPORTED);
    const legacyResourceRoutes = [
      cfg.getNodesRoute(clusterId),
      cfg.getAppsRoute(clusterId),
      cfg.getKubernetesRoute(clusterId),
      cfg.getDatabasesRoute(clusterId),
      cfg.getDesktopsRoute(clusterId),
    ];

    if (
      legacyResourceRoutes.some(route =>
        history.location.pathname.includes(route)
      )
    ) {
      const unifiedPath = cfg
        .getUnifiedResourcesRoute(clusterId)
        .replace(oldPrefix, newPrefix);

      history.replace(unifiedPath);
      return;
    }

    // keep current view just change the clusterId
    history.push(newPath);
  }

  // find active feature
  const feature = features
    .filter(feature => Boolean(feature.route))
    .find(f =>
      matchPath(history.location.pathname, {
        path: f.route.path,
        exact: f.route.exact ?? false,
      })
    );

  function handleBack() {
    const firstRouteForCategory = getFirstRouteForCategory(
      features,
      feature.category
    );

    history.push(firstRouteForCategory);
  }

  const title = feature?.route?.title || '';

  // instead of re-creating an expensive react-select component,
  // hide/show it instead
  const styles = {
    display: !hasClusterUrl ? 'none' : 'block',
  };

  return (
    <TopBarContainer navigationHidden={feature?.hideNavigation}>
      {feature?.hideNavigation && (
        <ButtonIconContainer onClick={handleBack}>
          <ArrowLeft size="medium" />
        </ButtonIconContainer>
      )}
      {!hasClusterUrl && (
        <Text fontSize="18px" bold data-testid="title">
          {title}
        </Text>
      )}
      <Text fontSize="18px" id="topbar-portal" ml={2}></Text>
      <ClusterSelector
        value={clusterId}
        width="384px"
        maxMenuHeight={200}
        mr="20px"
        onChange={changeCluster}
        onLoad={loadClusters}
        style={styles}
      />
      <Flex ml="auto" height="100%" alignItems="center">
        {!hasDockedElement && assistEnabled && (
          <ButtonIconContainer onClick={() => setShowAssist(true)}>
            <BrainIcon />
          </ButtonIconContainer>
        )}
        <Notifications />
        <UserMenuNav username={ctx.storeUser.state.username} />
      </Flex>

      {showAssist && (
        <Suspense fallback={null}>
          <Assist onClose={() => setShowAssist(false)} />
        </Suspense>
      )}
    </TopBarContainer>
  );
}

export const TopBarContainer = styled(TopNav)`
  height: 72px;
  background-color: inherit;
  padding-left: ${p => `${p.theme.space[p.navigationHidden ? 2 : 6]}px`};
  overflow-y: initial;
  flex-shrink: 0;
  border-bottom: 1px solid ${({ theme }) => theme.colors.spotBackground[0]};
`;
