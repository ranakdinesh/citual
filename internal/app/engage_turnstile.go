package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type engageTurnstileVerifier struct {
	secret string
	client *http.Client
}

func newEngageTurnstileVerifier(secret string) *engageTurnstileVerifier {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil
	}
	return &engageTurnstileVerifier{
		secret: secret,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (v *engageTurnstileVerifier) VerifyPublicLeadCaptcha(ctx context.Context, token, ipAddress string) (bool, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return false, nil
	}
	values := url.Values{}
	values.Set("secret", v.secret)
	values.Set("response", token)
	if strings.TrimSpace(ipAddress) != "" {
		values.Set("remoteip", strings.TrimSpace(ipAddress))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, turnstileVerifyURL, strings.NewReader(values.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := v.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, nil
	}
	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.Success, nil
}
