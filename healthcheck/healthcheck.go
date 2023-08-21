package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tombenke/go-12f-common/log"
	"github.com/tombenke/go-12f-common/must"
	"golang.org/x/exp/maps"
	"net/http"
	"sync"
	"time"
)

type ServiceNotAvailableError struct {
}

func (e ServiceNotAvailableError) Error() string {
	return "Service is not available"
}

type HealthCheck struct {
	config Config
	server *http.Server
	wg     *sync.WaitGroup
}

// Check is a health/readiness checker function
type Check func() error

type Config struct {
	Port   uint
	Checks map[string]Check
}

// Create a HealthCheck instance
func NewHealthCheck(wg *sync.WaitGroup, config Config) (*HealthCheck, error) {
	return &HealthCheck{wg: wg, config: config}, nil
}

// Setup the Healtcheck services and start listening on the HealtCheck port
func (h *HealthCheck) Startup() {
	log.Logger.Infof("HealthCheck services Startup")
	started := time.Now()
	mux := http.NewServeMux()

	for path, check := range h.config.Checks {
		log.Logger.Debugf("HealthCheck add endpoint: %s, %v", path, check)
		endpointPath := path
		checkFun := check
		mux.HandleFunc(endpointPath, func(w http.ResponseWriter, r *http.Request) {
			checkResults := make(map[string]string)
			err := checkFun()
			status := http.StatusOK
			if err != nil {
				// write out the response code and content type header
				status = http.StatusServiceUnavailable
				checkResults["error"] = err.Error()
			} else {
				duration := time.Since(started)
				checkResults["uptime"] = fmt.Sprintf("%v", duration.Seconds())
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(status)

			encoder := json.NewEncoder(w)
			encoder.SetIndent("", "    ")
			must.Must(encoder.Encode(checkResults))
		})
	}

	_, cancelCtx := context.WithCancel(context.Background())
	h.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", h.config.Port),
		Handler: mux,
	}

	// Start the blocking server call in a separate thread
	h.wg.Add(1)
	go func() {
		err := h.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Logger.Info("HealthCheck server is closed")
		} else if err != nil {
			log.Logger.Errorf("error listening for HealthCheck server: %s", err)
		}
		cancelCtx()
	}()

	h.waitUntilServerStarted()
}

func (h *HealthCheck) waitUntilServerStarted() {
	if len(h.config.Checks) > 0 {
		for {
			time.Sleep(10 * time.Millisecond)

			log.Logger.Info("Checking if HealthCheck server started...")
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", h.config.Port, maps.Keys(h.config.Checks)[0]))
			if err != nil {
				log.Logger.Errorf("HealthCheck server check failed: %s", err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Logger.Errorf("HealthCheck server response is Not OK: %d", resp.StatusCode)
				continue
			}

			// If reached this point then server is up and running!
			break
		}
	}
	log.Logger.Info("HealthCheck is up and running!")
}

// Shut down the HealtCheck services
func (h *HealthCheck) Shutdown() {
	defer h.wg.Done()
	log.Logger.Infof("HealthCheck services Shutdown")
	must.Must(h.server.Shutdown(context.Background()))
}
