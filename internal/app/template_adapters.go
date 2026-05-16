package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	identity "github.com/ranakdinesh/spur-identity"
	msgdomain "github.com/ranakdinesh/spur-messaging/core/domain"
	msgports "github.com/ranakdinesh/spur-messaging/core/ports"
	storagedomain "github.com/ranakdinesh/spur-storage/core/domain"
	storageports "github.com/ranakdinesh/spur-storage/core/ports"
	template "github.com/ranakdinesh/spur-template"
	"github.com/rs/zerolog"
)

type identityTemplateAdapter struct {
	services *template.Services
}

func newIdentityTemplateAdapter(services *template.Services) *identityTemplateAdapter {
	return &identityTemplateAdapter{services: services}
}

func (a *identityTemplateAdapter) SendOTP(ctx context.Context, recipient string, channel string, code string) error {
	if a.services == nil || a.services.IdentityNotifications == nil {
		return nil
	}
	tenantID := tenantIDFromContext(ctx)
	if tenantID == uuid.Nil {
		return nil
	}
	templateChannel := template.ChannelSMS
	if channel == "email" {
		templateChannel = template.ChannelEmail
	}
	return a.services.IdentityNotifications.SendOTP(ctx, tenantID, recipient, templateChannel, code)
}

func (a *identityTemplateAdapter) SendEmailVerification(ctx context.Context, message identity.EmailVerificationMessage) error {
	if a.services == nil || a.services.IdentityNotifications == nil {
		return nil
	}
	return a.services.IdentityNotifications.SendEmailVerification(ctx, template.EmailVerificationInput{
		TenantID:        message.TenantID,
		UserID:          message.UserID,
		Recipient:       message.Recipient,
		FirstName:       message.FirstName,
		VerificationURL: message.VerificationURL,
		TemplateKey:     message.TemplateKey,
	})
}

func (a *identityTemplateAdapter) StoreAvatar(ctx context.Context, input identity.StoreAvatarInput) (*identity.StoredProfileFile, error) {
	if a.services == nil || a.services.FileStore == nil {
		return nil, errors.New("template file store is not configured")
	}
	stored, err := a.services.FileStore.StoreAvatar(ctx, template.StoreFileInput{
		TenantID:    input.TenantID,
		UserID:      input.UserID,
		Purpose:     template.FilePurposeUserAvatar,
		FileName:    input.FileName,
		ContentType: input.ContentType,
		Content:     input.Content,
	})
	if err != nil {
		return nil, err
	}
	return &identity.StoredProfileFile{
		ObjectID:    stored.ObjectID,
		ObjectKey:   stored.ObjectKey,
		Bucket:      stored.Bucket,
		ContentType: stored.ContentType,
		SizeBytes:   stored.SizeBytes,
	}, nil
}

type templateMessagingAdapter struct {
	gateway msgports.MessagingGateway
}

func newTemplateMessagingAdapter(gateway msgports.MessagingGateway) template.MessageGateway {
	return &templateMessagingAdapter{gateway: gateway}
}

func (a *templateMessagingAdapter) Submit(ctx context.Context, tenantID uuid.UUID, req template.MessageRequest) (*template.MessageReceipt, error) {
	if a.gateway == nil {
		return nil, errors.New("messaging gateway is not configured")
	}
	receipt, err := a.gateway.Submit(ctx, tenantID, msgports.MessagingRequest{
		Channel:        mapTemplateChannel(req.Channel),
		Recipient:      req.Recipient,
		MessageType:    msgdomain.MessageTypeText,
		Subject:        req.Subject,
		HTMLBody:       req.HTMLBody,
		TextBody:       req.TextBody,
		Metadata:       req.Metadata,
		IdempotencyKey: req.IdempotencyKey,
		Category:       req.Category,
		CorrelationID:  req.CorrelationID,
	})
	if err != nil {
		return nil, err
	}
	return &template.MessageReceipt{
		MessageID:      receipt.MessageID,
		Accepted:       receipt.Accepted,
		Status:         string(receipt.Status),
		IdempotencyKey: receipt.IdempotencyKey,
		CorrelationID:  receipt.CorrelationID,
		CreatedAt:      receipt.CreatedAt,
	}, nil
}

type templateStorageAdapter struct {
	storage storageports.StorageService
}

func newTemplateStorageAdapter(storage storageports.StorageService) template.FileStorage {
	return &templateStorageAdapter{storage: storage}
}

func (a *templateStorageAdapter) Store(ctx context.Context, input template.StoreFileInput) (*template.StoredFile, error) {
	if a.storage == nil {
		return nil, errors.New("storage service is not configured")
	}
	object, err := a.storage.UploadObject(ctx, storageports.UploadObjectInput{
		TenantID:    input.TenantID,
		UserID:      input.UserID,
		Scope:       storagedomain.ObjectScopeUser,
		Module:      "identity",
		FileName:    input.FileName,
		ContentType: input.ContentType,
		Content:     input.Content,
	})
	if err != nil {
		return nil, err
	}
	return &template.StoredFile{
		ObjectID:    object.ID,
		ObjectKey:   object.ObjectKey,
		Bucket:      object.Bucket,
		ContentType: object.ContentType,
		SizeBytes:   object.SizeBytes,
	}, nil
}

type templateLogger struct {
	log *zerolog.Logger
}

func (l templateLogger) Warn(_ context.Context, message string, fields map[string]any) {
	if l.log == nil {
		return
	}
	event := l.log.Warn().Str("component", "template")
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg(message)
}

func mapTemplateChannel(channel template.Channel) msgdomain.Channel {
	switch channel {
	case template.ChannelSMS:
		return msgdomain.ChannelSMS
	default:
		return msgdomain.ChannelEmail
	}
}

func tenantIDFromContext(ctx context.Context) uuid.UUID {
	return uuid.Nil
}
