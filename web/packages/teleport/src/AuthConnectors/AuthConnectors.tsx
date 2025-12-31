/*
Copyright 2020-2021 Gravitational, Inc.
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

import React, { useState } from 'react';
import { Alert, Box, Flex, Indicator, Link, Text } from 'design';

import { FeatureBox, FeatureHeaderTitle } from 'teleport/components/Layout';
import ResourceEditor from 'teleport/components/ResourceEditor';
import useResources from 'teleport/components/useResources';

import {
  DesktopDescription,
  MobileDescription,
  ResponsiveAddButton,
  ResponsiveFeatureHeader,
} from 'teleport/AuthConnectors/styles/AuthConnectors.styles';

import EmptyList from './EmptyList';
import ConnectorList from './ConnectorList';
import DeleteConnectorDialog from './DeleteConnectorDialog';
import SelectConnectorTypeDialog from './SelectConnectorTypeDialog';
import useAuthConnectors, { State } from './useAuthConnectors';
import templates from './templates';
import { KindAuthConnectors } from 'teleport/services/resources';

export default function Container() {
  const state = useAuthConnectors();
  return <AuthConnectors {...state} />;
}

export function AuthConnectors(props: State) {
  const { attempt, items, remove, save } = props;
  const isEmpty = items.length === 0;
  const resources = useResources(items, templates);
  const [showSelectDialog, setShowSelectDialog] = useState(false);

  const getConnectorTypeName = (kind: string) => {
    switch (kind) {
      case 'oidc':
        return 'OIDC';
      case 'saml':
        return 'SAML';
      case 'github':
      default:
        return 'GitHub';
    }
  };

  const connectorKind = resources.item?.kind || 'github';
  const connectorTypeName = getConnectorTypeName(connectorKind);
  const title =
    resources.status === 'creating'
      ? `Creating a new ${connectorTypeName.toLowerCase()} connector`
      : `Editing ${connectorTypeName.toLowerCase()} connector`;
  const description =
    'Auth connectors allow Teleport to authenticate users via an external identity source such as Okta, Microsoft Entra ID, GitHub, etc. This authentication method is commonly known as single sign-on (SSO).';

  function handleOnSave(content: string) {
    const name = resources.item.name;
    const isNew = resources.status === 'creating';
    const kind = resources.item.kind as 'github' | 'oidc' | 'saml';
    return save(name, content, isNew, kind);
  }

  return (
    <FeatureBox>
      <ResponsiveFeatureHeader>
        <FeatureHeaderTitle>Auth Connectors</FeatureHeaderTitle>
        <MobileDescription typography="subtitle1">
          {description}
        </MobileDescription>
        <ResponsiveAddButton onClick={() => setShowSelectDialog(true)}>
          New Connector
        </ResponsiveAddButton>
      </ResponsiveFeatureHeader>
      {attempt.status === 'failed' && <Alert children={attempt.statusText} />}
      {attempt.status === 'processing' && (
        <Box textAlign="center" m={10}>
          <Indicator />
        </Box>
      )}
      {attempt.status === 'success' && (
        <Flex alignItems="start">
          {isEmpty && (
            <Flex width="100%" justifyContent="center">
              <EmptyList onCreate={resources.create} />
            </Flex>
          )}
          <>
            <ConnectorList
              items={items}
              onEdit={resources.edit}
              onDelete={resources.remove}
            />
            <DesktopDescription>
              <Text typography="h6" mb={3} caps>
                Auth Connectors
              </Text>
              <Text typography="subtitle1" mb={3}>
                {description}
              </Text>
              <Text typography="subtitle1" mb={2}>
                Please{' '}
                <Link
                  color="text.main"
                  // This URL is the OSS documentation for auth connectors
                  href="https://goteleport.com/docs/setup/admin/github-sso/"
                  target="_blank"
                >
                  view our documentation
                </Link>{' '}
                on how to configure auth connectors.
              </Text>
            </DesktopDescription>
          </>
        </Flex>
      )}
      {(resources.status === 'creating' || resources.status === 'editing') && (
        <ResourceEditor
          title={title}
          onSave={handleOnSave}
          text={resources.item.content}
          name={resources.item.name}
          isNew={resources.status === 'creating'}
          onClose={resources.disregard}
        />
      )}
      {resources.status === 'removing' && (
        <DeleteConnectorDialog
          name={resources.item.name}
          onClose={resources.disregard}
          onDelete={() => {
            const kind = resources.item.kind as 'github' | 'oidc' | 'saml';
            return remove(resources.item.name, kind);
          }}
        />
      )}
      {showSelectDialog && (
        <SelectConnectorTypeDialog
          onSelect={(kind: KindAuthConnectors) => {
            setShowSelectDialog(false);
            resources.create(kind);
          }}
          onClose={() => setShowSelectDialog(false)}
        />
      )}
    </FeatureBox>
  );
}
