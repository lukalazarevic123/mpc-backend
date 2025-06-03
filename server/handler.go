package server

import (
	"fmt"
	"io"
	"mpc-backend/config"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	router *mux.Router
	cors   *cors.Cors

	host string
	port int
}

func NewHandler(conf config.Configuration) *Handler {
	handler := &Handler{}
	handler.host = conf.ServerConf.Host
	handler.port = conf.ServerConf.Port

	handler.router = mux.NewRouter()

	handler.router.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	configCors(handler)

	return handler
}

func configCors(handler *Handler) {
	// Defaults cover HEAD, GET, POST
	opts := cors.Options{
		AllowedHeaders:   []string{"*"},
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	}

	// hub.cors = []handlers.CORSOption{headers, methods, origins}
	handler.cors = cors.New(opts)
}

func (h *Handler) Run() error {
	log.Info().Str("host", h.host).Int("port", h.port).Msg("Server started")
	return http.ListenAndServe(fmt.Sprintf("%s:%d", h.host, h.port), h.cors.Handler(h.router))
}

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, _ = io.WriteString(w, `{"alive": true}`)
}
