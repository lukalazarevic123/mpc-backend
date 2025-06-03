package config

import "fmt"

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

func (c *DbConfig) ConnectionString(driver string) string {
	connStr := fmt.Sprintf("%s://%s:%s@%s:%d/%s", driver, c.Username, c.Password, c.Host, c.Port, c.Database)
	if c.SSLMode != nil {
		connStr = connStr + fmt.Sprintf("?sslmode=%s", *c.SSLMode)
	}

	return connStr
}
