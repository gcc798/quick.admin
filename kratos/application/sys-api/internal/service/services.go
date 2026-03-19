package service

import "github.com/force-c/nai-tizi/kratos/application/sys-api/internal/biz"

type Services struct {
	Health     *HealthServiceService
	Auth       *AuthServiceService
	Captcha    *CaptchaServiceService
	Menu       *MenuServiceService
	User       *UserServiceService
	Role       *RoleServiceService
	Org        *OrgServiceService
	Config     *ConfigServiceService
	Dict       *DictServiceService
	LoginLog   *LoginLogServiceService
	OperLog    *OperLogServiceService
	StorageEnv *StorageEnvServiceService
	Attachment *AttachmentServiceService
}

func NewServices(uc *biz.Usecases) *Services {
	return &Services{
		Health:     NewHealthServiceService(uc.Health),
		Auth:       NewAuthServiceService(uc.Auth),
		Captcha:    NewCaptchaServiceService(uc.Captcha),
		Menu:       NewMenuServiceService(uc.Menu),
		User:       NewUserServiceService(uc.User),
		Role:       NewRoleServiceService(uc.Role),
		Org:        NewOrgServiceService(uc.Org),
		Config:     NewConfigServiceService(uc.Config),
		Dict:       NewDictServiceService(uc.Dict),
		LoginLog:   NewLoginLogServiceService(uc.LoginLog),
		OperLog:    NewOperLogServiceService(uc.OperLog),
		StorageEnv: NewStorageEnvServiceService(uc.StorageEnv),
		Attachment: NewAttachmentServiceService(uc.Attachment),
	}
}
