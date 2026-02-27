package init

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/fog/gerrit-cli/internal/cmdutil"
	"github.com/fog/gerrit-cli/internal/config"
	"github.com/fog/gerrit-cli/pkg/gerrit"
)

func NewCmdInit() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize gerrit-cli configuration",
		Long:  "Interactive setup wizard to configure Gerrit server connection.",
		Run:   runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Profile name (default: \"default\"): ")
	profileName, _ := reader.ReadString('\n')
	profileName = strings.TrimSpace(profileName)
	if profileName == "" {
		profileName = "default"
	}

	fmt.Print("Gerrit server URL: ")
	server, _ := reader.ReadString('\n')
	server = strings.TrimSpace(server)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("HTTP password (leave empty to use GERRIT_HTTP_PASSWORD env): ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if password == "" {
		password = os.Getenv("GERRIT_HTTP_PASSWORD")
	}

	fmt.Print("Testing connection... ")
	noAuthPrefix := false
	client := gerrit.NewClient(gerrit.Config{
		Server:   server,
		Username: username,
		Password: password,
	})
	// Test with an authenticated endpoint to properly detect /a/ prefix support.
	_, err := client.Get(context.Background(), "accounts/self")
	if err != nil {
		// Retry without /a/ prefix (nginx-authed instances)
		client = gerrit.NewClient(gerrit.Config{
			Server:       server,
			Username:     username,
			Password:     password,
			NoAuthPrefix: true,
		})
		_, err = client.Get(context.Background(), "accounts/self")
		if err != nil {
			fmt.Printf("FAILED\n")
			cmdutil.ExitIfError(fmt.Errorf("connection test failed: %w", err))
		}
		noAuthPrefix = true
	}
	// Fetch version for display.
	verClient := client
	data, err := verClient.Get(context.Background(), "config/server/version")
	if err != nil {
		data = []byte("unknown")
	}
	version := strings.Trim(string(data), "\"\n")
	if noAuthPrefix {
		fmt.Printf("OK (Gerrit %s, no /a/ auth prefix)\n", version)
	} else {
		fmt.Printf("OK (Gerrit %s)\n", version)
	}

	fmt.Print("Default project (optional, press Enter to skip): ")
	project, _ := reader.ReadString('\n')
	project = strings.TrimSpace(project)

	p := &config.Profile{
		Server:       server,
		Username:     username,
		Password:     password,
		Project:      project,
		NoAuthPrefix: noAuthPrefix,
	}

	// Load existing multi-config or start fresh.
	mc, err := config.LoadMulti()
	if err != nil {
		mc = &config.MultiConfig{
			Profiles: make(map[string]*config.Profile),
		}
	}
	mc.Profiles[profileName] = p

	// Set default if this is the first profile or if it's named "default".
	if mc.Default == "" || profileName == "default" {
		mc.Default = profileName
	}

	cmdutil.ExitIfError(config.SaveMulti(mc))
	fmt.Printf("Profile %q saved to %s\n", profileName, config.ConfigFilePath())
}
