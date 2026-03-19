package biz

import "github.com/force-c/nai-tizi/kratos/application/sys-api/internal/data"

type Usecases struct {
	Health     *HealthUsecase
	Auth       *AuthUsecase
	Captcha    *CaptchaUsecase
	Menu       *MenuUsecase
	User       *UserUsecase
	Role       *RoleUsecase
	Org        *OrgUsecase
	Config     *ConfigUsecase
	Dict       *DictUsecase
	LoginLog   *LoginLogUsecase
	OperLog    *OperLogUsecase
	StorageEnv *StorageEnvUsecase
	Attachment *AttachmentUsecase
}

func NewUsecases(repos *data.Repositories) *Usecases {
	return &Usecases{
		Health:     NewHealthUsecase(repos.Health),
		Auth:       NewAuthUsecase(repos.Auth),
		Captcha:    NewCaptchaUsecase(repos.Captcha),
		Menu:       NewMenuUsecase(repos.Menu),
		User:       NewUserUsecase(repos.User),
		Role:       NewRoleUsecase(repos.Role),
		Org:        NewOrgUsecase(repos.Org),
		Config:     NewConfigUsecase(repos.Config),
		Dict:       NewDictUsecase(repos.Dict),
		LoginLog:   NewLoginLogUsecase(repos.LoginLog),
		OperLog:    NewOperLogUsecase(repos.OperLog),
		StorageEnv: NewStorageEnvUsecase(repos.StorageEnv),
		Attachment: NewAttachmentUsecase(repos.Attachment),
	}
}
