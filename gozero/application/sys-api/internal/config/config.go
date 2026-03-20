// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type PostgresConf struct {
	Dsn string
}

type RedisConf struct {
	Addr     string
	Password string
	Db       int
}

type JwtConf struct {
	Secret string
	Expire int64
}

type AuthConf struct {
	TokenHeader string
}

type CaptchaImageConf struct {
	Enabled bool
}

type CaptchaSmsConf struct {
	Enabled bool
}

type CaptchaEmailConf struct {
	Enabled bool
}

type CaptchaConf struct {
	Image CaptchaImageConf
	Sms   CaptchaSmsConf
	Email CaptchaEmailConf
}

type Config struct {
	rest.RestConf
	SysRpc   zrpc.RpcClientConf
	Postgres PostgresConf
	Redis    RedisConf
	Jwt      JwtConf
	Auth     AuthConf
	Captcha  CaptchaConf
}
