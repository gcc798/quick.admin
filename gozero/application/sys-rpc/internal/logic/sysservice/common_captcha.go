package sysservicelogic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"math/rand"
	"strings"
	"time"

	"github.com/force-c/nai-tizi/application/sys-rpc/internal/svc"
	"github.com/force-c/nai-tizi/application/sys-rpc/pb"
	"github.com/google/uuid"
)

const (
	captchaImageKeyPrefix = "captcha:image:"
	captchaSmsKeyPrefix   = "captcha:sms:"
	captchaEmailKeyPrefix = "captcha:email:"
	captchaRatePrefix     = "captcha:ratelimit:"
)

func captchaEnabledTypes(svcCtx *svc.ServiceContext) []string {
	types := make([]string, 0, 3)
	if svcCtx.Config.Captcha.Image.Enabled {
		types = append(types, "image")
	}
	if svcCtx.Config.Captcha.Sms.Enabled {
		types = append(types, "sms")
	}
	if svcCtx.Config.Captcha.Email.Enabled {
		types = append(types, "email")
	}
	return types
}

func verifyImageCaptcha(ctx context.Context, svcCtx *svc.ServiceContext, uuidValue, code string) error {
	if !svcCtx.Config.Captcha.Image.Enabled {
		return nil
	}
	if uuidValue == "" || code == "" {
		return fmt.Errorf("验证码不能为空")
	}
	key := captchaImageKeyPrefix + uuidValue
	saved, err := svcCtx.Redis.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("验证码已过期或不存在")
	}
	_ = svcCtx.Redis.Del(ctx, key).Err()
	if !strings.EqualFold(saved, code) {
		return fmt.Errorf("验证码错误")
	}
	return nil
}

func newCaptchaData(captchaType, id string, data interface{}, expireAt time.Time) (*pb.CaptchaDataResp, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &pb.CaptchaDataResp{
		Id:       id,
		Type:     captchaType,
		DataJson: string(jsonData),
		ExpireAt: expireAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func generateImageCaptcha(ctx context.Context, svcCtx *svc.ServiceContext) (*pb.CaptchaDataResp, error) {
	id := uuid.NewString()
	code := randomCode(4)
	expire := 5 * time.Minute
	if err := svcCtx.Redis.Set(ctx, captchaImageKeyPrefix+id, code, expire).Err(); err != nil {
		return nil, err
	}
	svg := buildCaptchaSVG(code)
	image := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	return newCaptchaData("image", id, map[string]interface{}{"image": image}, time.Now().Add(expire))
}

func generateSmsCaptcha(ctx context.Context, svcCtx *svc.ServiceContext, phone string) (*pb.CaptchaDataResp, error) {
	if phone == "" {
		return nil, fmt.Errorf("手机号不能为空")
	}
	rateKey := captchaRatePrefix + "sms:" + phone
	exists, err := svcCtx.Redis.Exists(ctx, rateKey).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, fmt.Errorf("发送过于频繁，请稍后再试")
	}
	id := uuid.NewString()
	code := randomDigits(6)
	expire := 5 * time.Minute
	if err := svcCtx.Redis.HSet(ctx, captchaSmsKeyPrefix+id, map[string]interface{}{"code": code, "phone": phone}).Err(); err != nil {
		return nil, err
	}
	if err := svcCtx.Redis.Expire(ctx, captchaSmsKeyPrefix+id, expire).Err(); err != nil {
		return nil, err
	}
	_ = svcCtx.Redis.Set(ctx, rateKey, "1", time.Minute).Err()
	return newCaptchaData("sms", id, map[string]interface{}{"phone": maskPhone(phone)}, time.Now().Add(expire))
}

func generateEmailCaptcha(ctx context.Context, svcCtx *svc.ServiceContext, email string) (*pb.CaptchaDataResp, error) {
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}
	rateKey := captchaRatePrefix + "email:" + email
	exists, err := svcCtx.Redis.Exists(ctx, rateKey).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, fmt.Errorf("发送过于频繁，请稍后再试")
	}
	id := uuid.NewString()
	code := randomDigits(6)
	expire := 5 * time.Minute
	if err := svcCtx.Redis.HSet(ctx, captchaEmailKeyPrefix+id, map[string]interface{}{"code": code, "email": email}).Err(); err != nil {
		return nil, err
	}
	if err := svcCtx.Redis.Expire(ctx, captchaEmailKeyPrefix+id, expire).Err(); err != nil {
		return nil, err
	}
	_ = svcCtx.Redis.Set(ctx, rateKey, "1", time.Minute).Err()
	return newCaptchaData("email", id, map[string]interface{}{"email": maskEmail(email)}, time.Now().Add(expire))
}

func randomDigits(length int) string {
	const digits = "0123456789"
	return randomFromCharset(length, digits)
}

func randomCode(length int) string {
	const charset = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	return randomFromCharset(length, charset)
}

func randomFromCharset(length int, charset string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	out := make([]byte, length)
	for i := range out {
		out[i] = charset[r.Intn(len(charset))]
	}
	return string(out)
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) < 2 {
		return email
	}
	return parts[0][:1] + "***@" + parts[1]
}

func buildCaptchaSVG(code string) string {
	escaped := html.EscapeString(code)
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="120" height="40" viewBox="0 0 120 40">
<rect width="120" height="40" fill="#f5f5f5"/>
<path d="M0 10 C20 20, 40 0, 60 10 S100 20, 120 10" stroke="#d0d7de" fill="none"/>
<path d="M0 30 C20 20, 40 40, 60 30 S100 20, 120 30" stroke="#d0d7de" fill="none"/>
<text x="60" y="26" text-anchor="middle" font-size="22" font-family="monospace" fill="#1f2328" letter-spacing="4">%s</text>
</svg>`, escaped)
}
