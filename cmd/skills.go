package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const repo = "redhajuanda/krengki"

var upgradeDir string

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage skills and agent configs",
}

var skillsUpgradeCmd = &cobra.Command{
	Use:   "upgrade [version]",
	Short: "Replace dep-managed files with a specific or latest release",
	Long: `Replace all dep-managed files in the project with files from a specific release.
If no version is given, fetches the latest release from GitHub.
Falls back to embedded dep if offline or no releases found.

Examples:
  krengki skills upgrade
  krengki skills upgrade v1.0.0
  krengki skills upgrade v1.0.0 --dir ./my-app`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSkillsUpgrade,
}

func init() {
	skillsUpgradeCmd.Flags().StringVar(&upgradeDir, "dir", ".", "Project directory to upgrade")
	skillsCmd.AddCommand(skillsUpgradeCmd)
}

func runSkillsUpgrade(cmd *cobra.Command, args []string) error {
	projectDir := upgradeDir
	if projectDir == "." {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	info, err := os.Stat(projectDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("project dir not found: %s", projectDir)
	}

	version := ""
	if len(args) == 1 {
		version = args[0]
	}

	fmt.Printf("Upgrading dep files in: %s\n", projectDir)

	if version != "" || isOnline() {
		if version == "" {
			fmt.Println("Fetching latest release...")
			version, err = fetchLatestVersion()
			if err != nil {
				fmt.Printf("Warning: could not fetch latest version (%v), falling back to embedded\n", err)
			}
		}
	}

	if version != "" {
		fmt.Printf("Downloading dep from %s...\n", version)
		if err := downloadAndApplyDep(projectDir, version); err != nil {
			fmt.Printf("Warning: download failed (%v), falling back to embedded\n", err)
			version = ""
		}
	}

	if version == "" {
		fmt.Println("Using embedded dep...")
		if err := copyEmbeddedDep(projectDir); err != nil {
			return fmt.Errorf("copy dep: %w", err)
		}
	}

	fmt.Println("Recreating agent symlinks...")
	if err := createAgentSymlinks(projectDir); err != nil {
		return fmt.Errorf("create symlinks: %w", err)
	}

	fmt.Println("Done.")
	return nil
}

func fetchLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("no releases found")
	}
	return release.TagName, nil
}

// downloadAndApplyDep downloads the source tarball for the given version tag,
// extracts cmd/dep/ from it, and copies the files into projectDir.
func downloadAndApplyDep(projectDir, version string) error {
	url := fmt.Sprintf("https://github.com/%s/archive/refs/tags/%s.tar.gz", repo, version)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download returned %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	// tarball root dir is "krengki-{version-without-v}/"
	ver := strings.TrimPrefix(version, "v")
	depPrefix := fmt.Sprintf("krengki-%s/cmd/dep/", ver)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if !strings.HasPrefix(hdr.Name, depPrefix) {
			continue
		}

		rel := strings.TrimPrefix(hdr.Name, depPrefix)
		if rel == "" {
			continue
		}

		// skip symlink paths — handled by createAgentSymlinks
		relNorm := filepath.FromSlash(rel)
		if symlinkPaths[strings.TrimSuffix(relNorm, string(filepath.Separator))] {
			continue
		}

		dest := filepath.Join(projectDir, relNorm)

		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(dest, 0755); err != nil {
				return err
			}
			continue
		}

		if filepath.Base(dest) == ".gitkeep" {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		f, err := os.Create(dest)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(f, tr)
		f.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}

// isOnline does a quick DNS check to see if network is available.
func isOnline() bool {
	resp, err := http.Get("https://api.github.com")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}
