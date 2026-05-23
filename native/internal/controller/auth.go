package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gcc798/quick.admin/internal/utils/idgen"

	"github.com/gcc798/quick.admin/internal/container"
	"github.com/gcc798/quick.admin/internal/domain/model"
	"github.com/gcc798/quick.admin/internal/domain/request"
	"github.com/gcc798/quick.admin/internal/domain/response"
	"github.com/gcc798/quick.admin/internal/service"
	"github.com/gcc798/quick.admin/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthController 定义业务数据结构。
type AuthController interface {
	Login(c *gin.Context)        // 用户登录
	Logout(c *gin.Context)       // 用户登出
	RefreshToken(c *gin.Context) // 刷新访问令牌
}

type authController struct {
	ctr                    container.Container
	base                   *BaseController
	clientService          service.ClientService
	tokenManager           service.TokenManager
	concurrentLoginManager service.ConcurrentLoginManager
	strategyFactory        *StrategyFactory
	smsService             interface {
		SendVerificationCode(ctx context.Context, phonenumber string) (string, error)
	}
}

// NewAuthController 创建组件实例。
func NewAuthController(c container.Container) AuthController {
	clientService := service.NewClientService(c.GetDB(), c.GetRedis(), c.GetLogger())
	tokenManager := service.NewTokenManager(c.GetJWT(), c.GetRedis(), c.GetLogger())
	concurrentLoginManager := service.NewConcurrentLoginManager(c.GetRedis(), tokenManager, c.GetConfig(), c.GetLogger())

	strategyFactory := NewStrategyFactory()
	strategyFactory.Register(NewPasswordAuthStrategy(c))
	strategyFactory.Register(NewXcxAuthStrategy(c))
	strategyFactory.Register(NewEmailAuthStrategy(c))
	strategyFactory.Register(NewWechatOnlyAuthStrategy(c))
	strategyFactory.Register(NewSmsAuthStrategy(c))

	return &authController{
		ctr:                    c,
		base:                   NewBaseController(c),
		clientService:          clientService,
		tokenManager:           tokenManager,
		concurrentLoginManager: concurrentLoginManager,
		strategyFactory:        strategyFactory,
		smsService:             c.GetSMS(),
	}
}

