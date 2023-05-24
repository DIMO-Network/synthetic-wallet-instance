package config

// Settings contains the application config
type Settings struct {
	Environment string `yaml:"ENVIRONMENT"`
	Port        string `yaml:"PORT"`
	MonPort     string `yaml:"MON_PORT"`
	GRPCPort    string `yaml:"GRPC_PORT"`

	BIP32Seed string `json:"BIP32_SEED"`

	EnclaveCID  int `json:"ENCLAVE_CID"`
	EnclavePort int `json:"ENCLAVE_PORT"`
}
