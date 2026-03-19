package biz

import "github.com/force-c/nai-tizi/kratos/application/sys-rpc/internal/data"

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
	LoginLog   *LogUsecase
	OperLog    *LogUsecase
	StorageEnv *StorageEnvUsecase
	Attachment *AttachmentUsecase
}

func NewUsecases(resources *data.Resources) *Usecases {
	return &Usecases{
		Health:     NewHealthUsecase(),
		Auth:       NewAuthUsecase(resources),
		Captcha:    NewCaptchaUsecase(resources),
		Menu:       NewMenuUsecase(resources),
		User:       NewUserUsecase(resources),
		Role:       NewRoleUsecase(resources),
		Org:        NewOrgUsecase(resources),
		Config:     NewConfigUsecase(resources),
		Dict:       NewDictUsecase(resources),
		LoginLog:   NewLogUsecase(resources, "login"),
		OperLog:    NewLogUsecase(resources, "oper"),
		StorageEnv: NewStorageEnvUsecase(resources),
		Attachment: NewAttachmentUsecase(resources),
	}
}
