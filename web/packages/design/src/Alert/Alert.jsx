/*
Copyright 2019 Gravitational, Inc.

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

import React from 'react';
import styled from 'styled-components';
import PropTypes from 'prop-types';

import { space, color, width } from 'design/system';
import { fade } from 'design/theme/utils/colorManipulator';

const kind = props => {
  const { kind, theme } = props;
  switch (kind) {
    case 'danger':
      return {
        background: theme.colors.error.main,
        color: theme.colors.buttons.warning.text,
      };
    case 'info':
      return {
        background: theme.colors.info,
        color: theme.colors.text.primaryInverse,
      };
    case 'warning':
      return {
        background: theme.colors.warning.main,
        color: theme.colors.text.primaryInverse,
      };
    case 'success':
      return {
        background: theme.colors.success.main,
        color: theme.colors.text.primaryInverse,
      };
    case 'outline-danger':
      return {
        background: fade(theme.colors.error.main, 0.1),
        border: `${theme.borders[2]} ${theme.colors.error.main}`,
        borderRadius: `${theme.radii[3]}px`,
        boxShadow: 'none',
        justifyContent: 'normal',
      };
    case 'outline-info':
      return {
        background: fade(theme.colors.accent.main, 0.1),
        border: `${theme.borders[2]} ${theme.colors.accent.main}`,
        borderRadius: `${theme.radii[3]}px`,
        boxShadow: 'none',
        justifyContent: 'normal',
      };
    case 'outline-warn':
      return {
        background: fade(theme.colors.warning.main, 0.1),
        border: `${theme.borders[2]} ${theme.colors.warning.main}`,
        borderRadius: `${theme.radii[3]}px`,
        boxShadow: 'none',
        justifyContent: 'normal',
      };
    default:
      return {
        background: theme.colors.error.main,
        color: theme.colors.text.primaryInverse,
      };
  }
};

const Alert = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: ${p => p.theme.radii[1]}px;
  box-sizing: border-box;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.24);
  margin: 0 0 24px 0;
  min-height: 40px;
  padding: 8px 16px;
  overflow: auto;
  word-break: break-word;
  line-height: 1.5;
  ${space}
  ${kind}
  ${width}

  a {
    color: ${({ theme }) => theme.colors.light};
  }
`;

Alert.propTypes = {
  kind: PropTypes.oneOf([
    'danger',
    'info',
    'warning',
    'success',
    'outline-info',
    'outline-warn',
  ]),
  ...color.propTypes,
  ...space.propTypes,
  ...width.propTypes,
};

Alert.defaultProps = {
  kind: 'danger',
};

Alert.displayName = 'Alert';

export default Alert;
export const Danger = props => <Alert kind="danger" {...props} />;
export const Info = props => <Alert kind="info" {...props} />;
export const Warning = props => <Alert kind="warning" {...props} />;
export const Success = props => <Alert kind="success" {...props} />;
export const OutlineInfo = props => <Alert kind="outline-info" {...props} />;
export const OutlineWarn = props => <Alert kind="outline-warn" {...props} />;
