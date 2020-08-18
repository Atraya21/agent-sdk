package apic

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	apicClient "git.ecd.axway.org/apigov/apic_agents_sdk/pkg/api"
	corecfg "git.ecd.axway.org/apigov/apic_agents_sdk/pkg/config"
)

type mockResponse struct {
	fileName  string
	respCode  int
	errString string
}

type mockHTTPClient struct {
	Client
	respCount int
	responses []mockResponse
}

// Send - send the http request and returns the API Response
func (c *mockHTTPClient) Send(request apicClient.Request) (*apicClient.Response, error) {
	responseFile, _ := os.Open(c.responses[c.respCount].fileName) // APIC Environments
	dat, _ := ioutil.ReadAll(responseFile)

	response := apicClient.Response{
		Code:    c.responses[c.respCount].respCode,
		Body:    dat,
		Headers: map[string][]string{},
	}

	var err error
	if c.responses[c.respCount].errString != "" {
		err = fmt.Errorf(c.responses[c.respCount].errString)
	}
	c.respCount++
	return &response, err
}

func TestCheckAPIServerHealth(t *testing.T) {
	c, cfg := createServiceClient(nil)
	cfg.Environment = "Environment"
	mockClient := mockHTTPClient{
		respCount: 0,
		responses: []mockResponse{
			{
				fileName: "./testdata/apic-environment.json",
				respCode: http.StatusOK,
			},
			{
				fileName: "./testdata/apic-team-notfound.json",
				respCode: http.StatusOK,
			},
		},
	}
	c.apiClient = &mockClient
	c.tokenRequester = MockTokenGetter

	// Test DiscoveryAgent, PublishToEnvironment and with team not found specified
	mockClient.respCount = 0
	mockClient.responses[0].fileName = "./testdata/apiserver-environment.json"
	cfg.Mode = corecfg.PublishToEnvironment
	err := c.checkAPIServerHealth()
	assert.NotNil(t, err, "Expecting error to be returned from the health check with discovery agent in publishToEnvironment mode for invalid team name")

	// Test Team found
	mockClient.respCount = 0
	mockClient.responses = []mockResponse{
		{
			fileName: "./testdata/apiserver-environment.json",
			respCode: http.StatusOK,
		},
		{
			fileName: "./testdata/apic-team.json",
			respCode: http.StatusOK,
		},
	}
	c.cfg.SetEnvironmentID("")
	err = c.checkAPIServerHealth()
	assert.Nil(t, err, "An unexpected error was returned from the health check with discovery agent in publishToEnvironment mode")

	// Test TraceabilityAgent, publishToEnvironment
	cfg.AgentType = corecfg.TraceabilityAgent
	cfg.Mode = corecfg.PublishToEnvironment
	mockClient.respCount = 0
	mockClient.responses = []mockResponse{
		{
			fileName: "./testdata/apiserver-environment.json",
			respCode: http.StatusOK,
		},
	}
	err = c.checkAPIServerHealth()
	assert.Nil(t, err, "An unexpected error was returned from the health check with traceability agent in publishToEnvironment mode")
	assert.Equal(t, "e4e085bf70638a1d0170639297610000", cfg.GetEnvironmentID(), "The EnvironmentID was not set correctly, Traceability and publishToEnvironment mode")
}

func TestNewClientWithTLSConfig(t *testing.T) {
	tlsCfg := corecfg.NewTLSConfig()
	client, cfg := createServiceClient(tlsCfg)
	assert.NotNil(t, client)
	assert.NotNil(t, cfg)
}