// Login godoc
//
//	@Summary		用户登录
//	@Description	支持多种登录方式：密码登录(password)、邮箱验证码(email)、微信小程序(xcx)
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			request	body		request.LoginRequest	true	"登录请求参数"
//	@Success		200		{object}	response.Response{data=response.LoginResponse}
//	@Failure		400		{object}	response.Response	"参数错误"
//	@Failure		401		{object}	response.Response	"认证失败"
//	@Router			/login [post]
func (h *authController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}
	ctx := c.Request.Context()
	req.Normalize()
	req.LoginIP = c.ClientIP()
	loginAccount := resolveLoginAccount(&req)

	client, err := h.authenticateClient(ctx, &req)
	if err != nil {
		h.recordLoginLog(c, loginAccount, req.ClientID, 1, err.Error())
		response.FailCode(c, response.CodeUnauthorized, err.Error())
		return
	}

	resp, err := h.strategyFactory.Login(ctx, &req)
	if err != nil {
		h.ctr.GetLogger().Error("login failed",
			zap.String("clientId", client.ClientId),
			zap.String("grantType", req.GrantType),
			zap.String("username", req.Username),
			zap.Error(err))
		h.recordLoginLog(c, loginAccount, client.ClientId, 1, err.Error())
		response.FailCode(c, response.CodeUnauthorized, err.Error())
		return
	}

	user := resp.UserInfo

	useExisting, existingToken, err := h.concurrentLoginManager.HandleConcurrentLogin(
		ctx, user.UserId, client.ClientId, client.Timeout,
	)
	if err != nil {
		h.ctr.GetLogger().Warn("handle concurrent login failed", zap.Error(err))
	}

	if useExisting && existingToken != "" {
		h.ctr.GetLogger().Info("reuse existing token",
			zap.String("clientId", client.ClientId),
			zap.Int64("userId", user.UserId))
		h.recordLoginLog(c, user.Username, client.ClientId, 0, "登录成功（复用Token）")

		sysUser := &model.User{
			ID:          user.UserId,
			UserName:    user.Username,
			NickName:    user.Nickname,
			Phonenumber: user.Phonenumber,
			Email:       user.Email,
			Avatar:      user.Avatar,
			UserType:    user.UserType,
		}
		accessToken, refreshToken, accessExpiresIn, refreshExpiresIn, err := h.tokenManager.GenerateTokenPair(ctx, sysUser, client)
		if err != nil {
			h.ctr.GetLogger().Error("failed to generate token pair", zap.Error(err))
			response.FailCode(c, response.CodeServerError, "生成Token失败")
			return
		}

		response.Success(c, buildLoginResponse(accessToken, refreshToken, accessExpiresIn, refreshExpiresIn, client, user))
		return
	}

	sysUser := &model.User{
		ID:          user.UserId,
		UserName:    user.Username,
		NickName:    user.Nickname,
		Phonenumber: user.Phonenumber,
		Email:       user.Email,
		Avatar:      user.Avatar,
		UserType:    user.UserType,
	}
	accessToken, refreshToken, accessExpiresIn, refreshExpiresIn, err := h.tokenManager.GenerateTokenPair(ctx, sysUser, client)
	if err != nil {
		h.ctr.GetLogger().Error("failed to generate token pair", zap.Error(err))
		h.recordLoginLog(c, user.Username, client.ClientId, 1, "生成Token失败")
		response.FailCode(c, response.CodeServerError, "生成Token失败")
		return
	}

	err = h.concurrentLoginManager.RecordLogin(ctx, user.UserId, client.ClientId, accessToken, client.Timeout)
	if err != nil {
		h.ctr.GetLogger().Warn("record login failed", zap.Error(err))
	}

	h.ctr.GetLogger().Info("login success",
		zap.String("clientId", client.ClientId),
		zap.String("grantType", req.GrantType),
		zap.Int64("userId", user.UserId))
	h.recordLoginLog(c, user.Username, client.ClientId, 0, "登录成功")

	response.Success(c, buildLoginResponse(accessToken, refreshToken, accessExpiresIn, refreshExpiresIn, client, user))
}

func (h *authController) authenticateClient(ctx context.Context, req *LoginRequest) (*model.AuthClient, error) {
	return h.clientService.AuthenticateClientID(ctx, req.ClientID, req.GrantType)
}

func buildLoginResponse(accessToken, refreshToken string, accessExpiresIn, refreshExpiresIn int64, client *model.AuthClient, user *UserInfo) *LoginResponse {
	resp := &LoginResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        accessExpiresIn,
		RefreshExpiresIn: refreshExpiresIn,
		ExpireIn:         accessExpiresIn,
		RefreshExpireIn:  refreshExpiresIn,
		UserInfo:         user,
	}
	if client != nil {
		resp.ClientID = client.ClientId
	}
	if user != nil {
		resp.OpenID = user.OpenID
	}
	return resp
}

// Logout godoc
//
//	@Summary		用户登出
//	@Description	使当前用户的 RefreshToken 失效，AccessToken 会在过期后自动失效
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	response.Response{data=string}
//	@Router			/logout [post]
func (h *authController) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	tokenHeader := h.ctr.GetConfig().Auth.TokenHeader
	token := c.GetHeader(tokenHeader)
	token = strings.TrimPrefix(token, "Bearer ")

	if token == "" {
		response.Success(c, "ok")
		return
	}

	claims, err := h.ctr.GetJWT().ValidateToken(token)
	if err != nil {
		response.Success(c, "ok")
		return
	}

	_ = h.tokenManager.InvalidateToken(ctx, claims.UserId, claims.ClientId)

	if h.ctr.GetConfig().Auth.ShareToken {
		_ = h.concurrentLoginManager.InvalidateUserTokens(ctx, claims.UserId, claims.ClientId)
	}

	h.ctr.GetLogger().Info("logout success",
		zap.Int64("userId", claims.UserId),
		zap.String("clientId", claims.ClientId))

	response.Success(c, "ok")
}

