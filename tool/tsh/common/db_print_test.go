/*
Copyright 2023 Gravitational, Inc.

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

package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	apidefaults "github.com/gravitational/teleport/api/defaults"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/services"
)

func Test_printDatabaseTable(t *testing.T) {
	t.Parallel()

	rows := []databaseTableRow{
		{
			Proxy:        "proxy",
			Cluster:      "cluster1",
			DisplayName:  "db1",
			Description:  "describe db1",
			Protocol:     "postgres",
			Type:         "self-hosted",
			URI:          "localhost:5432",
			AllowedUsers: "[*]",
			Labels:       "Env=dev",
			Connect:      "tsh db connect db1",
		},
		{
			Proxy:         "proxy",
			Cluster:       "cluster1",
			DisplayName:   "db2",
			Description:   "describe db2",
			Protocol:      "mysql",
			Type:          "self-hosted",
			URI:           "localhost:3306",
			AllowedUsers:  "[alice]",
			DatabaseRoles: "[readonly]",
			Labels:        "Env=prod",
		},
	}

	tests := []struct {
		name   string
		cfg    printDatabaseTableConfig
		expect string
	}{
		{
			name: "tsh db ls",
			cfg: printDatabaseTableConfig{
				rows:                rows,
				showProxyAndCluster: false,
				verbose:             false,
			},
			// os.Stdin.Fd() fails during go test, so width is defaulted to 80 for truncated table.
			expect: `Name Description  Allowed Users Labels   Connect             
---- ------------ ------------- -------- ------------------- 
db1  describe db1 [*]           Env=dev  tsh db connect d... 
db2  describe db2 [alice]       Env=prod                     

`,
		},
		{
			name: "tsh db ls --verbose",
			cfg: printDatabaseTableConfig{
				rows:                rows,
				showProxyAndCluster: false,
				verbose:             true,
			},
			expect: `Name Description  Protocol Type        URI            Allowed Users Database Roles Labels   Connect            
---- ------------ -------- ----------- -------------- ------------- -------------- -------- ------------------ 
db1  describe db1 postgres self-hosted localhost:5432 [*]                          Env=dev  tsh db connect db1 
db2  describe db2 mysql    self-hosted localhost:3306 [alice]       [readonly]     Env=prod                    

`,
		},
		{
			name: "tsh db ls --verbose --all",
			cfg: printDatabaseTableConfig{
				rows:                rows,
				showProxyAndCluster: true,
				verbose:             true,
			},
			expect: `Proxy Cluster  Name Description  Protocol Type        URI            Allowed Users Database Roles Labels   Connect            
----- -------- ---- ------------ -------- ----------- -------------- ------------- -------------- -------- ------------------ 
proxy cluster1 db1  describe db1 postgres self-hosted localhost:5432 [*]                          Env=dev  tsh db connect db1 
proxy cluster1 db2  describe db2 mysql    self-hosted localhost:3306 [alice]       [readonly]     Env=prod                    

`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sb strings.Builder

			cfg := test.cfg
			cfg.writer = &sb

			printDatabaseTable(cfg)
			require.Equal(t, test.expect, sb.String())
		})
	}
}

func Test_formatDatabaseRolesForDB(t *testing.T) {
	t.Parallel()

	db, err := types.NewDatabaseV3(types.Metadata{
		Name: "db",
	}, types.DatabaseSpecV3{
		Protocol: "postgres",
		URI:      "localhost:5432",
	})
	require.NoError(t, err)

	dbWithAutoUser, err := types.NewDatabaseV3(types.Metadata{
		Name:   "dbWithAutoUser",
		Labels: map[string]string{"env": "prod"},
	}, types.DatabaseSpecV3{
		Protocol: "postgres",
		URI:      "localhost:5432",
		AdminUser: &types.DatabaseAdminUser{
			Name: "teleport-admin",
		},
	})
	require.NoError(t, err)

	roleAutoUser := &types.RoleV6{
		Metadata: types.Metadata{Name: "auto-user", Namespace: apidefaults.Namespace},
		Spec: types.RoleSpecV6{
			Options: types.RoleOptions{
				CreateDatabaseUserMode: types.CreateDatabaseUserMode_DB_USER_MODE_KEEP,
			},
			Allow: types.RoleConditions{
				Namespaces:     []string{apidefaults.Namespace},
				DatabaseLabels: types.Labels{"env": []string{"prod"}},
				DatabaseRoles:  []string{"roleA", "roleB"},
				DatabaseNames:  []string{"*"},
				DatabaseUsers:  []string{types.Wildcard},
			},
		},
	}

	tests := []struct {
		name          string
		database      types.Database
		accessChecker services.AccessChecker
		expect        string
	}{
		{
			name:     "nil accessChecker",
			database: dbWithAutoUser,
			expect:   "(unknown)",
		},
		{
			name:     "roles",
			database: dbWithAutoUser,
			accessChecker: services.NewAccessCheckerWithRoleSet(&services.AccessInfo{
				Username: "alice",
			}, "clustername", services.RoleSet{roleAutoUser}),
			expect: "[roleA roleB]",
		},
		{
			name:     "db without admin user",
			database: db,
			accessChecker: services.NewAccessCheckerWithRoleSet(&services.AccessInfo{
				Username: "alice",
			}, "clustername", services.RoleSet{roleAutoUser}),
			expect: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expect, formatDatabaseRolesForDB(test.database, test.accessChecker))
		})
	}
}
