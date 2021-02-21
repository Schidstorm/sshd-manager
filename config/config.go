package config


type Config struct {
	EtcdEndpoints []string

}

func GetConfig() *Config {
	return &Config{EtcdEndpoints: []string{"localhost:2379"}}
}