package config

// Settings contains the application config
type Settings struct {
	LogLevel string `yaml:"LOG_LEVEL"`

	Environment string `yaml:"ENVIRONMENT"`
	MonPort     string `yaml:"MON_PORT"`
	GRPCPort    string `yaml:"GRPC_PORT"`

	BIP32Seed string `yaml:"BIP32_SEED"`

	EnclaveCID  string `yaml:"ENCLAVE_CID"`
	EnclavePort string `yaml:"ENCLAVE_PORT"`

	MockEnclave bool   `yaml:"MOCK_ENCLAVE"`
	MockSeed    string `yaml:"MOCK_SEED"`
}
