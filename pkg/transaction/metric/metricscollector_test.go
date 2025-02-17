package metric

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Axway/agent-sdk/pkg/agent"
	"github.com/Axway/agent-sdk/pkg/cmd"
	"github.com/Axway/agent-sdk/pkg/config"
	"github.com/elastic/beats/v7/libbeat/paths"
	"github.com/stretchr/testify/assert"
)

func createCentralCfg(url, env string) *config.CentralConfiguration {

	cfg := config.NewCentralConfig(config.DiscoveryAgent).(*config.CentralConfiguration)
	cfg.URL = url
	cfg.TenantID = "123456"
	cfg.Environment = env
	authCfg := cfg.Auth.(*config.AuthConfiguration)
	authCfg.URL = url + "/auth"
	authCfg.Realm = "Broker"
	authCfg.ClientID = "DOSA_1111"
	authCfg.PrivateKey = "../../transaction/testdata/private_key.pem"
	authCfg.PublicKey = "../../transaction/testdata/public_key"
	cfg.PublishUsageEvents = true
	// cfg.PublishMetricEvents = true
	return cfg
}

var accessToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJ0ZXN0IiwiaWF0IjoxNjE0NjA0NzE0LCJleHAiOjE2NDYxNDA3MTQsImF1ZCI6InRlc3RhdWQiLCJzdWIiOiIxMjM0NTYiLCJvcmdfZ3VpZCI6IjEyMzQtMTIzNC0xMjM0LTEyMzQifQ.5Uqt0oFhMgmI-sLQKPGkHwknqzlTxv-qs9I_LmZ18LQ"

type testHTTPServer struct {
	lighthouseEventCount int
	transactionCount     int
	failUsageEvent       bool
	server               *httptest.Server
}

func (s *testHTTPServer) startServer() {
	s.server = httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.RequestURI, "/auth") {
			token := "{\"access_token\":\"" + accessToken + "\",\"expires_in\": 12235677}"
			resp.Write([]byte(token))
		}
		if strings.Contains(req.RequestURI, "/lighthouse") {
			if s.failUsageEvent {
				resp.WriteHeader(http.StatusBadRequest)
				return
			}
			s.lighthouseEventCount++
			req.ParseMultipartForm(1 << 20)
			for _, fileHeaders := range req.MultipartForm.File {
				for _, fileHeader := range fileHeaders {
					file, err := fileHeader.Open()
					if err != nil {
						return
					}
					body, _ := ioutil.ReadAll(file)
					var usageEvent LighthouseUsageEvent
					json.Unmarshal(body, &usageEvent)
					for _, report := range usageEvent.Report {
						for _, usage := range report.Usage {
							s.transactionCount += int(usage)
						}
					}
				}
			}
		}
	}))
}

func (s *testHTTPServer) closeServer() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *testHTTPServer) resetConfig() {
	s.lighthouseEventCount = 0
	s.transactionCount = 0
	s.failUsageEvent = false
}

func cleanUpCachedMetricFile() {
	os.Remove("./" + cacheFileName)
}

