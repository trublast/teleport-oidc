/*
Copyright 2015-2019 Gravitational, Inc.

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

package services

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/go-oidc/jose"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	"gopkg.in/check.v1"

	"github.com/gravitational/teleport/lib/defaults"
)

type UserSuite struct {
}

var _ = check.Suite(&UserSuite{})

func (s *UserSuite) TestTraits(c *check.C) {
	var tests = []struct {
		traitName string
	}{
		// Windows trait names are URLs.
		{
			traitName: "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsaccountname",
		},
		// Simple strings are the most common trait names.
		{
			traitName: "user-groups",
		},
	}

	for _, tt := range tests {
		user := &UserV2{
			Kind:    KindUser,
			Version: V2,
			Metadata: Metadata{
				Name:      "foo",
				Namespace: defaults.Namespace,
			},
			Spec: UserSpecV2{
				Traits: map[string][]string{
					tt.traitName: {"foo"},
				},
			},
		}

		data, err := json.Marshal(user)
		c.Assert(err, check.IsNil)

		_, err = UnmarshalUser(data)
		c.Assert(err, check.IsNil)
	}
}

func (s *UserSuite) TestOIDCMapping(c *check.C) {
	type input struct {
		comment       string
		claims        jose.Claims
		expectedRoles []string
		warnings      []string
	}
	testCases := []struct {
		comment  string
		mappings []ClaimMapping
		inputs   []input
	}{
		{
			comment: "no mappings",
			inputs: []input{
				{
					claims:        jose.Claims{"a": "b"},
					expectedRoles: nil,
				},
			},
		},
		{
			comment: "simple mappings",
			mappings: []ClaimMapping{
				{Claim: "role", Value: "admin", Roles: []string{"admin", "bob"}},
				{Claim: "role", Value: "user", Roles: []string{"user"}},
			},
			inputs: []input{
				{
					comment:       "no match",
					claims:        jose.Claims{"a": "b"},
					expectedRoles: nil,
				},
				{
					comment:       "no value match",
					claims:        jose.Claims{"role": "b"},
					expectedRoles: nil,
				},
				{
					comment:       "direct admin value match",
					claims:        jose.Claims{"role": "admin"},
					expectedRoles: []string{"admin", "bob"},
				},
				{
					comment:       "direct user value match",
					claims:        jose.Claims{"role": "user"},
					expectedRoles: []string{"user"},
				},
				{
					comment:       "direct user value match with array",
					claims:        jose.Claims{"role": []string{"user"}},
					expectedRoles: []string{"user"},
				},
			},
		},
		{
			comment: "regexp mappings match",
			mappings: []ClaimMapping{
				{Claim: "role", Value: "^admin-(.*)$", Roles: []string{"role-$1", "bob"}},
			},
			inputs: []input{
				{
					comment:       "no match",
					claims:        jose.Claims{"a": "b"},
					expectedRoles: nil,
				},
				{
					comment:       "no match - subprefix",
					claims:        jose.Claims{"role": "adminz"},
					expectedRoles: nil,
				},
				{
					comment:       "value with capture match",
					claims:        jose.Claims{"role": "admin-hello"},
					expectedRoles: []string{"role-hello", "bob"},
				},
				{
					comment:       "multiple value with capture match, deduplication",
					claims:        jose.Claims{"role": []string{"admin-hello", "admin-ola"}},
					expectedRoles: []string{"role-hello", "bob", "role-ola"},
				},
				{
					comment:       "first matches, second does not",
					claims:        jose.Claims{"role": []string{"hello", "admin-ola"}},
					expectedRoles: []string{"role-ola", "bob"},
				},
			},
		},
		{
			comment: "empty expands are skipped",
			mappings: []ClaimMapping{
				{Claim: "role", Value: "^admin-(.*)$", Roles: []string{"$2", "bob"}},
			},
			inputs: []input{
				{
					comment:       "value with capture match",
					claims:        jose.Claims{"role": "admin-hello"},
					expectedRoles: []string{"bob"},
				},
			},
		},
		{
			comment: "glob wildcard match",
			mappings: []ClaimMapping{
				{Claim: "role", Value: "*", Roles: []string{"admin"}},
			},
			inputs: []input{
				{
					comment:       "empty value match",
					claims:        jose.Claims{"role": ""},
					expectedRoles: []string{"admin"},
				},
				{
					comment:       "any value match",
					claims:        jose.Claims{"role": "zz"},
					expectedRoles: []string{"admin"},
				},
			},
		},
		{
			comment: "Whitespace/dashes",
			mappings: []ClaimMapping{
				{Claim: "groups", Value: "DemoCorp - Backend Engineers", Roles: []string{"backend"}},
				{Claim: "groups", Value: "DemoCorp - SRE Managers", Roles: []string{"approver"}},
				{Claim: "groups", Value: "DemoCorp - SRE", Roles: []string{"approver"}},
				{Claim: "groups", Value: "DemoCorp Infrastructure", Roles: []string{"approver", "backend"}},
			},
			inputs: []input{
				{
					comment: "Matches multiple groups",
					claims: jose.Claims{
						"groups": []string{"DemoCorp - Backend Engineers", "DemoCorp Infrastructure"},
					},
					expectedRoles: []string{"backend", "approver"},
				},
				{
					comment: "Matches one group",
					claims: jose.Claims{
						"groups": []string{"DemoCorp - SRE"},
					},
					expectedRoles: []string{"approver"},
				},
				{
					comment: "Matches one group with multiple roles",
					claims: jose.Claims{
						"groups": []string{"DemoCorp Infrastructure"},
					},
					expectedRoles: []string{"approver", "backend"},
				},
				{
					comment: "No match only due to case-sensitivity",
					claims: jose.Claims{
						"groups": []string{"Democorp - SRE"},
					},
					expectedRoles: []string(nil),
					warnings:      []string{`trait "Democorp - SRE" matches value "DemoCorp - SRE" case-insensitively and would have yielded "approver" role`},
				},
			},
		},
	}

	for i, testCase := range testCases {
		conn := OIDCConnectorV2{
			Spec: OIDCConnectorSpecV2{
				ClaimsToRoles: testCase.mappings,
			},
		}
		for _, input := range testCase.inputs {
			comment := check.Commentf("OIDC Test case %v %q, input %q", i, testCase.comment, input.comment)
			_, outRoles := TraitsToRoles(conn.GetTraitMappings(), OIDCClaimsToTraits(input.claims))
			c.Assert(outRoles, check.DeepEquals, input.expectedRoles, comment)
		}

		samlConn := SAMLConnectorV2{
			Spec: SAMLConnectorSpecV2{
				AttributesToRoles: claimMappingsToAttributeMappings(testCase.mappings),
			},
		}
		for _, input := range testCase.inputs {
			comment := check.Commentf("SAML Test case %v %v, input %#v", i, testCase.comment, input)
			warnings, outRoles := TraitsToRoles(samlConn.GetTraitMappings(), SAMLAssertionsToTraits(claimsToAttributes(input.claims)))
			c.Assert(outRoles, check.DeepEquals, input.expectedRoles, comment)
			c.Assert(warnings, check.DeepEquals, input.warnings, comment)
		}
	}
}

// claimMappingsToAttributeMappings converts oidc claim mappings to
// attribute mappings, used in tests
func claimMappingsToAttributeMappings(in []ClaimMapping) []AttributeMapping {
	var out []AttributeMapping
	for _, m := range in {
		out = append(out, AttributeMapping{
			Name:  m.Claim,
			Value: m.Value,
			Roles: append([]string{}, m.Roles...),
		})
	}
	return out
}

// claimsToAttributes maps jose.Claims type to attributes for testing
func claimsToAttributes(claims jose.Claims) saml2.AssertionInfo {
	info := saml2.AssertionInfo{
		Values: make(map[string]types.Attribute),
	}
	for claim, values := range claims {
		attr := types.Attribute{
			Name: claim,
		}
		switch val := values.(type) {
		case string:
			attr.Values = []types.AttributeValue{{Value: val}}
		case []string:
			for _, v := range val {
				attr.Values = append(attr.Values, types.AttributeValue{Value: v})
			}
		default:
			panic(fmt.Sprintf("unsupported type %T", val))
		}
		info.Values[claim] = attr
	}
	return info
}
