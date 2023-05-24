package config

// Settings contains the application config
type Settings struct {
	Environment string `yaml:"ENVIRONMENT"`
	Port        string `yaml:"PORT"`
	MonPort     string `yaml:"MON_PORT"`
}
