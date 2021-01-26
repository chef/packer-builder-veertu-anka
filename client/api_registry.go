package client

import (
	"encoding/json"
	"fmt"
)

// https://ankadocs.veertu.com/docs/anka-build-cloud/working-with-registry-and-api/#list-vms
type RegistryTemplateVersionResponse struct {
	Tag          string   `json:"tag"`
	SerialNumber string   `json:"number"`
	Description  string   `json:"description"`
	Images       []string `json:"images"`
	StateFiles   []string `json:"state_files"`
	ConfigFile   string   `json:"config_file"`
	NVRAM        string   `json:"nvram"`
	Size         int      `json:"size"`
}

type RegistryTemplateResponse struct {
	Name     string                            `json:"name"`
	ID       string                            `json:"id"`
	Size     int                               `json:"size"`
	Versions []RegistryTemplateVersionResponse `json:"versions"`
}

func (c *Client) RegistryList() ([]RegistryTemplateResponse, error) {
	response, err := registryRESTRequest("GET", fmt.Sprintf("%s/registry/v2/vm", c.RegistryURL), nil)
	if err != nil {
		return nil, err
	}
	if response.Status != statusOK {
		return nil, fmt.Errorf("failed to revert tag on registry: %s", response.Message)
	}

	var templates []RegistryTemplateResponse
	err = json.Unmarshal(response.Body, &templates)
	if err != nil {
		return nil, err
	}

	return templates, nil
}

// https://ankadocs.veertu.com/docs/anka-build-cloud/working-with-registry-and-api/#revert
func (c *Client) RegistryRevert(id string) error {
	response, err := registryRESTRequest("DELETE", fmt.Sprintf("%s/registry/revert?id=%s", c.RegistryURL, id), nil)
	if err != nil {
		return err
	}
	if response.Status != statusOK {
		return fmt.Errorf("failed to revert VM on registry: %s", response.Message)
	}

	return nil
}
