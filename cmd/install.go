package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
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
	installCmd.Flags().BoolVar(&flagAppRouter, "app", true, "Use App Router (default: true)")
	installCmd.Flags().BoolVar(&flagSrcDir, "src-dir", false, "Use src/ directory")
	installCmd.Flags().BoolVar(&flagTurbopack, "turbopack", false, "Enable Turbopack")
	installCmd.Flags().StringVar(&flagImportAlias, "import-alias", "", "Import alias (e.g. @/*)")
	installCmd.Flags().BoolVar(&flagSkipInstall, "skip-install", false, "Skip package install")
	installCmd.Flags().StringVar(&flagPackageMgr, "package-manager", "pnpm", "Package manager to use (npm, yarn, pnpm, bun)")
}

func runInstall(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	createNextAppSpec, err := resolveCreateNextAppSpec()
	if err != nil {
		return err
	}

	npxArgs := []string{createNextAppSpec, projectName}

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

func resolveCreateNextAppSpec() (string, error) {
	out, err := exec.Command("npm", "view", "create-next-app", "dist-tags", "--json").Output()
	if err != nil {
		return "", fmt.Errorf("resolve create-next-app dist-tags: %w", err)
	}

	var tags map[string]string
	if err := json.Unmarshal(out, &tags); err != nil {
		return "", fmt.Errorf("parse create-next-app dist-tags: %w", err)
	}

	tag, err := selectCreateNextAppTag(tags)
	if err != nil {
		return "", err
	}

	return "create-next-app@" + tag, nil
}

func selectCreateNextAppTag(tags map[string]string) (string, error) {
	if _, ok := tags["lts"]; ok {
		return "lts", nil
	}

	latest, ok := parseVersion(tags["latest"])
	if !ok {
		return "", fmt.Errorf("create-next-app dist-tags missing stable latest version")
	}

	var candidates []versionedTag
	for tag, value := range tags {
		if !strings.HasPrefix(tag, "next-") {
			continue
		}

		v, ok := parseVersion(value)
		if !ok || v.major >= latest.major {
			continue
		}

		candidates = append(candidates, versionedTag{tag: tag, version: v})
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("create-next-app dist-tags has no stable tag below latest %s", tags["latest"])
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[j].version.less(candidates[i].version)
	})

	return candidates[0].tag, nil
}

type versionedTag struct {
	tag     string
	version version
}

type version struct {
	major int
	minor int
	patch int
}

func parseVersion(value string) (version, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return version{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return version{}, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return version{}, false
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return version{}, false
	}

	return version{major: major, minor: minor, patch: patch}, true
}

func (v version) less(other version) bool {
	if v.major != other.major {
		return v.major < other.major
	}
	if v.minor != other.minor {
		return v.minor < other.minor
	}
	return v.patch < other.patch
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
