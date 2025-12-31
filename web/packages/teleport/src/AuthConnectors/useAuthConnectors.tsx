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

import { useEffect, useState } from 'react';
import useAttempt from 'shared/hooks/useAttemptNext';

import { Resource, KindAuthConnectors } from 'teleport/services/resources';
import useTeleport from 'teleport/useTeleport';

export default function useAuthConnectors() {
  const ctx = useTeleport();
  const [items, setItems] = useState<Resource<KindAuthConnectors>[]>([]);
  const { attempt, run } = useAttempt('processing');

  function fetchData() {
    return Promise.all([
      ctx.resourceService.fetchGithubConnectors(),
      ctx.resourceService.fetchOidcConnectors(),
      ctx.resourceService.fetchSamlConnectors(),
    ]).then(([github, oidc, saml]) => {
      setItems([...github, ...oidc, ...saml]);
    });
  }

  function save(name: string, yaml: string, isNew: boolean, kind?: KindAuthConnectors) {
    // Extract kind from yaml if not provided
    if (!kind) {
      const kindMatch = yaml.match(/^kind:\s*(\w+)/m);
      if (kindMatch && ['github', 'oidc', 'saml'].includes(kindMatch[1])) {
        kind = kindMatch[1] as KindAuthConnectors;
      } else {
        // Default to github for backward compatibility
        kind = 'github';
      }
    }

    if (isNew) {
      if (kind === 'oidc') {
        return ctx.resourceService.createOidcConnector(yaml).then(fetchData);
      } else if (kind === 'saml') {
        return ctx.resourceService.createSamlConnector(yaml).then(fetchData);
      } else {
        return ctx.resourceService.createGithubConnector(yaml).then(fetchData);
      }
    } else {
      if (kind === 'oidc') {
        return ctx.resourceService.updateOidcConnector(name, yaml).then(fetchData);
      } else if (kind === 'saml') {
        return ctx.resourceService.updateSamlConnector(name, yaml).then(fetchData);
      } else {
        return ctx.resourceService.updateGithubConnector(name, yaml).then(fetchData);
      }
    }
  }

  function remove(name: string, kind?: KindAuthConnectors) {
    // Try to find the item to determine its kind
    if (!kind) {
      const item = items.find(item => item.name === name);
      if (item) {
        kind = item.kind as KindAuthConnectors;
      } else {
        // Default to github for backward compatibility
        kind = 'github';
      }
    }

    if (kind === 'oidc') {
      return ctx.resourceService.deleteOidcConnector(name).then(fetchData);
    } else if (kind === 'saml') {
      return ctx.resourceService.deleteSamlConnector(name).then(fetchData);
    } else {
      return ctx.resourceService.deleteGithubConnector(name).then(fetchData);
    }
  }

  useEffect(() => {
    run(() => fetchData());
  }, []);

  return {
    items,
    attempt,
    save,
    remove,
  };
}

export type State = ReturnType<typeof useAuthConnectors>;
