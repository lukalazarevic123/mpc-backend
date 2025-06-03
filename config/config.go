package config

type Configuration struct {
	DbConfig   DbConfig
	ServerConf ServerConf
}

type DbConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
	SSLMode  *string
}

type ServerConf struct {
	Host string
	Port int
}
