package config

type AppConfig struct {
	HTTP HTTPConfig `yaml:"http"`
}

type HTTPConfig struct {
	Greeting string `yaml:"greeting"`
}

func (c *AppConfig) Print() {}
