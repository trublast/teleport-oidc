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

package externalauditstorage_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	athenatypes "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	gluetypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	eastypes "github.com/gravitational/teleport/api/types/externalauditstorage"
	"github.com/gravitational/teleport/lib/integrations/externalauditstorage"
)

func TestBootstrapInfra(t *testing.T) {
	t.Parallel()
	goodSpec := &eastypes.ExternalAuditStorageSpec{
		SessionRecordingsURI:   "s3://long-term-storage-bucket/session",
		AuditEventsLongTermURI: "s3://long-term-storage-bucket/events",
		AthenaResultsURI:       "s3://transient-storage-bucket/query_results",
		AthenaWorkgroup:        "teleport-workgroup",
		GlueDatabase:           "teleport-database",
		GlueTable:              "audit-events",
	}
	tt := []struct {
		desc                     string
		region                   string
		spec                     *eastypes.ExternalAuditStorageSpec
		errWanted                string
		locationConstraintWanted s3types.BucketLocationConstraint
	}{
		{
			desc:      "nil input",
			region:    "us-west-2",
			errWanted: "param Spec required",
		},
		{
			desc:      "empty region input",
			errWanted: "param Region required",
			spec:      goodSpec,
		},
		{
			desc:                     "us-west-2",
			region:                   "us-west-2",
			spec:                     goodSpec,
			locationConstraintWanted: s3types.BucketLocationConstraintUsWest2,
		},
		{
			desc:   "us-east-1",
			region: "us-east-1",
			spec:   goodSpec,
			// No location constraint wanted for us-east-1 because it is the
			// default and AWS has decided, in all their infinite wisdom, that
			// the CreateBucket API should fail if you explicitly pass the
			// default location constraint.
		},
		{
			desc:                     "eu-central-1",
			region:                   "eu-central-1",
			spec:                     goodSpec,
			locationConstraintWanted: s3types.BucketLocationConstraintEuCentral1,
		},
		{
			desc:                     "ap-south-1",
			region:                   "ap-south-1",
			spec:                     goodSpec,
			locationConstraintWanted: s3types.BucketLocationConstraintApSouth1,
		},
		{
			desc:                     "ap-southeast-1",
			region:                   "ap-southeast-1",
			spec:                     goodSpec,
			locationConstraintWanted: s3types.BucketLocationConstraintApSoutheast1,
		},
		{
			desc:                     "sa-east-1",
			region:                   "sa-east-1",
			spec:                     goodSpec,
			locationConstraintWanted: s3types.BucketLocationConstraintSaEast1,
		},
		{
			desc:      "invalid input transient and long-term share same bucket name",
			errWanted: "athena results bucket URI must not match audit events or session bucket URI",
			region:    "us-west-2",
			spec: &eastypes.ExternalAuditStorageSpec{
				SessionRecordingsURI:   "s3://long-term-storage-bucket/session",
				AuditEventsLongTermURI: "s3://long-term-storage-bucket/events",
				AthenaResultsURI:       "s3://long-term-storage-bucket/query_results",
				AthenaWorkgroup:        "teleport-workgroup",
				GlueDatabase:           "teleport-database",
				GlueTable:              "audit-events",
			},
		},
		{
			desc:      "invalid input audit events and session recordings have different URIs",
			errWanted: "audit events bucket URI must match session bucket URI",
			region:    "us-west-2",
			spec: &eastypes.ExternalAuditStorageSpec{
				SessionRecordingsURI:   "s3://long-term-storage-bucket-sessions/session",
				AuditEventsLongTermURI: "s3://long-term-storage-bucket-events/events",
				AthenaResultsURI:       "s3://transient-storage-bucket/query_results",
				AthenaWorkgroup:        "teleport-workgroup",
				GlueDatabase:           "teleport-database",
				GlueTable:              "audit-events",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			testCtx := context.Background()
			s3Clt := &mockBootstrapS3Client{buckets: make(map[string]bucket)}
			athenaClt := &mockBootstrapAthenaClient{}
			glueClt := &mockBootstrapGlueClient{}
			err := externalauditstorage.BootstrapInfra(testCtx, externalauditstorage.BootstrapInfraParams{
				Athena: athenaClt,
				Glue:   glueClt,
				S3:     s3Clt,
				Spec:   tc.spec,
				Region: tc.region,
			})
			if tc.errWanted != "" {
				require.ErrorContainsf(t, err, tc.errWanted, "the error returned did not contain: %s", tc.errWanted)
				return
			} else {
				require.NoError(t, err, "an unexpected error occurred in BootstrapInfra")
			}

			ltsBucket, err := url.Parse(tc.spec.AuditEventsLongTermURI)
			require.NoError(t, err)

			transientBucket, err := url.Parse(tc.spec.AthenaResultsURI)
			require.NoError(t, err)

			if b, ok := s3Clt.buckets[ltsBucket.Host]; ok {
				assert.Equal(t, tc.locationConstraintWanted, b.locationConstraint)
			} else {
				t.Fatalf("Long-term bucket: %s not created by bootstrap infra", ltsBucket.Host)
			}

			if b, ok := s3Clt.buckets[transientBucket.Host]; ok {
				assert.Equal(t, tc.locationConstraintWanted, b.locationConstraint)
			} else {
				t.Fatalf("Transient bucket: %s not created by bootstrap infra", transientBucket.Host)
			}

			assert.Equal(t, tc.spec.GlueDatabase, glueClt.database)
			assert.Equal(t, tc.spec.GlueTable, glueClt.table)
			assert.Equal(t, tc.spec.AthenaWorkgroup, athenaClt.workgroup)

			// Re-run bootstrap
			assert.NoError(t, externalauditstorage.BootstrapInfra(testCtx, externalauditstorage.BootstrapInfraParams{
				Athena: athenaClt,
				Glue:   glueClt,
				S3:     s3Clt,
				Spec:   tc.spec,
				Region: tc.region,
			}))
		})
	}
}

