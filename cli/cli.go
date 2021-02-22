package cli

import "github.com/spf13/cobra"

type CliConfig struct {
	ConfigFile string
}

func Run(handler func(config *CliConfig) error) {
	rootCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &CliConfig{
				ConfigFile: "",
			}
			flags := cmd.PersistentFlags()
			if v, err := flags.GetString("config"); err == nil {
				config.ConfigFile = v
			}

			return handler(config)
		},
	}

	rootCmd.PersistentFlags().String("config", "/etc/sshd/manager.yml", "Path of the configuration file.")

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
