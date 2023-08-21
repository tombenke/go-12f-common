package healthcheck_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
	"net/http"
	"os"
	"sync"
	"testing"
)

func TestHealthCheckServer(t *testing.T) {
	log.SetLevelStr("debug")
	wg := sync.WaitGroup{}
	hc, err := healthcheck.NewHealthCheck(&wg, healthcheck.Config{8082, map[string]healthcheck.Check{
		"/live":  func() error { return nil },
		"/ready": func() error { return nil },
	}})
	assert.Nil(t, err)
	hc.Startup()
	checkEndpoints(t)
	hc.Shutdown()
	wg.Wait()
}

func checkEndpoints(t *testing.T) {
	checkEndpoint(t, "http://localhost:8082/live")
	checkEndpoint(t, "http://localhost:8082/ready")
}

func checkEndpoint(t *testing.T, requestURL string) {
	res, err := http.Get(requestURL)

	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	assert.Equal(t, 200, res.StatusCode)
}
