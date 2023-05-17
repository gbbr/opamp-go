package config

// Supervisor is the Supervisor config file format.
type Supervisor struct {
	Server *OpAMPServer
	Agent  *Agent
}

type OpAMPServer struct {
	Endpoint string
}

type Agent struct {
	Executable  string
	LocalConfig string `koanf:"local_config"`
	Type        string
}
