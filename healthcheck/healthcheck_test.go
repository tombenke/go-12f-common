package healthcheck_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tombenke/go-12f-common/healthcheck"
	"github.com/tombenke/go-12f-common/log"
)

func TestHealthCheckServer(t *testing.T) {
	log.SetupDefault("debug", "text")
	wg := sync.WaitGroup{}
	hc := healthcheck.NewHealthCheck(
		&wg,
		healthcheck.Config{
			8082,
			map[string]healthcheck.Check{
				"/live":  func(ctx context.Context) error { return nil },
				"/ready": func(ctx context.Context) error { return nil },
			},
		},
	)
	hc.Startup(context.Background())
	checkEndpoints(t)
	hc.Shutdown(context.Background())
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
