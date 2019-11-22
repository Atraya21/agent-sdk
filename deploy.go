package apic

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	corecfg "git.ecd.axway.int/apigov/aws_apigw_discovery_agent/core/config"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var httpClient = http.DefaultClient
var log logrus.FieldLogger = logrus.WithField("package", "apic")

// SetLog sets the logger for the package.
func SetLog(newLog logrus.FieldLogger) {
	log = newLog
	return
}

// DeployAPI -
func (c *Client) DeployAPI(method string, apiServerBuffer []byte, agentMode corecfg.AgentMode, url string) (string, error) {
	request, err := c.SetHeader(method, url, bytes.NewBuffer(apiServerBuffer))
	if err != nil {
		return "", err
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return "", err
	}
	detail := make(map[string]*json.RawMessage)
	if !(response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated) {

		json.NewDecoder(response.Body).Decode(&detail)
		for k, v := range detail {
			buffer, _ := v.MarshalJSON()
			log.Debugf("HTTP response key %v: %v", k, string(buffer))
		}
		return "", errors.New(response.Status)
	}
	defer response.Body.Close()
	json.NewDecoder(response.Body).Decode(&detail)
	return handleResponse(method, agentMode, detail)
}

func handleResponse(method string, agentMode corecfg.AgentMode, detail map[string]*json.RawMessage) (string, error) {
	if strings.ToLower(method) == strings.ToLower("POST") {
		itemID := ""

		// Disconnected Mode
		if agentMode != corecfg.Connected {
			for k, v := range detail {
				buffer, _ := v.MarshalJSON()
				if k == "id" {
					itemID = string(buffer)
				}
				log.Debugf("HTTP response key %v: %v", k, string(buffer))
			}
			if itemID != "" {
				return strconv.Unquote(itemID)
			}
			// Connected Mode
		} else {
			for k, v := range detail {
				buffer, _ := v.MarshalJSON()
				if k == "metadata" {
					itemID = gjson.Get(string(buffer), "id").String()
				}
				log.Debugf("HTTP response key %v: %v", k, string(buffer))
			}
			if itemID != "" {
				return itemID, nil
			}
		}
	}
	// This is an update to catalog item (PUT)
	for k, v := range detail {
		buffer, _ := v.MarshalJSON()
		log.Debugf("HTTP response key %v: %v", k, string(buffer))
	}

	return "", nil

}

// SetHeader - set header
func (c *Client) SetHeader(method, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	var token string

	if token, err = c.tokenRequester.GetToken(); err != nil {
		return nil, err
	}

	request.Header.Add("X-Axway-Tenant-Id", c.cfg.GetTenantID())
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Content-Type", "application/json")
	return request, nil
}
