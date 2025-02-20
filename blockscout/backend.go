package blockscout

import (
	"encoding/json"
	"fmt"

	"github.com/blockscout/scoutup/utils"
	"github.com/ethereum/go-ethereum/common"
)

func getSmartContractUrl(backendURL string, address common.Address) string {
	return fmt.Sprintf("%s/api/v2/smart-contracts/%s", backendURL, address)
}

func retrieveProxyImplementationAddresses(backendURL string, proxy common.Address) ([]common.Address, error) {
	body, err := utils.MakeGetRequest(getSmartContractUrl(backendURL, proxy))
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

func isHealthy(backendURL string) bool {
	url := fmt.Sprintf("%s/api/health", backendURL)

	body, err := utils.MakeGetRequest(url)
	if err != nil {
		return false
	}

	type Response struct {
		Healthy bool `json:"healthy"`
	}

	var data Response
	if err := json.Unmarshal(body, &data); err != nil {
		return false
	}
	return data.Healthy
}
