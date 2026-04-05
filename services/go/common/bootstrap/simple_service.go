package bootstrap

import (
	"log"
	"net/http"
	"os"

	"github.com/KomokPol/kids_of_gpt/services/go/common/httpx"
)

func envOrDefault(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

// RunSimpleService starts a simple service with /healthz endpoint.
func RunSimpleService(serviceName, addrEnv, defaultAddr string) {
	addr := envOrDefault(addrEnv, defaultAddr)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": serviceName})
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	log.Printf("%s listening on %s", serviceName, addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("%s failed: %v", serviceName, err)
	}
}
