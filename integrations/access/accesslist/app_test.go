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

package accesslist

import (
	"context"
	"testing"
	"time"

	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/types/accesslist"
	"github.com/gravitational/teleport/api/types/header"
	"github.com/gravitational/teleport/integrations/access/common"
	"github.com/gravitational/teleport/integrations/access/common/teleport"
	"github.com/gravitational/teleport/lib/auth"
	"github.com/gravitational/teleport/lib/services"
)

type mockMessagingBot struct {
	lastReminderRecipients []common.Recipient
	recipients             map[string]*common.Recipient
}

func (m *mockMessagingBot) CheckHealth(ctx context.Context) error {
	return nil
}

func (m *mockMessagingBot) SendReviewReminders(ctx context.Context, recipient common.Recipient, accessList *accesslist.AccessList) error {
	m.lastReminderRecipients = append(m.lastReminderRecipients, recipient)
	return nil
}

func (m *mockMessagingBot) FetchRecipient(ctx context.Context, recipient string) (*common.Recipient, error) {
	fetchedRecipient, ok := m.recipients[recipient]
	if !ok {
		return nil, trace.NotFound("recipient %s not found", recipient)
	}

	return fetchedRecipient, nil
}

func (m *mockMessagingBot) SupportedApps() []common.App {
	return []common.App{
		NewApp(m),
	}
}

type mockPluginConfig struct {
	client teleport.Client
	bot    *mockMessagingBot
}

func (m *mockPluginConfig) GetTeleportClient(ctx context.Context) (teleport.Client, error) {
	return m.client, nil
}

func (m *mockPluginConfig) GetRecipients() common.RawRecipientsMap {
	return nil
}

func (m *mockPluginConfig) NewBot(clusterName string, webProxyAddr string) (common.MessagingBot, error) {
	return m.bot, nil
}

func (m *mockPluginConfig) GetPluginType() types.PluginType {
	return types.PluginTypeSlack
}

func TestAccessListReminders(t *testing.T) {
	t.Parallel()

	clock := clockwork.NewFakeClockAt(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))

	server, err := auth.NewTestServer(auth.TestServerConfig{
		Auth: auth.TestAuthServerConfig{
			Dir:   t.TempDir(),
			Clock: clockwork.NewFakeClock(),
		},
	})
	require.NoError(t, err)
	as := server.Auth()
	t.Cleanup(func() {
		require.NoError(t, as.Close())
	})

	bot := &mockMessagingBot{
		recipients: map[string]*common.Recipient{
			"owner1": {Name: "owner1"},
			"owner2": {Name: "owner2"},
		},
	}
	app := common.NewApp(&mockPluginConfig{client: as, bot: bot}, "test-plugin")
	app.Clock = clock
	ctx := context.Background()
	go func() {
		app.Run(ctx)
	}()

	ready, err := app.WaitReady(ctx)
	require.NoError(t, err)
	require.True(t, ready)

	t.Cleanup(func() {
		app.Terminate()
		<-app.Done()
		require.NoError(t, app.Err())
	})

	accessList, err := accesslist.NewAccessList(header.Metadata{
		Name: "test-access-list",
	}, accesslist.Spec{
		Title:  "test access list",
		Owners: []accesslist.Owner{{Name: "owner1"}, {Name: "not-found"}},
		Grants: accesslist.Grants{
			Roles: []string{"role"},
		},
		Audit: accesslist.Audit{
			NextAuditDate: clock.Now().Add(28 * 24 * time.Hour), // Four weeks out from today
			Notifications: accesslist.Notifications{
				Start: oneDay * 14, // Start alerting at two weeks before audit date
			},
		},
	})
	require.NoError(t, err)

	// No notifications for today
	advanceAndLookForRecipients(t, bot, as, clock, 0, accessList)

	// Advance by one week, expect no notifications.
	advanceAndLookForRecipients(t, bot, as, clock, oneDay*7, accessList)

	// Advance by one week, expect a notification. "not-found" will be missing as a recipient.
	advanceAndLookForRecipients(t, bot, as, clock, oneDay*7, accessList, "owner1")

	// Add a new owner.
	accessList.Spec.Owners = append(accessList.Spec.Owners, accesslist.Owner{Name: "owner2"})

	// Advance by one day, expect a notification only to the new owner.
	advanceAndLookForRecipients(t, bot, as, clock, oneDay, accessList, "owner2")

	// Advance by one day, expect no notifications.
	advanceAndLookForRecipients(t, bot, as, clock, oneDay, accessList)

	// Advance by five more days, to the next week, expect two notifications
	advanceAndLookForRecipients(t, bot, as, clock, oneDay*5, accessList, "owner1", "owner2")

	// Advance by one day, expect no notifications
	advanceAndLookForRecipients(t, bot, as, clock, oneDay, accessList)

	// Advance by one day, expect no notifications
	advanceAndLookForRecipients(t, bot, as, clock, oneDay, accessList)

	// Advance by five more days, to the next week, expect two notifications
	advanceAndLookForRecipients(t, bot, as, clock, oneDay*5, accessList, "owner1", "owner2")

	// Advance 60 days a day at a time, expect two notifications each time.
	for i := 0; i < 60; i++ {
		// Make sure we only get a notification once per day by iterating through each 6 hours at a time.
		for j := 0; j < 3; j++ {
			advanceAndLookForRecipients(t, bot, as, clock, 6*time.Hour, accessList)
		}
		advanceAndLookForRecipients(t, bot, as, clock, 6*time.Hour, accessList, "owner1", "owner2")
	}
}

