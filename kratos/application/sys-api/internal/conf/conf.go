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
	Mode      string       `yaml:"mode"`
	Endpoint  string       `yaml:"endpoint"`
	Timeout   string       `yaml:"timeout"`
	Discovery RPCDiscovery `yaml:"discovery"`
}

type RPCDiscovery struct {
	Driver  string       `yaml:"driver"`
	Service string       `yaml:"service"`
	Etcd    RegistryEtcd `yaml:"etcd"`
}

type RegistryEtcd struct {
	Endpoints []string `yaml:"endpoints"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Namespace string   `yaml:"namespace"`
	Timeout   string   `yaml:"timeout"`
}

type Data struct {
	Redis Redis `yaml:"redis"`
}

type Redis struct {
	Addr string `yaml:"addr"`
}
