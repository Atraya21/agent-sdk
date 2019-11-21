package apic

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIServer -
type APIServer struct {
	Name       string                 `json:"name"`
	Title      string                 `json:"title"`
	Tags       map[string]interface{} `json:"tags"`
	Attributes map[string]interface{} `json:"attributes"`
	Spec       map[string]interface{} `json:"spec"`
}

// Spec -
type Spec struct {
	Description string `json:"description"`
}

// CreateAPIServerBodyForAdd -
func (c *Client) CreateAPIServerBodyForAdd(apiID, apiName, stageName string, tags map[string]interface{}) ([]byte, error) {

	// attributes used for extraneous data
	attribute := make(map[string]interface{})
	attribute["apiID"] = apiID
	attribute["apiName"] = apiName
	attribute["stageName"] = stageName

	// spec needs to adhere to environment schema
	spec := make(map[string]interface{})
	spec["description"] = "API From AWS APIGateway (RestApiId: " + apiID + ", StageName: " + stageName + ")"

	apiServerService := APIServer{
		Name:       strings.ToLower(apiName), // name needs to be path friendly and follows this regex "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\
		Title:      fmt.Sprintf("%v (Stage: %v)", apiName, stageName),
		Attributes: attribute,
		Spec:       spec,
		Tags:       tags,
	}

	return json.Marshal(apiServerService)
}

// AddAPIServer -
func (c *Client) AddAPIServer(apiServerBuffer []byte) (string, error) {
	// Unit testing. For now just dummy up a return
	if isUnitTesting() {
		return "12345678", nil
	}

	url := c.cfg.GetAPIServerEnvironmentsURL()
	agentMode := c.cfg.GetAgentMode()
	return c.DeployAPI("POST", apiServerBuffer, agentMode, url)
}
