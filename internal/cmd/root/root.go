package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/internal/cmd/change"
	initCmd "github.com/fog/gerrit-cli/internal/cmd/init"
	"github.com/fog/gerrit-cli/internal/config"
)

var (
	cfgFile string
	profile string
	debug   bool
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetEnvPrefix("GERRIT")
	viper.AutomaticEnv()

	if cfgFile != "" {
		// Explicit --config: use as flat file, skip profile logic.
		viper.SetConfigFile(cfgFile)
		_ = viper.ReadInConfig()
		return
	}

	// Load multi-profile config.
	mc, err := config.LoadMulti()
	if err != nil {
		// Config doesn't exist yet (pre-init) — that's fine.
		return
	}

	// Determine which profile to use: --profile flag > GERRIT_PROFILE env > config default.
	name := profile
	if name == "" {
		name = os.Getenv("GERRIT_PROFILE")
	}

	p, err := mc.ResolveProfile(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Flatten profile values into viper so all existing code keeps working.
	viper.Set("server", p.Server)
	viper.Set("username", p.Username)
	viper.Set("password", p.Password)
	viper.Set("project", p.Project)
	viper.Set("no_auth_prefix", p.NoAuthPrefix)
}

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gerrit <command> <subcommand>",
		Short: "Gerrit code review CLI",
		Long:  "A command line tool for Gerrit code review workflows.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			name := cmd.Name()
			if name == "init" || name == "help" || name == "gerrit" || name == "version" {
				return
			}
			if !config.Exists(viper.ConfigFileUsed()) && !config.Exists(config.ConfigFilePath()) {
				fmt.Fprintln(os.Stderr, "Missing configuration. Run 'gerrit init' to set up.")
				os.Exit(1)
			}
		},
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.gerrit/.config.yml)")
	cmd.PersistentFlags().StringVar(&profile, "profile", "", "config profile name (overrides GERRIT_PROFILE)")
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug output")
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	cmd.AddCommand(
		initCmd.NewCmdInit(),
		change.NewCmdChange(),
		newVersionCmd(),
	)

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("gerrit-cli v0.1")
		},
	}
}

func Execute() {
	if err := NewCmdRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