// RefreshToken godoc
//
//	@Summary		刷新访问令牌
//	@Description	使用 RefreshToken 获取新的 AccessToken 和 RefreshToken（轮换机制）
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			request	body		response.RefreshTokenRequest	true	"刷新令牌请求参数"
//	@Success		200		{object}	response.Response{data=response.RefreshTokenResponse}
//	@Failure		400		{object}	response.Response	"参数错误"
//	@Failure		401		{object}	response.Response	"RefreshToken 无效或已过期"
//	@Router			/auth/refresh [post]
func (h *authController) RefreshToken(c *gin.Context) {
	var req response.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}
	ctx := c.Request.Context()

	client, err := h.clientService.AuthenticateClientID(ctx, req.ClientID, "refresh")
	if err != nil {
		h.ctr.GetLogger().Error("client authentication failed",
			zap.String("clientId", req.ClientID),
			zap.Error(err))
		response.FailCode(c, response.CodeUnauthorized, err.Error())
		return
	}

	newAccessToken, newRefreshToken, accessExpiresIn, refreshExpiresIn, err := h.tokenManager.RefreshAccessToken(ctx, req.RefreshToken, client)
	if err != nil {
		h.ctr.GetLogger().Error("refresh token failed",
			zap.String("clientId", client.ClientId),
			zap.Error(err))
		response.FailCode(c, response.CodeUnauthorized, err.Error())
		return
	}

	h.ctr.GetLogger().Info("refresh token success",
		zap.String("clientId", client.ClientId))

	response.Success(c, &response.RefreshTokenResponse{
		AccessToken:      newAccessToken,
		RefreshToken:     newRefreshToken,
		ExpiresIn:        accessExpiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	})
}

func (h *authController) recordLoginLog(c *gin.Context, username, clientId string, status int32, message string) {
	browser, osName := parseUserAgent(c.Request.UserAgent())
	logEntry := &model.LoginLog{
		ID:        idgen.MustNextID(),
		UserName:  username,
		Ipaddr:    utils.GetClientIP(c),
		Browser:   browser,
		Os:        osName,
		Status:    status,
		Msg:       message,
		LoginTime: utils.Now(),
		ClientId:  clientId,
	}
	if err := logEntry.Create(h.ctr.GetDB()); err != nil {
		h.ctr.GetLogger().Error("failed to record login log", zap.Error(err))
	}
}

func parseUserAgent(ua string) (browser, os string) {
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(lower, "chrome"):
		browser = "Chrome"
	case strings.Contains(lower, "safari"):
		browser = "Safari"
	case strings.Contains(lower, "firefox"):
		browser = "Firefox"
	case strings.Contains(lower, "edge"):
		browser = "Edge"
	case strings.Contains(lower, "msie") || strings.Contains(lower, "trident"):
		browser = "IE"
	default:
		browser = "Unknown"
	}
	switch {
	case strings.Contains(lower, "windows"):
		os = "Windows"
	case strings.Contains(lower, "mac os") || strings.Contains(lower, "macos"):
		os = "macOS"
	case strings.Contains(lower, "android"):
		os = "Android"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ios"):
		os = "iOS"
	case strings.Contains(lower, "linux"):
		os = "Linux"
	default:
		os = "Unknown"
	}
	return
}

func resolveLoginAccount(req *LoginRequest) string {
	if req.Username != "" {
		return req.Username
	}
	if req.Phonenumber != "" {
		return req.Phonenumber
	}
	if req.Email != "" {
		return req.Email
	}
	return "-"
}

// LoginRequest 定义业务数据结构。
type LoginRequest = request.LoginRequest

// LoginResponse 定义业务数据结构。
type LoginResponse = response.LoginResponse

