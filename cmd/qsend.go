package cmd

import (
	"qsend/config"

	"github.com/spf13/cobra"
)

const appName = "qsend"

var flags config.Config
var path string

func init() {
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(receiveCmd)
	rootCmd.AddCommand(configCmd)

	rootFlags := rootCmd.PersistentFlags()
	rootFlags.StringVarP(&path, "config", "c", "",
		"path to the config file, defaults to $XDG_CONFIG_HOME/qsend/config.json")
	rootFlags.StringVar(&flags.Bind, "bind", "", "address to bind the web server")
	rootFlags.Uint16VarP(&flags.Port, "port", "p", 0, "port to bind the web server")
	rootFlags.StringVar(&flags.TlsCert, "tls-cert", "", "path to TLS certificate")
	rootFlags.StringVar(&flags.TlsKey, "tls-key", "", "path to TLS private key")
	rootFlags.BoolVarP(&flags.Reversed.Value, "reversed", "r", false, "reverse QR code colors")
	receiveCmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", "",
		"output directory for receiving files")
}

func loadConfig() (config.Config, error) {
	cfg, err := config.Load(appName, path)
	if err != nil {
		return config.Config{}, err
	}
	cfgCopy := cfg.Config
	cfgCopy.Override(flags)
	cfgCopy.Normalize()

	if cfgCopy.Bind == "" {
		cfgCopy.Bind, err = selectBind()
		if err != nil {
			return config.Config{}, err
		}
	}
	return cfgCopy, nil
}

const version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:     "qsend",
	Version: version,
	PreRun: func(cmd *cobra.Command, args []string) {
		rootFlags := cmd.PersistentFlags()
		if rev := rootFlags.Lookup("reversed"); rev != nil {
			flags.Reversed.IsSet = rev.Changed
		}
		if len(args) == 0 {
			cmd.RunE = receiveCmd.RunE
		}
	},
	Args: cobra.MinimumNArgs(0),
	RunE: sendCmd.RunE,
}

// Execute is the main entry point for the CLI.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErrf("Error: %v\nRun `qsend help` for help.\n", err)
		return err
	}
	return nil
}
