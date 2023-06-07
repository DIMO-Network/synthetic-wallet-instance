package config

// Settings contains the application config
type Settings struct {
	Environment string `yaml:"ENVIRONMENT"`
	MonPort     string `yaml:"MON_PORT"`
	GRPCPort    string `yaml:"GRPC_PORT"`

	BIP32Seed string `yaml:"BIP32_SEED"`

	EnclaveCID  int `yaml:"ENCLAVE_CID"`
	EnclavePort int `yaml:"ENCLAVE_PORT"`

	MockEnclave bool   `yaml:"MOCK_ENCLAVE"`
	MockSeed    string `yaml:"MOCK_SEED"`
}