// UserInfo 定义业务数据结构。
type UserInfo = response.UserInfo

// IAuthStrategy 定义业务数据结构。
type IAuthStrategy interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	GrantType() string
}

// StrategyFactory 定义业务数据结构。
type StrategyFactory struct {
	strategies map[string]IAuthStrategy
}

// NewStrategyFactory 创建组件实例。
func NewStrategyFactory() *StrategyFactory {
	return &StrategyFactory{strategies: make(map[string]IAuthStrategy)}
}

// Register 执行业务逻辑。
func (f *StrategyFactory) Register(strategy IAuthStrategy) {
	f.strategies[strategy.GrantType()] = strategy
}

// GetStrategy 获取业务数据。
func (f *StrategyFactory) GetStrategy(grantType string) (IAuthStrategy, error) {
	s, ok := f.strategies[grantType]
	if !ok {
		return nil, fmt.Errorf("不支持的授权类型: %s", grantType)
	}
	return s, nil
}

// Login 执行业务逻辑。
func (f *StrategyFactory) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	s, err := f.GetStrategy(req.GrantType)
	if err != nil {
		return nil, err
	}
	return s.Login(ctx, req)
}

// CaptchaService 定义业务数据结构。
type CaptchaService struct{ redis *redis.Client }

// NewCaptchaService 创建组件实例。
func NewCaptchaService(r *redis.Client) *CaptchaService { return &CaptchaService{redis: r} }

// CaptchaCodeKey 定义业务配置值。
const CaptchaCodeKey = "global:captcha_codes:"

// ValidateCaptcha 执行业务逻辑。
func (s *CaptchaService) ValidateCaptcha(ctx context.Context, uuid, code string) error {
	if uuid == "" || code == "" {
		return fmt.Errorf("验证码不能为空")
	}
	key := CaptchaCodeKey + uuid
	saved, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("验证码已过期")
		}
		return fmt.Errorf("验证码验证失败")
	}
	defer s.redis.Del(ctx, key)
	if !strings.EqualFold(saved, code) {
		return fmt.Errorf("验证码错误")
	}
	return nil
}

// ValidateSmsCode 执行业务逻辑。
func (s *CaptchaService) ValidateSmsCode(ctx context.Context, phonenumber, code string) error {
	if phonenumber == "" || code == "" {
		return fmt.Errorf("手机号和验证码不能为空")
	}
	key := CaptchaCodeKey + phonenumber
	saved, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("验证码已过期")
		}
		return fmt.Errorf("验证码验证失败")
	}
	defer s.redis.Del(ctx, key)
	if saved != code {
		return fmt.Errorf("验证码错误")
	}
	return nil
}

// ValidateEmailCode 执行业务逻辑。
func (s *CaptchaService) ValidateEmailCode(ctx context.Context, email, code string) error {
	if email == "" || code == "" {
		return fmt.Errorf("邮箱和验证码不能为空")
	}
	key := CaptchaCodeKey + "email:" + email
	saved, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("验证码已过期")
		}
		return fmt.Errorf("验证码验证失败")
	}
	defer s.redis.Del(ctx, key)
	if saved != code {
		return fmt.Errorf("验证码错误")
	}
	return nil
}

// GenerateEmailCode 执行业务逻辑。
func (s *CaptchaService) GenerateEmailCode(ctx context.Context, email string) (string, error) {
	if email == "" {
		return "", fmt.Errorf("邮箱不能为空")
	}
	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	key := CaptchaCodeKey + "email:" + email
	err := s.redis.Set(ctx, key, code, 5*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("存储验证码失败: %w", err)
	}
	return code, nil
}

// PasswordAuthStrategy 定义业务数据结构。
type PasswordAuthStrategy struct {
	ctr container.Container
}

// NewPasswordAuthStrategy 创建组件实例。
func NewPasswordAuthStrategy(c container.Container) *PasswordAuthStrategy {
	return &PasswordAuthStrategy{ctr: c}
}