func TestMetricCollector(t *testing.T) {
	defer cleanUpCachedMetricFile()
	s := &testHTTPServer{}
	defer s.closeServer()
	s.startServer()
	paths.Paths.Data = "."

	cfg := createCentralCfg(s.server.URL, "demo")
	cfg.LighthouseURL = s.server.URL + "/lighthouse"
	cfg.SetEnvironmentID("267bd671-e5e2-4679-bcc3-bbe7b70f30fd")
	cmd.BuildDataPlaneType = "Azure"
	agent.Initialize(cfg)

	myCollector := createMetricCollector()
	metricCollector := myCollector.(*collector)

	testCases := []struct {
		name                     string
		loopCount                int
		apiTransactionCount      []int
		failUsageEventOnServer   []bool
		expectedLHEvents         []int
		expectedTransactionCount []int
	}{
		// Success case
		{
			name:                     "WithLighthouse",
			loopCount:                1,
			apiTransactionCount:      []int{5},
			failUsageEventOnServer:   []bool{false},
			expectedLHEvents:         []int{1},
			expectedTransactionCount: []int{5},
		},
		// Success case with no usage report
		{
			name:                     "WithLighthouseNoUsageReport",
			loopCount:                1,
			apiTransactionCount:      []int{0},
			failUsageEventOnServer:   []bool{false},
			expectedLHEvents:         []int{0},
			expectedTransactionCount: []int{0},
		},
		// Test case with failing request to LH, the subsequent successful request should contain the total count since initial failure
		{
			name:                     "WithLighthouseWithFailure",
			loopCount:                3,
			apiTransactionCount:      []int{5, 10, 2},
			failUsageEventOnServer:   []bool{false, true, false},
			expectedLHEvents:         []int{1, 1, 2},
			expectedTransactionCount: []int{5, 5, 17},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			for l := 0; l < test.loopCount; l++ {
				for i := 0; i < test.apiTransactionCount[l]; i++ {
					metricCollector.AddMetric("111", "111", "200", 10, "", "")
				}
				s.failUsageEvent = test.failUsageEventOnServer[l]
				metricCollector.Execute()
				assert.Equal(t, test.expectedLHEvents[l], s.lighthouseEventCount)
				assert.Equal(t, test.expectedTransactionCount[l], s.transactionCount)
			}
			s.resetConfig()
		})
	}
}

func TestMetricCollectorCache(t *testing.T) {
	defer cleanUpCachedMetricFile()
	s := &testHTTPServer{}
	defer s.closeServer()
	s.startServer()

	cfg := createCentralCfg(s.server.URL, "demo")
	cfg.LighthouseURL = s.server.URL + "/lighthouse"
	cfg.SetEnvironmentID("267bd671-e5e2-4679-bcc3-bbe7b70f30fd")
	cmd.BuildDataPlaneType = "Azure"
	agent.Initialize(cfg)

	paths.Paths.Data = "."
	myCollector := createMetricCollector()
	metricCollector := myCollector.(*collector)

	metricCollector.AddMetric("111", "111", "200", 5, "", "")
	metricCollector.AddMetric("111", "111", "200", 10, "", "")
	metricCollector.Execute()
	metricCollector.AddMetric("111", "111", "401", 15, "", "")
	metricCollector.AddMetric("222", "222", "200", 20, "", "")
	metricCollector.AddMetric("222", "222", "200", 10, "", "")

	// No event generation/publish, store the cache
	metricCollector.storage.save()
	// Validate only one usage report sent with first 2 transactions
	assert.Equal(t, 1, s.lighthouseEventCount)
	assert.Equal(t, 2, s.transactionCount)
	s.resetConfig()

	// Recreate the collector that loads the stored metrics, so 3 transactions
	myCollector = createMetricCollector()
	metricCollector = myCollector.(*collector)

	metricCollector.AddMetric("111", "111", "200", 5, "", "")
	metricCollector.AddMetric("111", "111", "200", 10, "", "")
	metricCollector.AddMetric("111", "111", "401", 15, "", "")
	metricCollector.AddMetric("222", "222", "200", 20, "", "")
	metricCollector.AddMetric("222", "222", "200", 10, "", "")

	metricCollector.Execute()
	// Validate only one usage report sent with 3 previous transactions and 5 new transactions
	assert.Equal(t, 1, s.lighthouseEventCount)
	assert.Equal(t, 8, s.transactionCount)

	s.resetConfig()
	// Recreate the collector that loads the stored metrics, 0 transactions
	myCollector = createMetricCollector()
	metricCollector = myCollector.(*collector)

	metricCollector.Execute()
	// Validate only no usage report sent as no previous or new transactions
	assert.Equal(t, 0, s.lighthouseEventCount)
	assert.Equal(t, 0, s.transactionCount)
}
