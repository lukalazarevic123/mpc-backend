package server

import (
	"encoding/json"
	"fmt"
	"io"
	"mpc-backend/config"
	crud "mpc-backend/core"
	"mpc-backend/types"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	router *mux.Router
	cors   *cors.Cors

	host string
	port int

	crudHandler *crud.CRUD
	hub         *Hub
}

func NewHandler(conf config.Configuration, crudHandler *crud.CRUD) *Handler {
	handler := &Handler{}
	handler.host = conf.ServerConf.Host
	handler.port = conf.ServerConf.Port

	handler.crudHandler = crudHandler

	handler.router = mux.NewRouter()
	handler.hub = NewHub()

	handler.router.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	handler.router.HandleFunc("/organizations", handler.CreateOrganizationHandler).Methods("POST")
	handler.router.HandleFunc("/organizations/{address}", handler.GetOrganizationsByAddressHandler).Methods("GET")

	handler.router.HandleFunc("/ws/{address}", handler.WebSocketHandler).Methods("GET")

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

func (h *Handler) CreateOrganizationHandler(w http.ResponseWriter, r *http.Request) {

	var orgReq types.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&orgReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.crudHandler.CreateOrganization(orgReq.Name, orgReq.Threshold, orgReq.Participants); err != nil {
		log.Error().Err(err).Msg("CRUD Error")
		http.Error(w, "Server Error", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(orgReq)
}

func (h *Handler) GetOrganizationsByAddressHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	if address == "" {
		http.Error(w, "Missing address parameter", http.StatusBadRequest)
		return
	}

	orgs, err := h.crudHandler.GetOrganizationsByAddress(address)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get organizations by address")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address, ok := vars["address"]
	if !ok || address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	// Register the new connection with the Hub.
	h.hub.RegisterConnection(address, ws)

	// Listen for close or messages (for demonstration, we simply wait until error)
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			h.hub.UnregisterConnection(address)
			ws.Close()
			break
		}
	}
}