// GrantType 执行业务逻辑。
func (s *PasswordAuthStrategy) GrantType() string { return "password" }

// Login 执行业务逻辑。
func (s *PasswordAuthStrategy) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("用户名和密码不能为空")
	}

	if s.ctr.GetConfig().Captcha.Image.Enabled {
		if req.Uuid != "" && req.Code != "" {
			if err := NewCaptchaService(s.ctr.GetRedis()).ValidateCaptcha(ctx, req.Uuid, req.Code); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("请输入图形验证码")
		}
	}

	if err := s.checkBruteForce(ctx, req.Username); err != nil {
		return nil, err
	}
	var um model.User
	user, err := um.FindByUsername(s.ctr.GetDB(), req.Username)
	if err != nil {
		s.incrementErrorCount(ctx, req.Username)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("用户名或密码错误")
		}
		s.ctr.GetLogger().Error("failed to query user", zap.Error(err))
		return nil, fmt.Errorf("登录失败")
	}
	if err := utils.VerifyPassword(user.Password, req.Password); err != nil {
		s.incrementErrorCount(ctx, req.Username)
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if user.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	s.clearErrorCount(ctx, req.Username)

	var clientModel model.AuthClient
	client, err := clientModel.FindByClientId(s.ctr.GetDB(), req.ClientID)
	if err != nil {
		s.ctr.GetLogger().Error("failed to query client", zap.Error(err))
		return nil, fmt.Errorf("客户端配置查询失败")
	}
	token, expiresIn, err := s.ctr.GetJWT().GenerateToken(
		user.ID,
		user.UserName,
		client.ClientId,
		client.DeviceType,
		client.Timeout,
	)
	if err != nil {
		s.ctr.GetLogger().Error("failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成token失败")
	}
	if client.ActiveTimeout > 0 {
		tokenHash := generateTokenHash(token)
		activeKey := "token:active:" + tokenHash
		_ = s.ctr.GetRedis().Set(ctx, activeKey, time.Now().Unix(), time.Duration(client.ActiveTimeout)*time.Second).Err()
	}
	return &LoginResponse{AccessToken: token, ExpiresIn: expiresIn, UserInfo: newLoginUserInfo(user)}, nil
}
func (s *PasswordAuthStrategy) checkBruteForce(ctx context.Context, username string) error {
	key := "pwd_err_cnt:" + username
	count, err := s.ctr.GetRedis().Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		s.ctr.GetLogger().Warn("failed to get password error count", zap.Error(err))
		return nil
	}
	if count >= 5 {
		ttl, _ := s.ctr.GetRedis().TTL(ctx, key).Result()
		return fmt.Errorf("密码错误次数过多，请%d分钟后再试", int(ttl.Minutes())+1)
	}
	return nil
}
func (s *PasswordAuthStrategy) incrementErrorCount(ctx context.Context, username string) {
	key := "pwd_err_cnt:" + username
	pipe := s.ctr.GetRedis().Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 10*time.Minute)
	_, _ = pipe.Exec(ctx)
}
func (s *PasswordAuthStrategy) clearErrorCount(ctx context.Context, username string) {
	_ = s.ctr.GetRedis().Del(ctx, "pwd_err_cnt:"+username).Err()
}

// XcxAuthStrategy 定义业务数据结构。
type XcxAuthStrategy struct {
	ctr container.Container
}

const (
	miniProgramUserType     int32 = 1
	defaultMiniProgramOrgID int64 = 1880159541355577346
)

// NewXcxAuthStrategy 创建组件实例。
func NewXcxAuthStrategy(c container.Container) *XcxAuthStrategy {
	return &XcxAuthStrategy{ctr: c}
}

// GrantType 执行业务逻辑。
func (s *XcxAuthStrategy) GrantType() string { return "xcx" }

