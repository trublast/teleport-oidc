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

import React from 'react';
import { MemoryRouter } from 'react-router';
import { render, screen } from 'design/utils/testing';

import { ContextProvider } from 'teleport/index';
import TeleportContext from 'teleport/teleportContext';
import cfg from 'teleport/config';

import { IntegrationTiles } from './IntegrationTiles';

test('render', async () => {
  render(
    <MemoryRouter>
      <IntegrationTiles />
    </MemoryRouter>
  );

  expect(screen.getByText(/amazon web services/i)).toBeInTheDocument();
  expect(screen.queryByText(/no permission/i)).not.toBeInTheDocument();
  expect(screen.getAllByTestId('svg')).toHaveLength(2);
  expect(screen.getAllByRole('link')).toHaveLength(2);

  const tile = screen.getByTestId('tile-aws-oidc');
  expect(tile).toBeEnabled();
  expect(tile.getAttribute('href')).toBeTruthy();
});

test('render disabled', async () => {
  render(
    <MemoryRouter>
      <IntegrationTiles
        hasIntegrationAccess={false}
        hasExternalAuditStorage={false}
      />
    </MemoryRouter>
  );

  expect(screen.getByText(/lacking permission/i)).toBeInTheDocument();
  expect(screen.queryByRole('link')).not.toBeInTheDocument();

  const tile = screen.getByTestId('tile-aws-oidc');
  expect(tile).not.toHaveAttribute('href');

  // The element has disabled attribute, but it's in the format `disabled=""`
  // so "toBeDisabled" interprets it as false.
  // eslint-disable-next-line jest-dom/prefer-enabled-disabled
  expect(tile).toHaveAttribute('disabled');
});

test('dont render External Audit Storage for enterprise unless it is cloud', async () => {
  cfg.isEnterprise = true;
  const ctx = new TeleportContext();
  ctx.isEnterprise = true;
  ctx.isCloud = false;

  render(
    <MemoryRouter>
      <ContextProvider ctx={ctx}>
        <IntegrationTiles />
      </ContextProvider>
    </MemoryRouter>
  );

  expect(
    screen.queryByText(/AWS External Audit Storage/i)
  ).not.toBeInTheDocument();
});
