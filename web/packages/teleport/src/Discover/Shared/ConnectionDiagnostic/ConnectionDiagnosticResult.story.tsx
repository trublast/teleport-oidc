/**
 * Copyright 2022 Gravitational, Inc.
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

import { ConnectionDiagnosticResult } from './ConnectionDiagnosticResult';

import type { Props } from './ConnectionDiagnosticResult';
import type { ConnectionDiagnosticTrace } from 'teleport/services/agents';

export default {
  title: 'Teleport/Discover/Shared/ConnectionDiagnostic',
};

export const Init = () => (
  <ConnectionDiagnosticResult {...props} diagnosis={null} />
);

export const NumberAndDescriptionOnSameLine = () => (
  <ConnectionDiagnosticResult
    {...props}
    numberAndDescriptionOnSameLine
    diagnosis={null}
  />
);

export const DiagnosisSuccess = () => (
  <ConnectionDiagnosticResult
    {...props}
    attempt={{ status: 'success' }}
    diagnosis={diagnosisSuccess}
  />
);

export const DiagnosisFailed = () => (
  <ConnectionDiagnosticResult
    {...props}
    attempt={{ status: 'success' }}
    diagnosis={diagnosisFailed}
  />
);

export const DiagnosisLoading = () => (
  <ConnectionDiagnosticResult {...props} attempt={{ status: 'processing' }} />
);

export const NoAccess = () => (
  <ConnectionDiagnosticResult {...props} canTestConnection={false} />
);

export const Error = () => (
  <ConnectionDiagnosticResult
    {...props}
    attempt={{ status: 'failed', statusText: 'some error message' }}
  />
);

const diagnosisSuccess = {
  id: 'id',
  labels: [],
  success: true,
  message: 'some diagnosis message',
  traces: [
    {
      traceType: 'rbac node',
      status: 'success',
      details: 'Resource exists.',
    },
    {
      traceType: 'network connectivity',
      status: 'success',
      details: 'Host is alive and reachable.',
    },
    {
      traceType: 'rbac principal',
      status: 'success',
      details: 'Successfully authenticated.',
    },
    {
      traceType: 'node ssh server',
      status: 'success',
      details: 'Established an SSH connection.',
    },
    {
      traceType: 'node ssh session',
      status: 'success',
      details: 'Created an SSH session.',
    },
    {
      traceType: 'node principal',
      status: 'success',
      details: 'User exists message.',
    },
  ] as ConnectionDiagnosticTrace[],
};

const diagnosisFailed = {
  id: 'id',
  labels: [],
  success: false,
  message: 'some diagnosis message',
  traces: [
    {
      traceType: 'rbac node',
      status: 'success',
      details: 'Resource exists.',
    },
    {
      traceType: 'network connectivity',
      status: 'success',
      details: 'Host is alive and reachable.',
    },
    {
      traceType: 'rbac principal',
      status: 'failed',
      details:
        'Long explanation with line breaks as to why rbac principal check failed.\nThis sentence should start after a line break. Morbi sollicitudin nisi at sem iaculis porta. Cras fringilla, nunc a rhoncus convallis, mauris lacus cursus nibh, quis posuere sem sapien vitae turpis. Integer vitae nulla et nulla varius placerat sed quis magna.\nNunc rhoncus justo sit amet lorem euismod condimentum.',
      error:
        'And here is a multiline error message.\nThis is another sentence which should start after a line break. Maecenas gravida, mauris in tincidunt rutrum, odio arcu ullamcorper elit, quis pretium sapien dui quis nibh. Pellentesque suscipit at erat in tempus. Morbi id neque vel turpis egestas eleifend. Phasellus consectetur commodo luctus. Fusce tempor eleifend tempus. Aenean posuere consequat nisl.\nAnother line break, oh my.',
    },
    {
      traceType: 'node ssh session',
      status: 'failed',
      details: 'Why node ssh session might have failed',
      error: 'Some extra error log 2',
    },
  ] as ConnectionDiagnosticTrace[],
};

const props: Props = {
  attempt: { status: '' },
  diagnosis: null,
  canTestConnection: true,
  testConnection: () => null,
  stepNumber: 2,
  stepDescription: 'Verify that your example database is accessible',
};