// Login 执行业务逻辑。
func (s *XcxAuthStrategy) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if req.Phonenumber == "" || req.Code == "" || req.WxCode == "" {
		return nil, fmt.Errorf("手机号、验证码和微信code不能为空")
	}
	//if err := NewCaptchaService(s.ctr.GetRedis()).ValidateSmsCode(ctx, req.Phonenumber, req.Code); err != nil {
	//	return nil, err
	//}
	if !s.ctr.GetConfig().WeChat.Enabled {
		return nil, fmt.Errorf("微信小程序登录未启用")
	}
	wxResp, err := s.ctr.GetWeChat().Code2Session(req.WxCode)
	if err != nil {
		return nil, err
	}
	if wxResp.OpenID == "" {
		return nil, fmt.Errorf("获取微信OpenID失败")
	}
	var um model.User
	user, err := um.FindByPhonenumber(s.ctr.GetDB(), req.Phonenumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.ctr.GetLogger().Error("failed to query user by phonenumber", zap.Error(err))
		return nil, fmt.Errorf("查询用户失败")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		loginTime := time.Now().Unix()
		newUser := &model.User{
			UserName:    req.Phonenumber,
			NickName:    req.Phonenumber,
			UserType:    miniProgramUserType,
			OrgID:       defaultMiniProgramOrgID,
			Phonenumber: req.Phonenumber,
			Sex:         2,
			Status:      0,
			OpenId:      wxResp.OpenID,
			UnionId:     wxResp.UnionID,
			LoginIp:     req.LoginIP,
			LoginDate:   loginTime,
		}
		if err := um.Create(s.ctr.GetDB(), newUser); err != nil {
			s.ctr.GetLogger().Error("failed to create wechat user", zap.Error(err))
			return nil, fmt.Errorf("创建用户失败")
		}
		user = newUser
	} else {
		if err := s.ctr.GetDB().Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{"open_id": wxResp.OpenID, "union_id": wxResp.UnionID}).Error; err != nil {
			s.ctr.GetLogger().Warn("failed to update user openid", zap.Error(err))
		}
		user.OpenId = wxResp.OpenID
		user.UnionId = wxResp.UnionID
	}
	if user.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	var clientModel model.AuthClient
	client, err := clientModel.FindByClientId(s.ctr.GetDB(), req.ClientID)
	if err != nil {
		s.ctr.GetLogger().Error("failed to query client", zap.Error(err))
		return nil, fmt.Errorf("客户端配置查询失败")
	}
	token, expiresIn, err := s.ctr.GetJWT().GenerateToken(
		user.ID,
		user.UserName,
		client.ClientId,
		client.DeviceType,
		client.Timeout,
	)
	if err != nil {
		s.ctr.GetLogger().Error("failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成token失败")
	}
	_ = um.UpdateLoginInfo(s.ctr.GetDB(), user.ID, req.LoginIP, time.Now().Unix())
	if client.ActiveTimeout > 0 {
		tokenHash := generateTokenHash(token)
		activeKey := "token:active:" + tokenHash
		_ = s.ctr.GetRedis().Set(ctx, activeKey, time.Now().Unix(), time.Duration(client.ActiveTimeout)*time.Second).Err()
	}
	return &LoginResponse{AccessToken: token, ExpiresIn: expiresIn, UserInfo: newLoginUserInfo(user)}, nil
}

// WechatOnlyAuthStrategy 纯微信小程序登录策略（仅需微信code，自动注册）
type WechatOnlyAuthStrategy struct {
	ctr container.Container
}

// NewWechatOnlyAuthStrategy 创建组件实例。
func NewWechatOnlyAuthStrategy(c container.Container) *WechatOnlyAuthStrategy {
	return &WechatOnlyAuthStrategy{ctr: c}
}

// GrantType 执行业务逻辑。
func (s *WechatOnlyAuthStrategy) GrantType() string { return "wechat" }

