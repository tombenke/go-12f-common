package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/tombenke/go-12f-common/v2/log"
	"github.com/tombenke/go-12f-common/v2/must"
	"github.com/tombenke/go-12f-common/v2/oti"
	"golang.org/x/exp/maps"
)

type ServiceNotAvailableError struct{}

func (e ServiceNotAvailableError) Error() string {
	return "Service is not available"
}

type HealthCheck struct {
	config Config
	server *http.Server
	wg     *sync.WaitGroup
}

// Check is a health/readiness checker function
type Check func(ctx context.Context) error

type Config struct {
	Port   uint
	Checks map[string]Check
}

// Create a HealthCheck instance
func NewHealthCheck(wg *sync.WaitGroup, config Config) HealthCheck {
	return HealthCheck{wg: wg, config: config}
}

// Setup the Healtcheck services and start listening on the HealtCheck port
func (h *HealthCheck) Startup(ctx context.Context) {
	_, logger := h.getLogger(ctx)
	logger.Info("Starting up")
	started := time.Now()
	mux := http.NewServeMux()

	for path, check := range h.config.Checks {
		logger.Debug("Adding endpoint", "path", path)
		endpointPath := path
		checkFun := check
		mux.HandleFunc(endpointPath, func(w http.ResponseWriter, r *http.Request) {
			checkResults := make(map[string]string)
			err := checkFun(ctx)
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
			logger.Info("Server closed")
		} else if err != nil {
			logger.Error("Error listening for server", "error", err)
		}
		cancelCtx()
	}()

	h.waitUntilServerStarted(ctx)
}

func (h *HealthCheck) waitUntilServerStarted(ctx context.Context) {
	_, logger := h.getLogger(ctx)
	if len(h.config.Checks) > 0 {
		for {
			time.Sleep(10 * time.Millisecond)

			logger.Info("Checking if server started...")
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", h.config.Port, maps.Keys(h.config.Checks)[0]))
			if err != nil {
				logger.Error("Server check failed", "error", err)
				continue
			}
			logger.Debug("Checking response", "statusCode", resp.StatusCode)
			if err := resp.Body.Close(); err != nil {
				logger.Error("Failed to close response body", "err", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				logger.Error("Server response is not OK", "statusCode", resp.StatusCode)
				continue
			}

			// If reached this point then server is up and running!
			logger.Info("Server is up and running")
			break
		}
	}
	logger.Info("HealthCheck is up and running!")
}

// Shut down the HealtCheck services
func (h *HealthCheck) Shutdown(ctx context.Context) {
	defer h.wg.Done()
	slog.InfoContext(ctx, "Shutdown", string(oti.FieldComponent), "HealthCheck")
	must.Must(h.server.Shutdown(context.Background()))
}

func (h *HealthCheck) getLogger(ctx context.Context) (context.Context, *slog.Logger) {
	return log.With(ctx, string(oti.FieldComponent), "HealthCheck")
}
