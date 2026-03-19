package data

import "github.com/force-c/nai-tizi/kratos/application/sys-api/internal/conf"

type Repositories struct {
	RPC        *RPCClientSet
	Health     *HealthRepo
	Auth       *AuthRepo
	Captcha    *CaptchaRepo
	Menu       *MenuRepo
	User       *UserRepo
	Role       *RoleRepo
	Org        *OrgRepo
	Config     *ConfigRepo
	Dict       *DictRepo
	LoginLog   *LoginLogRepo
	OperLog    *OperLogRepo
	StorageEnv *StorageEnvRepo
	Attachment *AttachmentRepo
}

func NewRepositories(cfg conf.RPC) (*Repositories, error) {
	clients, err := NewRPCClientSet(cfg)
	if err != nil {
		return nil, err
	}
	return &Repositories{
		RPC:        clients,
		Health:     NewHealthRepo(clients),
		Auth:       NewAuthRepo(clients),
		Captcha:    NewCaptchaRepo(clients),
		Menu:       NewMenuRepo(clients),
		User:       NewUserRepo(clients),
		Role:       NewRoleRepo(clients),
		Org:        NewOrgRepo(clients),
		Config:     NewConfigRepo(clients),
		Dict:       NewDictRepo(clients),
		LoginLog:   NewLoginLogRepo(clients),
		OperLog:    NewOperLogRepo(clients),
		StorageEnv: NewStorageEnvRepo(clients),
		Attachment: NewAttachmentRepo(clients),
	}, nil
}

func (r *Repositories) Close() error {
	if r == nil || r.RPC == nil {
		return nil
	}
	return r.RPC.Close()
}