// Login 执行业务逻辑。
func (s *WechatOnlyAuthStrategy) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if req.WxCode == "" {
		return nil, fmt.Errorf("微信code不能为空")
	}
	wechatManager := s.ctr.GetWeChat()
	if wechatManager == nil {
		return nil, fmt.Errorf("微信小程序登录未启用")
	}
	wxResp, err := wechatManager.Code2Session(req.WxCode)
	if err != nil {
		return nil, err
	}
	if wxResp.OpenID == "" {
		return nil, fmt.Errorf("获取微信OpenID失败")
	}
	var um model.User
	user, err := um.FindByOpenId(s.ctr.GetDB(), wxResp.OpenID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.ctr.GetLogger().Error("failed to query user by openid", zap.Error(err))
		return nil, fmt.Errorf("查询用户失败")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		loginTime := time.Now().Unix()
		// 自动创建用户
		newUser := &model.User{
			UserName:  "wx_" + wxResp.OpenID[:16],
			NickName:  "微信用户",
			UserType:  miniProgramUserType,
			OrgID:     defaultMiniProgramOrgID,
			Status:    0,
			OpenId:    wxResp.OpenID,
			UnionId:   wxResp.UnionID,
			Sex:       2,
			LoginIp:   req.LoginIP,
			LoginDate: loginTime,
		}
		if err := um.Create(s.ctr.GetDB(), newUser); err != nil {
			s.ctr.GetLogger().Error("failed to create wechat user", zap.Error(err))
			return nil, fmt.Errorf("创建用户失败")
		}
		user = newUser
		s.ctr.GetLogger().Info("wechat user auto registered",
			zap.Int64("userId", user.ID),
			zap.String("openId", wxResp.OpenID))
	} else {
		// 更新UnionId（如果有）
		if wxResp.UnionID != "" && user.UnionId != wxResp.UnionID {
			if err := s.ctr.GetDB().Model(&model.User{}).Where("id = ?", user.ID).Update("union_id", wxResp.UnionID).Error; err != nil {
				s.ctr.GetLogger().Warn("failed to update user unionid", zap.Error(err))
			}
		}
	}
	if user.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}
	var clientModel model.AuthClient
	client, err := clientModel.FindByClientId(s.ctr.GetDB(), req.ClientID)
	if err != nil {
		s.ctr.GetLogger().Error("failed to query client", zap.Error(err))
		return nil, fmt.Errorf("客户端配置查询失败")
	}
	token, expiresIn, err := s.ctr.GetJWT().GenerateToken(
		user.ID,
		user.UserName,
		client.ClientId,
		client.DeviceType,
		client.Timeout,
	)
	if err != nil {
		s.ctr.GetLogger().Error("failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成token失败")
	}
	_ = um.UpdateLoginInfo(s.ctr.GetDB(), user.ID, req.LoginIP, time.Now().Unix())
	if client.ActiveTimeout > 0 {
		tokenHash := generateTokenHash(token)
		activeKey := "token:active:" + tokenHash
		_ = s.ctr.GetRedis().Set(ctx, activeKey, time.Now().Unix(), time.Duration(client.ActiveTimeout)*time.Second).Err()
	}
	return &LoginResponse{AccessToken: token, ExpiresIn: expiresIn, UserInfo: newLoginUserInfo(user)}, nil
}

// EmailAuthStrategy 定义业务数据结构。
type EmailAuthStrategy struct {
	ctr container.Container
}

// NewEmailAuthStrategy 创建组件实例。
func NewEmailAuthStrategy(c container.Container) *EmailAuthStrategy {
	return &EmailAuthStrategy{ctr: c}
}

// GrantType 执行业务逻辑。
func (s *EmailAuthStrategy) GrantType() string { return "email" }

