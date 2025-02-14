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

import React, { useState, useEffect, useMemo } from 'react';
import { SortType } from 'design/DataTable/types';
import { Flex } from 'design';
import { StyledPanel } from 'design/DataTable/StyledTable';
import { SearchPanel } from 'shared/components/Search';
import { StyledArrowBtn } from 'design/DataTable/Pager/StyledPager';
import { CircleArrowLeft, CircleArrowRight } from 'design/Icon';

import { makeAdvancedSearchQueryForLabel } from 'shared/utils/advancedSearchLabelQuery';

import { Desktop } from 'teleport/services/desktops';
import { Node } from 'teleport/services/nodes';
import { useServerSidePagination } from 'teleport/components/hooks';
import useTeleport from 'teleport/useTeleport';
import cfg, { UrlResourcesParams } from 'teleport/config';
import Ctx from 'teleport/teleportContext';

import { RoleResource } from 'teleport/services/resources';

import { TableWrapper, ServerSideListProps } from '../common';
import { CommonListProps, LockResourceKind } from '../../common';

import { Nodes } from './Nodes';
import { Desktops } from './Desktops';
import { Roles } from './Roles';

import type {
  ResourceLabel,
  ResourceFilter,
  ResourcesResponse,
} from 'teleport/services/agents';

export function ServerSideSupportedList(props: CommonListProps) {
  const ctx = useTeleport();

  const [resourceFilter, setResourceFilter] = useState<ResourceFilter>({});

  const {
    fetchStatus,
    fetchNext,
    fetchPrev,
    fetch,
    attempt: fetchAttempt,
    pageIndicators,
    fetchedData,
  } = useServerSidePagination({
    fetchFunc: getFetchFuncForServerSidePaginating(
      ctx,
      props.selectedResourceKind
    ),
    clusterId: cfg.proxyCluster, // Locking only supported with root cluster
    params: resourceFilter,
    pageSize: props.pageSize,
  });

  useEffect(() => {
    // Resetting the filter will trigger a fetch.
    setResourceFilter({
      sort: getDefaultSort(props.selectedResourceKind),
      search: '',
      query: '',
    });
  }, [props.selectedResourceKind]);

  useEffect(() => {
    fetch();
  }, [resourceFilter]);

  useEffect(() => {
    props.setAttempt(fetchAttempt);
  }, [fetchAttempt]);

  function updateSort(sort: SortType) {
    setResourceFilter({ ...resourceFilter, sort });
  }

  function updateSearch(search: string) {
    setResourceFilter({ ...resourceFilter, query: '', search });
  }

  function updateQuery(query: string) {
    setResourceFilter({ ...resourceFilter, search: '', query });
  }

  function onResourceLabelClick(label: ResourceLabel) {
    const query = makeAdvancedSearchQueryForLabel(label, resourceFilter);
    setResourceFilter({ ...resourceFilter, search: '', query });
  }

  const table = useMemo(() => {
    const listProps: ServerSideListProps = {
      fetchStatus,
      customSort: {
        dir: resourceFilter.sort?.dir,
        fieldName: resourceFilter.sort?.fieldName,
        onSort: updateSort,
      },
      onLabelClick: onResourceLabelClick,
      selectedResources: props.selectedResources,
      toggleSelectResource: props.toggleSelectResource,
    };

    switch (props.selectedResourceKind) {
      case 'role':
        return (
          <Roles roles={fetchedData.agents as RoleResource[]} {...listProps} />
        );
      case 'node':
        return <Nodes nodes={fetchedData.agents as Node[]} {...listProps} />;
      case 'windows_desktop':
        return (
          <Desktops desktops={fetchedData.agents as Desktop[]} {...listProps} />
        );
      default:
        console.error(
          `[ServerSideSupportedList.tsx] table not defined for resource kind ${props.selectedResourceKind}`
        );
    }
  }, [props.attempt, fetchedData, fetchStatus, props.selectedResources]);

  return (
    <TableWrapper
      className={fetchStatus === 'loading' ? 'disabled' : ''}
      css={`
        border-radius: 8px;
        overflow: hidden;
        box-shadow: ${props => props.theme.boxShadow[0]};
      `}
    >
      <SearchPanel
        updateQuery={updateQuery}
        updateSearch={updateSearch}
        pageIndicators={{
          from: pageIndicators.from,
          to: pageIndicators.to,
          total: pageIndicators.totalCount,
        }}
        hideAdvancedSearch={props.selectedResourceKind === 'role'} // Roles don't support advanced search.
        filter={resourceFilter}
        showSearchBar={true}
        disableSearch={fetchStatus === 'loading'}
      />
      {table}
      <StyledPanel borderBottomLeftRadius={3} borderBottomRightRadius={3}>
        <Flex justifyContent="flex-end" width="100%">
          <Flex alignItems="center" mr={2}></Flex>
          <Flex>
            <StyledArrowBtn
              onClick={fetchPrev}
              title="Previous page"
              disabled={!fetchPrev || fetchStatus === 'loading'}
              mx={0}
            >
              <CircleArrowLeft />
            </StyledArrowBtn>
            <StyledArrowBtn
              ml={0}
              onClick={fetchNext}
              title="Next page"
              disabled={!fetchNext || fetchStatus === 'loading'}
            >
              <CircleArrowRight />
            </StyledArrowBtn>
          </Flex>
        </Flex>
      </StyledPanel>
    </TableWrapper>
  );
}

function getDefaultSort(kind: LockResourceKind): SortType {
  if (kind === 'node') {
    return { fieldName: 'hostname', dir: 'ASC' };
  }
  return { fieldName: 'name', dir: 'ASC' };
}

function getFetchFuncForServerSidePaginating(
  ctx: Ctx,
  resourceKind: LockResourceKind
): (
  clusterId: string,
  params: UrlResourcesParams
) => Promise<ResourcesResponse<unknown>> {
  if (resourceKind === 'role') {
    return async (clusterId, params) => {
      const { items, startKey } = await ctx.resourceService.fetchRoles(params);
      return { agents: items, startKey };
    };
  }
  if (resourceKind === 'node') {
    return ctx.nodeService.fetchNodes;
  }

  if (resourceKind === 'windows_desktop') {
    return ctx.desktopService.fetchDesktops;
  }
}
