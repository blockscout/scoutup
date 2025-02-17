package utils

import (
	"os"
	"strings"

	"github.com/joho/godotenv"

	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"net/http"
)

func PatchDotEnv(path string, envs map[string]string) error {
	dotEnv, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dotEnv.Close()

	env, err := godotenv.Parse(dotEnv)
	if err != nil {
		return err
	}

	env = mergeMaps(env, envs)
	return godotenv.Write(env, path)
}

func mergeMaps(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func NameToContainerName(prefix string, name string) string {
	container_name := strings.ToLower(name)
	container_name = strings.ReplaceAll(container_name, " ", "-")
	return prefix + "-" + container_name
}

func GetSmartContract(backendURL string, address common.Address) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v2/smart-contracts/%s", backendURL, address)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("invalid response status code: %s", resp.Status))
	}

	return io.ReadAll(resp.Body)
}

func RetrieveProxyImplementationAddresses(backendURL string, proxy common.Address) ([]common.Address, error) {
	body, err := GetSmartContract(backendURL, proxy)
	if err != nil {
		return nil, err
	}

	type Implementation struct {
		Address common.Address `json:"address"`
	}

	type Response struct {
		Implementations []Implementation `json:"implementations"`
	}

	var data Response
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	var addresses []common.Address
	for _, implementation := range data.Implementations {
		addresses = append(addresses, implementation.Address)
	}
	return addresses, nil
}
