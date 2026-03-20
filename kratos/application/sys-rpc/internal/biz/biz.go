package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewHealthUsecase,
	NewAuthUsecase,
	NewCaptchaUsecase,
	NewMenuUsecase,
	NewUserUsecase,
	NewRoleUsecase,
	NewOrgUsecase,
	NewConfigUsecase,
	NewDictUsecase,
	NewLoginLogUsecase,
	NewOperLogUsecase,
	NewStorageEnvUsecase,
	NewAttachmentUsecase,
)
