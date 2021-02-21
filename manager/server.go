package manager

type Server struct {
	Key string `yaml:"key"`
	Hostname string `yaml:"hostname"`
	Groups []string `yaml:"groups"`
}
