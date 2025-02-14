/*
Copyright 2020 Gravitational, Inc.

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

const path = require('path');
const fs = require('fs');
const configFactory = require('@gravitational/build/webpack/webpack.base');

// include open source stories
const stories = ['../packages/**/*.story.@(js|jsx|ts|tsx)'];

const tsconfigPath = path.join(__dirname, '../../tsconfig.json');

const enterpriseTeleportExists = fs.existsSync(
  path.join(__dirname, '/../../e/web')
);

// include enterprise stories if available (**/* pattern ignores dot dir names)
if (enterpriseTeleportExists) {
  stories.unshift('../../e/web/**/*.story.@(js|jsx|ts|tsx)');
}

module.exports = {
  core: {
    builder: 'webpack5',
  },
  reactOptions: {
    fastRefresh: true,
  },
  typescript: {
    reactDocgen: false,
  },
  addons: ['@storybook/addon-toolbars'],
  stories,
  webpackFinal: async (storybookConfig, { configType }) => {
    // configType has a value of 'DEVELOPMENT' or 'PRODUCTION'
    // You can change the configuration based on that.
    // 'PRODUCTION' is used when building the static version of storybook.
    storybookConfig.devtool = false;
    storybookConfig.resolve = {
      ...storybookConfig.resolve,
      ...configFactory.createDefaultConfig().resolve,
    };

    // Access Graph requires a separate repo to be cloned. At the moment, only the Vite config is
    // configured to resolve access-graph. However, Storybook uses Webpack and since our usual
    // Webpack config doesn't need to know about access-graph, we manually to manually configure
    // Storybook's Webpack here to resolve access-graph to the special mock.
    //
    // See https://github.com/gravitational/teleport.e/issues/2675.
    storybookConfig.resolve.alias['access-graph'] = path.join(
      __dirname,
      'mocks',
      'AccessGraph.tsx'
    );

    if (!enterpriseTeleportExists) {
      delete storybookConfig.resolve.alias['e-teleport'];
      // Unlike e-teleport, e-teleterm cannot be removed from aliases because code in OSS teleterm
      // depends directly on e-teleterm, see https://github.com/gravitational/teleport/issues/17706.
      //
      // Instead of removing e-teleterm, we have to mock individual files on a case-by-case basis.
      //
      // TODO(ravicious): Remove e-teleterm alias once #17706 gets addressed.
      storybookConfig.resolve.alias['e-teleterm'] = path.join(
        __dirname,
        'mocks',
        'e-teleterm'
      );
    }

    storybookConfig.optimization = {
      splitChunks: {
        cacheGroups: {
          stories: {
            maxSize: 500000, // 500kb
            chunks: 'all',
            name: 'stories',
            test: /packages/,
          },
        },
      },
    };

    storybookConfig.module.rules.push({
      resourceQuery: /raw/,
      type: 'asset/source',
    });

    storybookConfig.module.rules.push({
      test: /\.(ts|tsx)$/,
      use: [
        {
          loader: require.resolve('babel-loader'),
        },
        {
          loader: require.resolve('ts-loader'),
          options: {
            onlyCompileBundledFiles: true,
            configFile: tsconfigPath,
            transpileOnly: configType === 'DEVELOPMENT',
            compilerOptions: {
              jsx: 'preserve',
            },
          },
        },
      ],
    });

    return storybookConfig;
  },
};
