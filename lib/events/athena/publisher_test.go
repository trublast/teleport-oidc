// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package athena

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	apievents "github.com/gravitational/teleport/api/types/events"
)

func init() {
	// Override maxS3BasedSize so we don't have to allocate 2GiB to test it.
	// Do this in init to avoid any race.
	maxS3BasedSize = maxDirectMessageSize * 4
}

// TODO(tobiaszheller): Those UT just cover basic stuff. When we will have consumer
// there will be UT which will cover whole flow of message with encoding/decoding.
func Test_EmitAuditEvent(t *testing.T) {
	veryLongString := strings.Repeat("t", maxS3BasedSize+1)
	tests := []struct {
		name          string
		in            apievents.AuditEvent
		publishErrors []error
		uploader      s3uploader
		wantCheck     func(t *testing.T, out []fakeQueueMessage)
		wantErrorMsg  string
	}{
		{
			name: "valid publish",
			in: &apievents.AppCreate{
				Metadata: apievents.Metadata{
					ID:   uuid.NewString(),
					Time: time.Now().UTC(),
				},
			},
			wantCheck: func(t *testing.T, out []fakeQueueMessage) {
				require.Len(t, out, 1)
				require.False(t, out[0].s3Based)
			},
		},
		{
			name: "should succeed after some retries",
			in: &apievents.AppCreate{
				Metadata: apievents.Metadata{
					ID:   uuid.NewString(),
					Time: time.Now().UTC(),
				},
			},
			publishErrors: []error{
				errors.New("error1"), errors.New("error2"),
			},
			wantCheck: func(t *testing.T, out []fakeQueueMessage) {
				require.Len(t, out, 1)
				require.False(t, out[0].s3Based)
			},
		},
		{
			name: "big message via s3",
			in: &apievents.AppCreate{
				Metadata: apievents.Metadata{
					ID:   uuid.NewString(),
					Time: time.Now().UTC(),
					Code: strings.Repeat("d", 2*maxDirectMessageSize),
				},
			},
			uploader: mockUploader{},
			wantCheck: func(t *testing.T, out []fakeQueueMessage) {
				require.Len(t, out, 1)
				require.True(t, out[0].s3Based)
			},
		},
		{
			name: "very big untrimmable event",
			in: &apievents.AppCreate{
				Metadata: apievents.Metadata{
					ID:   uuid.NewString(),
					Time: time.Now().UTC(),
					Code: veryLongString,
				},
			},
			uploader:     mockUploader{},
			wantErrorMsg: "message too large to publish",
		},
		{
			name: "very big trimmable event",
			in: &apievents.DatabaseSessionQuery{
				Metadata: apievents.Metadata{
					ID:   uuid.NewString(),
					Time: time.Now().UTC(),
				},
				DatabaseQuery: veryLongString,
			},
			uploader: mockUploader{},
			wantCheck: func(t *testing.T, out []fakeQueueMessage) {
				require.Len(t, out, 1)
				require.True(t, out[0].s3Based)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := newFakeQueue()
			p := &publisher{
				PublisherConfig: PublisherConfig{
					MessagePublisher: fq,
					Uploader:         tt.uploader,
				},
			}
			err := p.EmitAuditEvent(context.Background(), tt.in)
			if tt.wantErrorMsg != "" {
				require.ErrorContains(t, err, tt.wantErrorMsg)
				return
			}
			require.NoError(t, err)
			out := fq.dequeue()
			tt.wantCheck(t, out)
		})
	}
}

type mockUploader struct{}

func (m mockUploader) Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
	return &manager.UploadOutput{}, nil
}
