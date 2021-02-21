package manager


type Account struct {
	Key string `yaml:"key"`
	Label string `yaml:"label"`
	PublicKey string `yaml:"publicKey"`
	Groups []string `yaml:"groups"`
}