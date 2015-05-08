package highbatch

import (
	"github.com/BurntSushi/toml"
	"os"
)

var Conf Config

type Config struct {
	Master MasterConfig
	Worker WorkerConfig
}

type MasterConfig struct {
	Host string
	Port string
}

type WorkerConfig struct {
	Host     string
	Port     string
	LogLevel int
	IsMaster bool
}

func loadConfig() (c Config) {
	if _, err := toml.DecodeFile("config.toml", &Conf); err != nil {
		le(err)
	}
	c = Conf

	if os.Getenv("HighBatchIsMaster") != "" {
		c.Worker.IsMaster = true
	}

	return
}
