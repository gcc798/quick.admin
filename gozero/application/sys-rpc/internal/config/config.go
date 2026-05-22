package config

import "github.com/zeromicro/go-zero/zrpc"

type PostgresConf struct {
	Dsn string
}

type RedisConf struct {
	Addr     string
	Password string
	Db       int
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

type WechatConf struct {
	Enabled bool
	AppId   string
	Secret  string
}

type Config struct {
	zrpc.RpcServerConf
	Postgres   PostgresConf
	CacheRedis RedisConf
	Captcha    CaptchaConf
	Wechat     WechatConf
}
