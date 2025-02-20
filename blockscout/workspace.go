package blockscout

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"

	"github.com/ethereum/go-ethereum/log"
)

const scoutupGlobalWorkspace = "scoutup"

var (
	//go:embed embed/docker-compose.yml
	dockerComposeYml []byte

	//go:embed embed/common-blockscout.env
	commonBlockscoutEnv []byte

	//go:embed embed/common-frontend.env
	commonFrontendEnv []byte
)

func globalWorkspace() string {
	return path.Join(os.TempDir(), scoutupGlobalWorkspace)
}

func createGlobalWorkspace() (string, error) {
	globalWorkspace := globalWorkspace()
	err := os.MkdirAll(globalWorkspace, 0755)
	if err != nil {
		return "", err
	}
	return globalWorkspace, nil
}

func CleanupGlobalWorkspace(log log.Logger) error {
	globalWorkspace := globalWorkspace()
	instanceWorkspaces, err := os.ReadDir(globalWorkspace)
	if err != nil {
		return err
	}

	for _, workspace := range instanceWorkspaces {
		if workspace.IsDir() {
			log.Info("Cleaning up instance workspace", "workspace", workspace.Name())
			err := cleanupInstanceWorkspace(filepath.Join(globalWorkspace, workspace.Name()))
			if err != nil {
				log.Error("Failed to cleanup instance workspace", "workspace", workspace.Name(), "error", err)
			}
		}
	}

	return nil
}

func createInstanceWorkspace(globalWorkspace string, genesisJSON []byte) (string, error) {
	workspace, err := os.MkdirTemp(globalWorkspace, "instance")
	if err != nil {
		return "", err
	}

	files := map[string][]byte{
		"docker-compose.yml":    dockerComposeYml,
		"common-blockscout.env": commonBlockscoutEnv,
		"common-frontend.env":   commonFrontendEnv,
		"genesis.json":          genesisJSON,
	}

	for name, content := range files {
		err := os.WriteFile(path.Join(workspace, name), content, 0644)
		if err != nil {
			return "", err
		}
	}
	return workspace, nil
}

func cleanupInstanceWorkspace(dir string) error {
	err := checkWorkspace(dir)
	if err != nil {
		return fmt.Errorf("not a scoutup workspace: %w", err)
	}

	cmd := exec.Command("docker", "compose", "down")
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to remove docker containers: %w", err)
	}

	err = os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to clean directory: %w", err)
	}

	return nil
}

func checkWorkspace(dir string) error {
	expected := []string{"docker-compose.yml", "common-blockscout.env", "common-frontend.env", "logs", "genesis.json"}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	if len(entries) != len(expected) {
		return fmt.Errorf("expected %d files but got %d", len(expected), len(entries))
	}

	actual := []string{}
	for _, entry := range entries {
		actual = append(actual, entry.Name())
	}

	sort.Strings(expected)
	sort.Strings(actual)

	for i := range expected {
		if expected[i] != actual[i] {
			return fmt.Errorf("expected %s but got %s", expected[i], actual[i])
		}
	}

	return nil
}
