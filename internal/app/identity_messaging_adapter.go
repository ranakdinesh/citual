package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	identity "github.com/ranakdinesh/spur-identity"
	msgdomain "github.com/ranakdinesh/spur-messaging/core/domain"
	msgports "github.com/ranakdinesh/spur-messaging/core/ports"
	"github.com/rs/zerolog"
)

type identityMessagingAdapter struct {
	gateway msgports.MessagingGateway
	log     *zerolog.Logger
}

func newIdentityMessagingAdapter(gateway msgports.MessagingGateway, log *zerolog.Logger) identity.CommunicationPort {
	return &identityMessagingAdapter{gateway: gateway, log: log}
}

func (a *identityMessagingAdapter) SendOTP(ctx context.Context, recipient string, channel string, code string) error {
	if a.gateway == nil {
		a.warn("otp", channel, recipient, nil)
		return nil
	}
	msgChannel := msgdomain.ChannelSMS
	if channel == "email" {
		msgChannel = msgdomain.ChannelEmail
	}
	tenantID := tenantIDFromContext(ctx)
	if tenantID == uuid.Nil {
		a.warn("otp", channel, recipient, errors.New("tenant id missing from context"))
		return nil
	}
	_, err := a.gateway.Submit(ctx, tenantID, msgports.MessagingRequest{
		Channel:        msgChannel,
		Recipient:      recipient,
		MessageType:    msgdomain.MessageTypeText,
		Subject:        "Your Citual verification code",
		TextBody:       fmt.Sprintf("Your Citual verification code is %s.", code),
		Category:       "transactional",
		IdempotencyKey: "identity:otp:" + recipient + ":" + code,
		Metadata: map[string]string{
			"source": "identity",
			"kind":   "otp",
		},
	})
	if err != nil {
		a.warn("otp", channel, recipient, err)
		return nil
	}
	return nil
}

func (a *identityMessagingAdapter) SendEmailVerification(ctx context.Context, message identity.EmailVerificationMessage) error {
	if a.gateway == nil {
		a.warn("email_verification", "email", message.Recipient, nil)
		return nil
	}
	tenantID := message.TenantID
	if tenantID == uuid.Nil {
		tenantID = tenantIDFromContext(ctx)
	}
	if tenantID == uuid.Nil {
		a.warn("email_verification", "email", message.Recipient, errors.New("tenant id missing from context"))
		return nil
	}

	htmlBody := fmt.Sprintf(
		`<p>Hello %s,</p><p>Please verify your email address to activate your Citual account.</p><p><a href="%s">Verify email</a></p><p>If you did not request this, you can ignore this message.</p>`,
		message.FirstName,
		message.VerificationURL,
	)
	textBody := fmt.Sprintf("Hello %s,\n\nPlease verify your email address to activate your Citual account:\n%s\n\nIf you did not request this, you can ignore this message.", message.FirstName, message.VerificationURL)
	verificationHash := sha256.Sum256([]byte(message.VerificationURL))

	_, err := a.gateway.Submit(ctx, tenantID, msgports.MessagingRequest{
		Channel:        msgdomain.ChannelEmail,
		Recipient:      message.Recipient,
		MessageType:    msgdomain.MessageTypeText,
		Subject:        "Verify your Citual email address",
		HTMLBody:       htmlBody,
		TextBody:       textBody,
		Category:       "transactional",
		IdempotencyKey: "identity:email-verification:" + message.Recipient + ":" + hex.EncodeToString(verificationHash[:8]),
		CorrelationID:  "identity.email_verification",
		Metadata: map[string]string{
			"source":       "identity",
			"kind":         "email_verification",
			"template_key": message.TemplateKey,
		},
	})
	if err != nil {
		a.warn("email_verification", "email", message.Recipient, err)
		return nil
	}
	return nil
}

func (a *identityMessagingAdapter) warn(kind, channel, recipient string, err error) {
	if a.log == nil {
		return
	}
	event := a.log.Warn().
		Str("component", "identity_messaging_adapter").
		Str("kind", kind).
		Str("channel", channel).
		Str("recipient", recipient)
	if err != nil {
		event = event.Err(err)
	}
	event.Msg("identity notification was not dispatched through messaging")
}

func tenantIDFromContext(ctx context.Context) uuid.UUID {
	return uuid.Nil
}
