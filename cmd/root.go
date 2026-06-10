package cmd

import (
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags; falls back to module version info.
var Version = "dev"

func resolvedVersion() string {
	if Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return Version
}

var rootCmd = &cobra.Command{
	Use:     "krengki",
	Short:   "Krengki - project scaffolding CLI",
	Version: resolvedVersion(),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(skillsCmd)
}