// Login 执行业务逻辑。
func (s *EmailAuthStrategy) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Code == "" {
		return nil, fmt.Errorf("邮箱和验证码不能为空")
	}

	if err := NewCaptchaService(s.ctr.GetRedis()).ValidateEmailCode(ctx, req.Email, req.Code); err != nil {
		return nil, err
	}

	var um model.User
	user, err := um.FindByEmail(s.ctr.GetDB(), req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("邮箱或验证码错误")
		}
		s.ctr.GetLogger().Error("failed to query user by email", zap.Error(err))
		return nil, fmt.Errorf("查询用户失败")
	}

	if user.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}

	var clientModel model.AuthClient
	client, err := clientModel.FindByClientId(s.ctr.GetDB(), req.ClientID)
	if err != nil {
		s.ctr.GetLogger().Error("failed to query client", zap.Error(err))
		return nil, fmt.Errorf("客户端配置查询失败")
	}

	token, expiresIn, err := s.ctr.GetJWT().GenerateToken(
		user.ID,
		user.UserName,
		client.ClientId,
		client.DeviceType,
		client.Timeout,
	)
	if err != nil {
		s.ctr.GetLogger().Error("failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成token失败")
	}

	if client.ActiveTimeout > 0 {
		tokenHash := generateTokenHash(token)
		activeKey := "token:active:" + tokenHash
		_ = s.ctr.GetRedis().Set(ctx, activeKey, time.Now().Unix(), time.Duration(client.ActiveTimeout)*time.Second).Err()
	}

	return &LoginResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
		UserInfo:    newLoginUserInfo(user),
	}, nil
}

// SmsAuthStrategy 短信验证码登录策略
type SmsAuthStrategy struct {
	ctr container.Container
}

// NewSmsAuthStrategy 创建组件实例。
func NewSmsAuthStrategy(c container.Container) *SmsAuthStrategy {
	return &SmsAuthStrategy{ctr: c}
}

// GrantType 执行业务逻辑。
func (s *SmsAuthStrategy) GrantType() string { return "sms" }

// Login 执行业务逻辑。
func (s *SmsAuthStrategy) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	if req.Phonenumber == "" || req.Code == "" {
		return nil, fmt.Errorf("手机号和验证码不能为空")
	}

	if err := NewCaptchaService(s.ctr.GetRedis()).ValidateSmsCode(ctx, req.Phonenumber, req.Code); err != nil {
		return nil, err
	}

	var um model.User
	user, err := um.FindByPhonenumber(s.ctr.GetDB(), req.Phonenumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("手机号或验证码错误")
		}
		s.ctr.GetLogger().Error("failed to query user by phonenumber", zap.Error(err))
		return nil, fmt.Errorf("查询用户失败")
	}

	if user.Status != 0 {
		return nil, fmt.Errorf("用户已被停用")
	}

	var clientModel model.AuthClient
	client, err := clientModel.FindByClientId(s.ctr.GetDB(), req.ClientID)
	if err != nil {
		s.ctr.GetLogger().Error("failed to query client", zap.Error(err))
		return nil, fmt.Errorf("客户端配置查询失败")
	}

	token, expiresIn, err := s.ctr.GetJWT().GenerateToken(
		user.ID,
		user.UserName,
		client.ClientId,
		client.DeviceType,
		client.Timeout,
	)
	if err != nil {
		s.ctr.GetLogger().Error("failed to generate token", zap.Error(err))
		return nil, fmt.Errorf("生成token失败")
	}

	if client.ActiveTimeout > 0 {
		tokenHash := generateTokenHash(token)
		activeKey := "token:active:" + tokenHash
		_ = s.ctr.GetRedis().Set(ctx, activeKey, time.Now().Unix(), time.Duration(client.ActiveTimeout)*time.Second).Err()
	}

	return &LoginResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
		UserInfo:    newLoginUserInfo(user),
	}, nil
}

func newLoginUserInfo(user *model.User) *UserInfo {
	if user == nil {
		return nil
	}
	return &UserInfo{
		UserId:      user.ID,
		Username:    user.UserName,
		Nickname:    user.NickName,
		Phonenumber: user.Phonenumber,
		Email:       user.Email,
		Avatar:      user.Avatar,
		OrgID:       user.OrgID,
		UserType:    user.UserType,
		OpenID:      user.OpenId,
		UnionID:     user.UnionId,
	}
}

func generateTokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
