package client

import (
	"fmt"
)

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
