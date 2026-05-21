package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	engagedomain "github.com/ranakdinesh/spur-engage/core/domain"
	engageports "github.com/ranakdinesh/spur-engage/core/ports"
	msgdomain "github.com/ranakdinesh/spur-messaging/core/domain"
	msgports "github.com/ranakdinesh/spur-messaging/core/ports"
)

func TestMessagingEngageInboxAdapterPublishesIntoEngageInbox(t *testing.T) {
	tenantID := uuid.New()
	occurredAt := time.Now().UTC()
	inbox := &fakeEngageInboxService{}
	adapter := newMessagingEngageInboxAdapter(inbox)

	err := adapter.PublishInboxEvent(context.Background(), tenantID, msgports.EngageInboxEvent{
		Channel:                msgdomain.ChannelEmail,
		ExternalConversationID: "conversation-123",
		ExternalMessageID:      "email-message-123",
		Direction:              "inbound",
		Sender:                 "lead@example.com",
		Recipient:              "sales@citual.test",
		ParticipantAddress:     "lead@example.com",
		MessageType:            msgdomain.MessageTypeText,
		TextBody:               "Can Citual help with automation and security?",
		Provider:               "sendgrid",
		ProviderStatus:         "replied",
		OccurredAt:             occurredAt,
		Metadata:               map[string]any{"source": "test"},
	})
	if err != nil {
		t.Fatalf("PublishInboxEvent returned error: %v", err)
	}
	if inbox.calls != 1 {
		t.Fatalf("expected one inbox ingest call, got %d", inbox.calls)
	}
	if inbox.tenantID != tenantID {
		t.Fatalf("expected tenant %s, got %s", tenantID, inbox.tenantID)
	}
	input := inbox.input
	if input.Channel != engagedomain.InboxChannelEmail {
		t.Fatalf("expected email inbox channel, got %q", input.Channel)
	}
	if input.Direction != engagedomain.InboxDirectionInbound {
		t.Fatalf("expected inbound direction, got %q", input.Direction)
	}
	if input.ParticipantAddress != "lead@example.com" {
		t.Fatalf("expected participant address, got %q", input.ParticipantAddress)
	}
	if input.ExternalMessageID == nil || *input.ExternalMessageID != "email-message-123" {
		t.Fatalf("expected external message id to be mapped, got %#v", input.ExternalMessageID)
	}
	if input.TextBody == nil || *input.TextBody == "" {
		t.Fatalf("expected text body to be mapped")
	}
	if input.Provider == nil || *input.Provider != "sendgrid" {
		t.Fatalf("expected provider to be mapped, got %#v", input.Provider)
	}
}

type fakeEngageInboxService struct {
	calls    int
	tenantID uuid.UUID
	input    engageports.InboxMessageEventInput
}

func (s *fakeEngageInboxService) IngestMessageEvent(_ context.Context, tenantID uuid.UUID, input engageports.InboxMessageEventInput) (*engagedomain.InboxThread, *engagedomain.InboxMessage, error) {
	s.calls++
	s.tenantID = tenantID
	s.input = input
	return &engagedomain.InboxThread{ID: uuid.New(), TenantID: tenantID}, &engagedomain.InboxMessage{ID: uuid.New(), TenantID: tenantID}, nil
}

func (s *fakeEngageInboxService) GetInboxThread(context.Context, uuid.UUID, uuid.UUID) (*engagedomain.InboxThread, error) {
	return nil, nil
}

func (s *fakeEngageInboxService) ListInboxThreads(context.Context, uuid.UUID, engageports.InboxThreadListOptions) (*engageports.ListResult[engagedomain.InboxThread], error) {
	return nil, nil
}

func (s *fakeEngageInboxService) ListInboxMessages(context.Context, uuid.UUID, uuid.UUID, engageports.ListOptions) (*engageports.ListResult[engagedomain.InboxMessage], error) {
	return nil, nil
}

func (s *fakeEngageInboxService) SendInboxReply(context.Context, uuid.UUID, uuid.UUID, engageports.InboxReplyInput) (*engagedomain.InboxThread, *engagedomain.InboxMessage, error) {
	return nil, nil, nil
}

func (s *fakeEngageInboxService) UpdateInboxThread(context.Context, uuid.UUID, uuid.UUID, engageports.InboxThreadUpdateInput) (*engagedomain.InboxThread, error) {
	return nil, nil
}

func (s *fakeEngageInboxService) MarkInboxThreadRead(context.Context, uuid.UUID, uuid.UUID) (*engagedomain.InboxThread, error) {
	return nil, nil
}

func (s *fakeEngageInboxService) ArchiveInboxThread(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
