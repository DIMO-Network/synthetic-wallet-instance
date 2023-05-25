package config

// Settings contains the application config
type Settings struct {
	Environment string `yaml:"ENVIRONMENT"`
	Port        string `yaml:"PORT"`
	MonPort     string `yaml:"MON_PORT"`
	GRPCPort    string `yaml:"GRPC_PORT"`

	BIP32Seed string `yaml:"BIP32_SEED"`

	EnclaveCID  int `yaml:"ENCLAVE_CID"`
	EnclavePort int `yaml:"ENCLAVE_PORT"`
}
