package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	engagedomain "github.com/ranakdinesh/spur-engage/core/domain"
	engageports "github.com/ranakdinesh/spur-engage/core/ports"
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

func (a *identityTemplateAdapter) StoreTenantBrandAsset(ctx context.Context, input identity.StoreTenantBrandAssetInput) (*identity.StoredProfileFile, error) {
	if a.services == nil || a.services.FileStore == nil {
		return nil, errors.New("template file store is not configured")
	}
	purpose := template.FilePurposeTenantLogo
	if input.Purpose == "favicon" {
		purpose = template.FilePurposeTenantFavicon
	}
	stored, err := a.services.FileStore.StoreTenantBrandAsset(ctx, template.StoreFileInput{
		TenantID:    input.TenantID,
		Purpose:     purpose,
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

type identityEngageBrandRepository struct {
	service identity.TenantBrandService
}

func newIdentityEngageBrandRepository() *identityEngageBrandRepository {
	return &identityEngageBrandRepository{}
}

func (r *identityEngageBrandRepository) GetBrandProfile(ctx context.Context, tenantID uuid.UUID) (*engagedomain.BrandProfile, error) {
	if r.service == nil {
		return nil, errors.New("identity brand service is not configured")
	}
	brand, err := r.service.GetTenantBrandProfile(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return mapIdentityBrandToEngage(brand), nil
}

func (r *identityEngageBrandRepository) UpsertBrandProfile(ctx context.Context, brand *engagedomain.BrandProfile) (*engagedomain.BrandProfile, error) {
	if r.service == nil {
		return nil, errors.New("identity brand service is not configured")
	}
	if brand == nil {
		return nil, errors.New("brand profile is required")
	}
	saved, err := r.service.UpsertTenantBrandProfile(ctx, identity.TenantBrandProfileInput{
		TenantID:            brand.TenantID,
		BrandName:           brand.BrandName,
		LegalName:           valueOf(brand.LegalName),
		Tagline:             valueOf(brand.Tagline),
		Description:         valueOf(brand.Description),
		WebsiteURL:          valueOf(brand.WebsiteURL),
		PrimaryColor:        brand.PrimaryColor,
		SecondaryColor:      brand.SecondaryColor,
		AccentColor:         brand.AccentColor,
		BackgroundColor:     brand.BackgroundColor,
		TextColor:           brand.TextColor,
		FontFamily:          valueOf(brand.FontFamily),
		ToneOfVoice:         valueOf(brand.ToneOfVoice),
		AudienceSummary:     valueOf(brand.AudienceSummary),
		ValueProposition:    valueOf(brand.ValueProposition),
		CTAStyle:            valueOf(brand.CTAStyle),
		PrivacyPolicyURL:    valueOf(brand.PrivacyPolicyURL),
		TermsURL:            valueOf(brand.TermsURL),
		ContactEmail:        valueOf(brand.ContactEmail),
		ContactPhone:        valueOf(brand.ContactPhone),
		WhatsAppNumber:      valueOf(brand.WhatsAppNumber),
		LinkedInURL:         valueOf(brand.LinkedInURL),
		FacebookURL:         valueOf(brand.FacebookURL),
		XURL:                valueOf(brand.XURL),
		InstagramURL:        valueOf(brand.InstagramURL),
		YouTubeURL:          valueOf(brand.YouTubeURL),
		DefaultUTMSource:    valueOf(brand.DefaultUTMSource),
		DefaultUTMMedium:    valueOf(brand.DefaultUTMMedium),
		DefaultUTMCampaign:  valueOf(brand.DefaultUTMCampaign),
		ComplianceFooter:    valueOf(brand.ComplianceFooter),
		DisallowedClaims:    brand.DisallowedClaims,
		RequiredDisclosures: brand.RequiredDisclosures,
		Metadata:            brand.Metadata,
	})
	if err != nil {
		return nil, err
	}
	return mapIdentityBrandToEngage(saved), nil
}

func mapIdentityBrandToEngage(brand *identity.TenantBrandProfile) *engagedomain.BrandProfile {
	if brand == nil {
		return nil
	}
	return &engagedomain.BrandProfile{
		ID:                  uuid.Nil,
		TenantID:            brand.TenantID,
		BrandName:           brand.BrandName,
		LegalName:           ptrOf(brand.LegalName),
		Tagline:             ptrOf(brand.Tagline),
		Description:         ptrOf(brand.Description),
		WebsiteURL:          ptrOf(brand.WebsiteURL),
		LogoURL:             ptrOf(brand.LogoURL),
		FaviconURL:          ptrOf(brand.FaviconURL),
		PrimaryColor:        brand.PrimaryColor,
		SecondaryColor:      brand.SecondaryColor,
		AccentColor:         brand.AccentColor,
		BackgroundColor:     brand.BackgroundColor,
		TextColor:           brand.TextColor,
		FontFamily:          ptrOf(brand.FontFamily),
		ToneOfVoice:         ptrOf(brand.ToneOfVoice),
		AudienceSummary:     ptrOf(brand.AudienceSummary),
		ValueProposition:    ptrOf(brand.ValueProposition),
		CTAStyle:            ptrOf(brand.CTAStyle),
		PrivacyPolicyURL:    ptrOf(brand.PrivacyPolicyURL),
		TermsURL:            ptrOf(brand.TermsURL),
		ContactEmail:        ptrOf(brand.ContactEmail),
		ContactPhone:        ptrOf(brand.ContactPhone),
		LinkedInURL:         ptrOf(brand.LinkedInURL),
		FacebookURL:         ptrOf(brand.FacebookURL),
		InstagramURL:        ptrOf(brand.InstagramURL),
		XURL:                ptrOf(brand.XURL),
		YouTubeURL:          ptrOf(brand.YouTubeURL),
		WhatsAppNumber:      ptrOf(brand.WhatsAppNumber),
		DefaultUTMSource:    ptrOf(brand.DefaultUTMSource),
		DefaultUTMMedium:    ptrOf(brand.DefaultUTMMedium),
		DefaultUTMCampaign:  ptrOf(brand.DefaultUTMCampaign),
		ComplianceFooter:    ptrOf(brand.ComplianceFooter),
		DisallowedClaims:    brand.DisallowedClaims,
		RequiredDisclosures: brand.RequiredDisclosures,
		Metadata:            brand.Metadata,
		CreatedAt:           brand.CreatedAt,
		UpdatedAt:           brand.UpdatedAt,
	}
}

func ptrOf(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func valueOf(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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
		MessageID:         receipt.MessageID,
		Accepted:          receipt.Accepted,
		Status:            string(receipt.Status),
		IdempotencyKey:    receipt.IdempotencyKey,
		CorrelationID:     receipt.CorrelationID,
		ProviderMessageID: receipt.ProviderMessageID,
		CreatedAt:         receipt.CreatedAt,
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

type messagingEngageInboxAdapter struct {
	inbox engageports.InboxService
}

func newMessagingEngageInboxAdapter(inbox engageports.InboxService) msgports.EngageInboxPublisher {
	return &messagingEngageInboxAdapter{inbox: inbox}
}

func (a *messagingEngageInboxAdapter) PublishInboxEvent(ctx context.Context, tenantID uuid.UUID, event msgports.EngageInboxEvent) error {
	if a.inbox == nil {
		return nil
	}
	occurredAt := event.OccurredAt
	_, _, err := a.inbox.IngestMessageEvent(ctx, tenantID, engageports.InboxMessageEventInput{
		Channel:                mapEngageInboxChannel(event.Channel),
		ExternalConversationID: stringPtr(event.ExternalConversationID),
		ExternalMessageID:      stringPtr(event.ExternalMessageID),
		Direction:              engagedomain.InboxDirection(event.Direction),
		Sender:                 stringPtr(event.Sender),
		Recipient:              stringPtr(event.Recipient),
		ParticipantAddress:     event.ParticipantAddress,
		MessageType:            string(event.MessageType),
		TextBody:               stringPtr(event.TextBody),
		MediaURL:               stringPtr(event.MediaURL),
		Provider:               stringPtr(event.Provider),
		ProviderStatus:         stringPtr(event.ProviderStatus),
		OccurredAt:             &occurredAt,
		Metadata:               event.Metadata,
	})
	return err
}

type engageTemplateAdapter struct {
	services *template.Services
}

func (a *engageTemplateAdapter) SendInboxReply(ctx context.Context, tenantID uuid.UUID, input engageports.DispatchInboxReplyInput) (*engageports.DispatchReceipt, error) {
	if a.services == nil || a.services.EngageMessages == nil {
		return nil, errors.New("template engage messaging is not configured")
	}
	receipt, err := a.services.EngageMessages.SendReply(ctx, template.EngageReplyInput{
		TenantID:       tenantID,
		ThreadID:       input.ThreadID,
		Channel:        mapTemplateInboxChannel(input.Channel),
		Recipient:      input.Recipient,
		TextBody:       input.TextBody,
		Subject:        input.Subject,
		IdempotencyKey: input.IdempotencyKey,
		CorrelationID:  input.CorrelationID,
		Metadata:       input.Metadata,
	})
	if err != nil {
		return nil, err
	}
	return &engageports.DispatchReceipt{
		MessageID:         receipt.MessageID,
		Accepted:          receipt.Accepted,
		Status:            receipt.Status,
		IdempotencyKey:    receipt.IdempotencyKey,
		CorrelationID:     receipt.CorrelationID,
		ProviderMessageID: receipt.ProviderMessageID,
		CreatedAt:         receipt.CreatedAt,
	}, nil
}

type templateLogger struct {
	log *zerolog.Logger
}

func mapEngageInboxChannel(channel msgdomain.Channel) engagedomain.InboxChannel {
	switch channel {
	case msgdomain.ChannelWhatsApp:
		return engagedomain.InboxChannelWhatsApp
	case msgdomain.ChannelEmail:
		return engagedomain.InboxChannelEmail
	case msgdomain.ChannelSMS:
		return engagedomain.InboxChannelSMS
	default:
		return engagedomain.InboxChannelOther
	}
}

func mapTemplateInboxChannel(channel engagedomain.InboxChannel) template.Channel {
	switch channel {
	case engagedomain.InboxChannelSMS:
		return template.ChannelSMS
	case engagedomain.InboxChannelWhatsApp:
		return template.ChannelWhatsApp
	default:
		return template.ChannelEmail
	}
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
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
	case template.ChannelWhatsApp:
		return msgdomain.ChannelWhatsApp
	default:
		return msgdomain.ChannelEmail
	}
}

func tenantIDFromContext(ctx context.Context) uuid.UUID {
	return uuid.Nil
}
