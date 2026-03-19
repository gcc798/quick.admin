package data

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

const (
	captchaImageTTL = 2 * time.Minute
	captchaSMSTTL   = 5 * time.Minute
	captchaEmailTTL = 5 * time.Minute
	captchaRateTTL  = 60 * time.Second
)

var phoneRegexp = regexp.MustCompile(`^1\d{10}$`)

func (r *Resources) CreateImageCaptcha(ctx context.Context) (string, string, time.Time, error) {
	if r == nil || r.Redis == nil {
		return "", "", time.Time{}, fmt.Errorf("redis is not configured")
	}
	captchaID, err := randomHex(12)
	if err != nil {
		return "", "", time.Time{}, err
	}
	code := randomDigitsString(4)
	expireAt := time.Now().Add(captchaImageTTL)
	if err := r.Redis.Set(ctx, captchaKey("image", captchaID), code, captchaImageTTL).Err(); err != nil {
		return "", "", time.Time{}, err
	}
	return captchaID, imageDataURI(code), expireAt, nil
}

func (r *Resources) CreateSMSCaptcha(ctx context.Context, phone string) (string, time.Time, error) {
	if r == nil || r.Redis == nil {
		return "", time.Time{}, fmt.Errorf("redis is not configured")
	}
	phone = strings.TrimSpace(phone)
	if !phoneRegexp.MatchString(phone) {
		return "", time.Time{}, fmt.Errorf("invalid phone number")
	}
	rateKey := captchaRateKey("sms", phone)
	exists, err := r.Redis.Exists(ctx, rateKey).Result()
	if err != nil {
		return "", time.Time{}, err
	}
	if exists > 0 {
		return "", time.Time{}, fmt.Errorf("captcha sent too frequently")
	}
	captchaID, err := randomHex(12)
	if err != nil {
		return "", time.Time{}, err
	}
	expireAt := time.Now().Add(captchaSMSTTL)
	data := map[string]any{
		"code":   randomDigitsString(6),
		"target": phone,
	}
	if err := r.Redis.HSet(ctx, captchaKey("sms", captchaID), data).Err(); err != nil {
		return "", time.Time{}, err
	}
	if err := r.Redis.Expire(ctx, captchaKey("sms", captchaID), captchaSMSTTL).Err(); err != nil {
		return "", time.Time{}, err
	}
	if err := r.Redis.Set(ctx, rateKey, "1", captchaRateTTL).Err(); err != nil {
		return "", time.Time{}, err
	}
	return captchaID, expireAt, nil
}

func (r *Resources) CreateEmailCaptcha(ctx context.Context, email string) (string, time.Time, error) {
	if r == nil || r.Redis == nil {
		return "", time.Time{}, fmt.Errorf("redis is not configured")
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return "", time.Time{}, fmt.Errorf("invalid email")
	}
	rateKey := captchaRateKey("email", email)
	exists, err := r.Redis.Exists(ctx, rateKey).Result()
	if err != nil {
		return "", time.Time{}, err
	}
	if exists > 0 {
		return "", time.Time{}, fmt.Errorf("captcha sent too frequently")
	}
	captchaID, err := randomHex(12)
	if err != nil {
		return "", time.Time{}, err
	}
	expireAt := time.Now().Add(captchaEmailTTL)
	data := map[string]any{
		"code":   randomDigitsString(6),
		"target": email,
	}
	if err := r.Redis.HSet(ctx, captchaKey("email", captchaID), data).Err(); err != nil {
		return "", time.Time{}, err
	}
	if err := r.Redis.Expire(ctx, captchaKey("email", captchaID), captchaEmailTTL).Err(); err != nil {
		return "", time.Time{}, err
	}
	if err := r.Redis.Set(ctx, rateKey, "1", captchaRateTTL).Err(); err != nil {
		return "", time.Time{}, err
	}
	return captchaID, expireAt, nil
}

func (r *Resources) VerifyImageCaptcha(ctx context.Context, captchaID, code string) error {
	if r == nil || r.Redis == nil {
		return fmt.Errorf("redis is not configured")
	}
	storedCode, err := r.Redis.Get(ctx, captchaKey("image", strings.TrimSpace(captchaID))).Result()
	if err != nil {
		return fmt.Errorf("captcha expired or not found")
	}
	_ = r.Redis.Del(ctx, captchaKey("image", strings.TrimSpace(captchaID))).Err()
	if !strings.EqualFold(strings.TrimSpace(storedCode), strings.TrimSpace(code)) {
		return errors.New("captcha is invalid")
	}
	return nil
}

func (r *Resources) VerifySMSCaptcha(ctx context.Context, phone, code string) error {
	return r.verifyTargetCaptcha(ctx, "sms", phone, code)
}

func (r *Resources) VerifyEmailCaptcha(ctx context.Context, email, code string) error {
	return r.verifyTargetCaptcha(ctx, "email", strings.ToLower(strings.TrimSpace(email)), code)
}

func captchaKey(kind, id string) string {
	return "captcha:" + kind + ":" + id
}

func captchaRateKey(kind, target string) string {
	return "captcha:" + kind + ":ratelimit:" + target
}

func imageDataURI(code string) string {
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="120" height="40" viewBox="0 0 120 40"><rect width="120" height="40" fill="#f5f7fa"/><text x="50%%" y="50%%" dominant-baseline="middle" text-anchor="middle" font-family="monospace" font-size="24" fill="#1f2937">%s</text></svg>`,
		html.EscapeString(code),
	)
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

func randomDigitsString(length int) string {
	value, err := randomHex(length)
	if err != nil || value == "" {
		return strings.Repeat("0", length)
	}
	var builder strings.Builder
	builder.Grow(length)
	for i := 0; i < length; i++ {
		builder.WriteByte('0' + value[i]%10)
	}
	return builder.String()
}

func (r *Resources) verifyTargetCaptcha(ctx context.Context, kind, target, code string) error {
	if r == nil || r.Redis == nil {
		return fmt.Errorf("redis is not configured")
	}
	target = strings.TrimSpace(target)
	code = strings.TrimSpace(code)
	if target == "" || code == "" {
		return errors.New("captcha target and code are required")
	}
	iter := r.Redis.Scan(ctx, 0, captchaKey(kind, "*"), 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		values, err := r.Redis.HGetAll(ctx, key).Result()
		if err != nil || len(values) == 0 {
			continue
		}
		if values["target"] != target {
			continue
		}
		_ = r.Redis.Del(ctx, key).Err()
		if values["code"] != code {
			return errors.New("captcha is invalid")
		}
		return nil
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return errors.New("captcha expired or not found")
}
