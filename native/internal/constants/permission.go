package constants

// 权限资源常量
const (
	// 角色管理
	ResourceRole           = "role"
	ResourceRoleRead       = "role.read"
	ResourceRoleCreate     = "role.create"
	ResourceRoleUpdate     = "role.update"
	ResourceRoleDelete     = "role.delete"
	ResourceRoleAssign     = "role.assign"
	ResourceRolePermission = "role.permission"

	// 用户管理
	ResourceUser       = "user"
	ResourceUserRead   = "user.read"
	ResourceUserCreate = "user.create"
	ResourceUserUpdate = "user.update"
	ResourceUserDelete = "user.delete"

	// 组织管理
	ResourceOrg       = "org"
	ResourceOrgRead   = "org.read"
	ResourceOrgCreate = "org.create"
	ResourceOrgUpdate = "org.update"
	ResourceOrgDelete = "org.delete"

	// 菜单管理
	ResourceMenu       = "menu"
	ResourceMenuRead   = "menu.read"
	ResourceMenuCreate = "menu.create"
	ResourceMenuUpdate = "menu.update"
	ResourceMenuDelete = "menu.delete"

	// 字典管理
	ResourceDict       = "dict"
	ResourceDictRead   = "dict.read"
	ResourceDictCreate = "dict.create"
	ResourceDictUpdate = "dict.update"
	ResourceDictDelete = "dict.delete"

	// 配置管理
	ResourceConfig       = "config"
	ResourceConfigRead   = "config.read"
	ResourceConfigCreate = "config.create"
	ResourceConfigUpdate = "config.update"
	ResourceConfigDelete = "config.delete"

	// 登录日志管理
	ResourceLoginLog       = "login_log"
	ResourceLoginLogRead   = "login_log.read"
	ResourceLoginLogCreate = "login_log.create"
	ResourceLoginLogUpdate = "login_log.update"
	ResourceLoginLogDelete = "login_log.delete"

	// 操作日志管理
	ResourceOperLog       = "oper_log"
	ResourceOperLogRead   = "oper_log.read"
	ResourceOperLogCreate = "oper_log.create"
	ResourceOperLogUpdate = "oper_log.update"
	ResourceOperLogDelete = "oper_log.delete"

	// 存储环境管理
	ResourceStorageEnv       = "storage_env"
	ResourceStorageEnvRead   = "storage_env.read"
	ResourceStorageEnvCreate = "storage_env.create"
	ResourceStorageEnvUpdate = "storage_env.update"
	ResourceStorageEnvDelete = "storage_env.delete"
	ResourceStorageEnvManage = "storage_env.manage"

	// 附件管理
	ResourceAttachment         = "attachment"
	ResourceAttachmentRead     = "attachment.read"
	ResourceAttachmentCreate   = "attachment.create"
	ResourceAttachmentUpdate   = "attachment.update"
	ResourceAttachmentDelete   = "attachment.delete"
	ResourceAttachmentUpload   = "attachment.upload"
	ResourceAttachmentDownload = "attachment.download"
	ResourceAttachmentBind     = "attachment.bind"
)