type mockClient struct {
	mock.Mock
	teleport.Client
}

func (m *mockClient) ListAccessLists(ctx context.Context, pageSize int, pageToken string) ([]*accesslist.AccessList, string, error) {
	args := m.Called(ctx, pageSize, pageToken)
	return args.Get(0).([]*accesslist.AccessList), args.String(1), args.Error(2)
}

func TestAccessListReminders_BadClient(t *testing.T) {
	t.Parallel()

	clock := clockwork.NewFakeClockAt(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))

	server, err := auth.NewTestServer(auth.TestServerConfig{
		Auth: auth.TestAuthServerConfig{
			Dir:   t.TempDir(),
			Clock: clockwork.NewFakeClock(),
		},
	})
	require.NoError(t, err)
	as := server.Auth()
	t.Cleanup(func() {
		require.NoError(t, as.Close())
	})

	// Use this mock client so that we can force ListAccessLists to return an error.
	client := &mockClient{
		Client: as,
	}
	client.On("ListAccessLists", mock.Anything, mock.Anything, mock.Anything).Return(([]*accesslist.AccessList)(nil), "", trace.BadParameter("error"))

	bot := &mockMessagingBot{
		recipients: map[string]*common.Recipient{
			"owner1": {Name: "owner1"},
			"owner2": {Name: "owner2"},
		},
	}
	app := common.NewApp(&mockPluginConfig{client: client, bot: bot}, "test-plugin")
	app.Clock = clock
	ctx := context.Background()
	go func() {
		app.Run(ctx)
	}()

	ready, err := app.WaitReady(ctx)
	require.NoError(t, err)
	require.True(t, ready)

	t.Cleanup(func() {
		app.Terminate()
		<-app.Done()
		require.NoError(t, app.Err())
	})

	clock.BlockUntil(1)
	for i := 1; i <= 6; i++ {
		clock.Advance(3 * time.Hour)
		clock.BlockUntil(1)
		client.AssertNumberOfCalls(t, "ListAccessLists", i)
	}
}

func advanceAndLookForRecipients(t *testing.T,
	bot *mockMessagingBot,
	alSvc services.AccessLists,
	clock clockwork.FakeClock,
	advance time.Duration,
	accessList *accesslist.AccessList,
	recipients ...string) {
	t.Helper()

	ctx := context.Background()

	_, err := alSvc.UpsertAccessList(ctx, accessList)
	require.NoError(t, err)

	bot.lastReminderRecipients = nil

	var expectedRecipients []common.Recipient
	if len(recipients) > 0 {
		expectedRecipients = make([]common.Recipient, len(recipients))
		for i, r := range recipients {
			expectedRecipients[i] = common.Recipient{Name: r}
		}
	}
	clock.Advance(advance)
	clock.BlockUntil(1)

	require.ElementsMatch(t, expectedRecipients, bot.lastReminderRecipients)
}
