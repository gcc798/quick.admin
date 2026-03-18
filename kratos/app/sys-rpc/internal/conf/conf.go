package conf

type Bootstrap struct {
	Server Server `yaml:"server"`
	Auth   Auth   `yaml:"auth"`
	Data   Data   `yaml:"data"`
}

type Auth struct {
	AllowConcurrent bool `yaml:"allowConcurrent"`
	ShareToken      bool `yaml:"shareToken"`
}

type Server struct {
	GRPC GRPC `yaml:"grpc"`
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
