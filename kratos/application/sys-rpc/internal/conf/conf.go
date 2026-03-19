package conf

type Bootstrap struct {
	Server        Server        `yaml:"server"`
	Registry      Registry      `yaml:"registry"`
	Auth          Auth          `yaml:"auth"`
	JWT           JWT           `yaml:"jwt"`
	Data          Data          `yaml:"data"`
	Observability Observability `yaml:"observability"`
}

type Auth struct {
	AllowConcurrent bool `yaml:"allowConcurrent"`
	ShareToken      bool `yaml:"shareToken"`
}

type JWT struct {
	Secret string `yaml:"secret"`
	Expire int64  `yaml:"expire"`
	Issuer string `yaml:"issuer"`
}

type Server struct {
	GRPC GRPC `yaml:"grpc"`
}

type Registry struct {
	Mode        string       `yaml:"mode"`
	Driver      string       `yaml:"driver"`
	Service     string       `yaml:"service"`
	Namespace   string       `yaml:"namespace"`
	RegisterTTL string       `yaml:"registerTTL"`
	Etcd        RegistryEtcd `yaml:"etcd"`
}

type RegistryEtcd struct {
	Endpoints []string `yaml:"endpoints"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Timeout   string   `yaml:"timeout"`
}

type GRPC struct {
	Network string `yaml:"network"`
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
}

type Data struct {
	Database Database `yaml:"database"`
	Redis    Redis    `yaml:"redis"`
	WeChat   WeChat   `yaml:"wechat"`
}

type Database struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type WeChat struct {
	AppID   string `yaml:"appId"`
	Secret  string `yaml:"secret"`
	Timeout string `yaml:"timeout"`
}

type Observability struct {
	DBSlowThresholdMs    int64  `yaml:"dbSlowThresholdMs"`
	RedisSlowThresholdMs int64  `yaml:"redisSlowThresholdMs"`
	DBPoolSampleInterval string `yaml:"dbPoolSampleInterval"`
}
