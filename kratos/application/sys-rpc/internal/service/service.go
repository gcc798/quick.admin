package service

import "github.com/google/wire"

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewHealthServiceService,
	NewAuthServiceService,
	NewCaptchaServiceService,
	NewMenuServiceService,
	NewUserServiceService,
	NewRoleServiceService,
	NewOrgServiceService,
	NewConfigServiceService,
	NewDictServiceService,
	NewLoginLogServiceService,
	NewOperLogServiceService,
	NewStorageEnvServiceService,
	NewAttachmentServiceService,
)
