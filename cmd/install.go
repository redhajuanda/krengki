package cmd

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"os/exec"
)

//go:embed all:dep
var depFS embed.FS

var (
	flagTypeScript  bool
	flagJavaScript  bool
	flagTailwind    bool
	flagESLint      bool
	flagAppRouter   bool
	flagSrcDir      bool
	flagTurbopack   bool
	flagImportAlias string
	flagSkipInstall bool
	flagPackageMgr  string
)

var installCmd = &cobra.Command{
	Use:   "install [project-name]",
	Short: "Create a new Next.js project",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstall,
}

func init() {
	installCmd.Flags().BoolVar(&flagTypeScript, "typescript", true, "Use TypeScript (default: true)")
	installCmd.Flags().BoolVar(&flagJavaScript, "javascript", false, "Use JavaScript")
	installCmd.Flags().BoolVar(&flagTailwind, "tailwind", true, "Add Tailwind CSS (default: true)")
	installCmd.Flags().BoolVar(&flagESLint, "eslint", false, "Add ESLint")
	installCmd.Flags().BoolVar(&flagAppRouter, "app", false, "Use App Router")
	installCmd.Flags().BoolVar(&flagSrcDir, "src-dir", false, "Use src/ directory")
	installCmd.Flags().BoolVar(&flagTurbopack, "turbopack", false, "Enable Turbopack")
	installCmd.Flags().StringVar(&flagImportAlias, "import-alias", "", "Import alias (e.g. @/*)")
	installCmd.Flags().BoolVar(&flagSkipInstall, "skip-install", false, "Skip package install")
	installCmd.Flags().StringVar(&flagPackageMgr, "package-manager", "pnpm", "Package manager to use (npm, yarn, pnpm, bun)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	npxArgs := []string{"create-next-app", projectName}

	if flagJavaScript {
		npxArgs = append(npxArgs, "--js")
	} else {
		npxArgs = append(npxArgs, "--ts")
	}

	if flagTailwind {
		npxArgs = append(npxArgs, "--tailwind")
	} else {
		npxArgs = append(npxArgs, "--no-tailwind")
	}

	if flagESLint {
		npxArgs = append(npxArgs, "--eslint")
	} else {
		npxArgs = append(npxArgs, "--no-eslint")
	}

	if flagAppRouter {
		npxArgs = append(npxArgs, "--app")
	} else {
		npxArgs = append(npxArgs, "--no-app")
	}

	if flagSrcDir {
		npxArgs = append(npxArgs, "--src-dir")
	} else {
		npxArgs = append(npxArgs, "--no-src-dir")
	}

	if flagTurbopack {
		npxArgs = append(npxArgs, "--turbopack")
	} else {
		npxArgs = append(npxArgs, "--no-turbopack")
	}

	if flagImportAlias != "" {
		npxArgs = append(npxArgs, "--import-alias", flagImportAlias)
	}

	if flagSkipInstall {
		npxArgs = append(npxArgs, "--skip-install")
	}

	npxArgs = append(npxArgs, "--use-"+flagPackageMgr)

	fmt.Printf("Creating Next.js project: %s\n", projectName)
	fmt.Printf("Running: npx %v\n\n", npxArgs)

	c := exec.Command("npx", npxArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	if err := c.Run(); err != nil {
		return err
	}

	projectDir, err := filepath.Abs(projectName)
	if err != nil {
		return err
	}

	fmt.Println("\nCopying dep files...")
	if err := copyEmbeddedDep(projectDir); err != nil {
		return fmt.Errorf("copy dep: %w", err)
	}

	fmt.Println("Creating agent symlinks...")
	if err := createAgentSymlinks(projectDir); err != nil {
		return fmt.Errorf("create symlinks: %w", err)
	}

	fmt.Printf("\nDone! cd %s\n", projectName)
	return nil
}

// symlinkPaths are paths (relative to project root) that must be symlinks, not real dirs.
// Skip them during copy; createAgentSymlinks handles them.
var symlinkPaths = map[string]bool{
	filepath.Join(".claude", "skills"):   true,
	filepath.Join(".claude", "commands"): true,
	filepath.Join(".cursor", "skills"):   true,
	filepath.Join(".cursor", "commands"): true,
}

// copyEmbeddedDep copies all files from embedded dep/ into destDir, overwriting duplicates.
func copyEmbeddedDep(destDir string) error {
	return fs.WalkDir(depFS, "dep", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// strip leading "dep/" prefix to get relative path inside project
		rel, err := filepath.Rel("dep", path)
		if err != nil {
			return err
		}

		// skip paths that must be symlinks
		if symlinkPaths[rel] {
			return fs.SkipDir
		}

		dest := filepath.Join(destDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		// skip .gitkeep placeholder files
		if d.Name() == ".gitkeep" {
			return nil
		}

		return copyEmbeddedFile(path, dest)
	})
}

func copyEmbeddedFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	in, err := depFS.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// createAgentSymlinks creates .claude and .cursor dirs with skills/commands symlinked to .agents equivalents.
func createAgentSymlinks(projectDir string) error {
	agentsDirs := []string{
		filepath.Join(projectDir, ".agents", "skills"),
		filepath.Join(projectDir, ".agents", "commands"),
	}
	for _, d := range agentsDirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	type link struct {
		linkPath string
		target   string // relative to linkPath's parent
	}

	links := []link{
		{filepath.Join(projectDir, ".claude", "skills"), filepath.Join("..", ".agents", "skills")},
		{filepath.Join(projectDir, ".claude", "commands"), filepath.Join("..", ".agents", "commands")},
		{filepath.Join(projectDir, ".cursor", "skills"), filepath.Join("..", ".agents", "skills")},
		{filepath.Join(projectDir, ".cursor", "commands"), filepath.Join("..", ".agents", "commands")},
	}

	for _, l := range links {
		if err := os.MkdirAll(filepath.Dir(l.linkPath), 0755); err != nil {
			return err
		}
		// remove if exists (file or old symlink)
		_ = os.Remove(l.linkPath)
		if err := os.Symlink(l.target, l.linkPath); err != nil {
			return err
		}
		fmt.Printf("  symlink: %s -> %s\n", l.linkPath, l.target)
	}

	return nil
}