type mockBootstrapS3Client struct {
	buckets map[string]bucket
}

type bucket struct {
	locationConstraint s3types.BucketLocationConstraint
}

type mockBootstrapAthenaClient struct {
	workgroup string
}

type mockBootstrapGlueClient struct {
	table    string
	database string
}

func (c *mockBootstrapS3Client) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	if _, ok := c.buckets[*params.Bucket]; ok {
		// bucket already exists
		return nil, &s3types.BucketAlreadyExists{Message: aws.String("The bucket already exists")}
	}

	var locationConstraint s3types.BucketLocationConstraint
	if params.CreateBucketConfiguration != nil {
		locationConstraint = params.CreateBucketConfiguration.LocationConstraint
	}
	c.buckets[*params.Bucket] = bucket{
		locationConstraint: locationConstraint,
	}

	return &s3.CreateBucketOutput{}, nil
}

func (c *mockBootstrapS3Client) PutObjectLockConfiguration(ctx context.Context, params *s3.PutObjectLockConfigurationInput, optFns ...func(*s3.Options)) (*s3.PutObjectLockConfigurationOutput, error) {
	if _, ok := c.buckets[*params.Bucket]; !ok {
		// bucket doesn't exist return no such bucket error
		return nil, &s3types.NoSuchBucket{Message: aws.String("The bucket doesn't exist")}
	}

	return &s3.PutObjectLockConfigurationOutput{}, nil
}
func (c *mockBootstrapS3Client) PutBucketVersioning(ctx context.Context, params *s3.PutBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
	if _, ok := c.buckets[*params.Bucket]; !ok {
		// bucket doesn't exist return no such bucket error
		return nil, &s3types.NoSuchBucket{Message: aws.String("The bucket doesn't exist")}
	}

	return &s3.PutBucketVersioningOutput{}, nil
}

func (c *mockBootstrapS3Client) PutBucketLifecycleConfiguration(ctx context.Context, params *s3.PutBucketLifecycleConfigurationInput, optFns ...func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error) {
	if _, ok := c.buckets[*params.Bucket]; !ok {
		// bucket doesn't exist return no such bucket error
		return nil, &s3types.NoSuchBucket{Message: aws.String("The bucket doesn't exist")}
	}

	return &s3.PutBucketLifecycleConfigurationOutput{}, nil
}

func (c *mockBootstrapAthenaClient) CreateWorkGroup(ctx context.Context, params *athena.CreateWorkGroupInput, optFns ...func(*athena.Options)) (*athena.CreateWorkGroupOutput, error) {
	if c.workgroup != "" {
		return nil, &athenatypes.InvalidRequestException{Message: aws.String("workgroup is already created")}
	}

	c.workgroup = *params.Name

	return &athena.CreateWorkGroupOutput{}, nil
}

func (c *mockBootstrapGlueClient) UpdateTable(ctx context.Context, params *glue.UpdateTableInput, optFns ...func(*glue.Options)) (*glue.UpdateTableOutput, error) {
	if c.table == "" {
		return nil, &gluetypes.InvalidInputException{Message: aws.String("the table does not exist")}
	}

	return &glue.UpdateTableOutput{}, nil
}

func (c *mockBootstrapGlueClient) CreateTable(ctx context.Context, params *glue.CreateTableInput, optFns ...func(*glue.Options)) (*glue.CreateTableOutput, error) {
	if c.table != "" {
		return nil, &gluetypes.AlreadyExistsException{Message: aws.String("table already exists")}
	}

	c.table = *params.TableInput.Name

	return &glue.CreateTableOutput{}, nil
}

// Creates a new database in a Data Catalog.
func (c *mockBootstrapGlueClient) CreateDatabase(ctx context.Context, params *glue.CreateDatabaseInput, optFns ...func(*glue.Options)) (*glue.CreateDatabaseOutput, error) {
	if c.database != "" {
		return nil, &gluetypes.AlreadyExistsException{Message: aws.String("database already exists")}
	}

	c.database = *params.DatabaseInput.Name

	return &glue.CreateDatabaseOutput{}, nil
}
