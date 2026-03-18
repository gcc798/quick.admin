package conf

type Bootstrap struct {
	Server Server `yaml:"server"`
	Data   Data   `yaml:"data"`
}

type Server struct {
	HTTP HTTP `yaml:"http"`
	RPC  RPC  `yaml:"rpc"`
}

type HTTP struct {
	Network string `yaml:"network"`
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
}

type RPC struct {
	Endpoint string `yaml:"endpoint"`
}

type Data struct {
	Redis Redis `yaml:"redis"`
}

type Redis struct {
	Addr string `yaml:"addr"`
}
