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

import { Flex, Text, ButtonSecondary } from 'design';
import { AuthProviderType } from 'shared/services';
import Dialog, {
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from 'design/Dialog';

import { ConnectorBox } from 'teleport/AuthConnectors/styles/ConnectorBox.styles';
import getSsoIcon from 'teleport/AuthConnectors/ssoIcons/getSsoIcon';
import { KindAuthConnectors } from 'teleport/services/resources';

export default function SelectConnectorTypeDialog({
  onSelect,
  onClose,
}: Props) {
  return (
    <Dialog
      dialogCss={() => ({ maxWidth: '800px', width: '100%' })}
      disableEscapeKeyDown={false}
      onClose={onClose}
      open={true}
    >
      <DialogHeader>
        <DialogTitle>Select Connector Type</DialogTitle>
      </DialogHeader>
      <DialogContent>
        <Text typography="paragraph" mb={4} textAlign="center">
          Choose the type of authentication connector you want to create
        </Text>
        <Flex flexWrap="wrap" justifyContent="center" gap={3}>
          {renderConnectorItem('github', () => onSelect('github'))}
          {renderConnectorItem('oidc',   () => onSelect('oidc'))}
          {renderConnectorItem('saml',   () => onSelect('saml'))}
        </Flex>
      </DialogContent>
      <DialogFooter>
        <ButtonSecondary onClick={onClose}>Cancel</ButtonSecondary>
      </DialogFooter>
    </Dialog>
  );
}

function renderConnectorItem(kind: AuthProviderType, onClick: () => void) {
  const { desc, SsoIcon, info } = getSsoIcon(kind);
  return (
    <ConnectorBox as="button" onClick={onClick}>
      <Flex width="100%">
        <SsoIcon
          fontSize="50px"
          style={{
            left: 0,
            fontSize: '72px',
          }}
        />
      </Flex>

      <Text typography="body2" mt={4} fontSize={4} color="text.primary" bold>
        {desc}
      </Text>
      {info && (
        <Text mt={2} color="text.primary" transform="none">
          {info}
        </Text>
      )}
    </ConnectorBox>
  );
}

type Props = {
  onSelect: (kind: KindAuthConnectors) => void;
  onClose: () => void;
};

