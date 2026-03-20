package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/force-c/nai-tizi/internal/domain/model"
	"github.com/force-c/nai-tizi/internal/infrastructure/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	PasswordLogin(req *PasswordLoginReq) (*LoginResp, error)
	XcxLogin(req *XcxLoginReq) (*LoginResp, error)
	UpdateLogin(userId int64, ip string) error
	GetProfile(userId int64) (*model.User, error)
}

type authService struct {
	db       *gorm.DB
	jwt      *jwt.Jwt
	wxAppID  string
	wxSecret string
}

func NewAuthService(db *gorm.DB, jwtService *jwt.Jwt, wxAppID, wxSecret string) AuthService {
	return &authService{
		db:       db,
		jwt:      jwtService,
		wxAppID:  wxAppID,
		wxSecret: wxSecret,
	}
}

type PasswordLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResp struct {
	AccessToken string `json:"accessToken"`
	ExpireIn    int64  `json:"expireIn"`
	UserId      int64  `json:"userId"`
}

func (s *authService) PasswordLogin(req *PasswordLoginReq) (*LoginResp, error) {
	u, err := (&model.User{}).FindByUsername(s.db, req.Username)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return nil, errors.New("password_wrong")
	}
	token, expireIn, err := s.jwt.GenerateToken(u.ID, u.UserName, "", "")
	if err != nil {
		return nil, err
	}
	return &LoginResp{AccessToken: token, ExpireIn: expireIn, UserId: u.ID}, nil
}

type XcxLoginReq struct {
	Code     string `json:"code"`
	NickName string `json:"nickName"`
	Phone    string `json:"phone"`
}

func (s *authService) XcxLogin(req *XcxLoginReq) (*LoginResp, error) {
	if s.wxAppID == "" || s.wxSecret == "" {
		return nil, errors.New("wechat_config_missing")
	}
	url := "https://api.weixin.qq.com/sns/jscode2session?appid=" + s.wxAppID + "&secret=" + s.wxSecret + "&js_code=" + req.Code + "&grant_type=authorization_code"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r struct {
		OpenID  string `json:"openid"`
		UnionID string `json:"unionid"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.ErrCode != 0 || r.OpenID == "" {
		return nil, errors.New("wechat_code2session_failed")
	}
	u, err := (&model.User{}).FindByOpenId(s.db, r.OpenID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		nu := &model.User{OpenId: r.OpenID, UnionId: r.UnionID, UserName: "wx_" + r.OpenID, NickName: req.NickName}
		if e := (&model.User{}).Create(s.db, nu); e != nil {
			return nil, e
		}
		u = nu
	} else if err != nil {
		return nil, err
	}
	token, expireIn, err := s.jwt.GenerateToken(u.ID, u.UserName, "", "")
	if err != nil {
		return nil, err
	}
	return &LoginResp{AccessToken: token, ExpireIn: expireIn, UserId: u.ID}, nil
}

func (s *authService) UpdateLogin(userId int64, ip string) error {
	return (&model.User{}).UpdateLoginInfo(s.db, userId, ip, time.Now().Unix())
}
func (s *authService) GetProfile(userId int64) (*model.User, error) {
	var u model.User
	tx := s.db.Where("user_id = ?", userId).Limit(1).Find(&u)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &u, nil
}
