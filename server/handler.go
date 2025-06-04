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

	//god forgive me for this atrocity
	handler.router.HandleFunc("/organizations", handler.CreateOrganizationHandler).Methods("POST")
	handler.router.HandleFunc("/organizations/{address}", handler.GetOrganizationsByAddressHandler).Methods("GET")
	handler.router.HandleFunc("/organization", handler.GetOrganizationByNameHandler).Methods("GET")

	handler.router.HandleFunc("/ws/{address}", handler.WebSocketHandler).Methods("GET")
	handler.router.HandleFunc("/ws/organization/{name}/{address}", handler.OrganizationWebSocketHandler).Methods("GET")

	handler.router.HandleFunc("/transaction/initiate", handler.InitiateTransactionHandler).Methods("POST")
	handler.router.HandleFunc("/transaction/confirm", handler.ConfirmTransactionHandler).Methods("POST")

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

	orgID, err := h.crudHandler.CreateOrganization(orgReq.Name, orgReq.Threshold, orgReq.Participants)
	if err != nil {
		log.Error().Err(err).Msg("CRUD Error")
		http.Error(w, "Server Error", http.StatusBadGateway)
		return
	}

	for _, participant := range orgReq.Participants {
		inviteMsg := types.InvitationMessage{
			OrganizationID:   orgID,
			OrganizationName: orgReq.Name,
			Message:          fmt.Sprintf("You've been invited to join organization %s", orgReq.Name),
		}
		h.hub.NotifyUser(participant.Address, inviteMsg)
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

func (h *Handler) GetOrganizationByNameHandler(w http.ResponseWriter, r *http.Request) {
	// Option 1: Get the name from a query parameter, e.g., /organization?name=Acme
	orgName := r.URL.Query().Get("name")
	if orgName == "" {
		http.Error(w, "Missing organization name", http.StatusBadRequest)
		return
	}

	org, err := h.crudHandler.GetOrganizationByName(orgName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch organization by name")
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) OrganizationWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgName := vars["name"]
	address := vars["address"]
	if orgName == "" || address == "" {
		http.Error(w, "Missing organization name or address", http.StatusBadRequest)
		return
	}

	// Look up the organization using the CRUD handler.
	org, err := h.crudHandler.GetOrganizationByName(orgName)
	if err != nil {
		log.Error().Err(err).Msg("Organization not found")
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Upgrade the HTTP connection to a WebSocket.
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	// Register the new connection and join the organization room using the org’s ID.
	h.hub.RegisterConnection(address, ws)
	h.hub.JoinOrganizationRoom(fmt.Sprintf("%d", org.ID), address)

	// Listen for messages or closure.
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			h.hub.UnregisterConnection(address)
			ws.Close()
			break
		}
	}
}

func (h *Handler) InitiateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var txReq types.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Look up organization by name.
	org, err := h.crudHandler.GetOrganizationByName(txReq.OrganizationName)
	if err != nil {
		log.Error().Err(err).Msg("Organization not found")
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Construct the transaction notification.
	notification := types.TransactionNotification{
		OrganizationID: org.ID,
		Initiator:      txReq.Initiator,
		Message:        fmt.Sprintf("Transaction initiated by: %s", txReq.Initiator),
	}

	orgIDKey := fmt.Sprintf("%d", org.ID)

	// Broadcast the transaction notification to everyone in the organization's room.
	h.hub.BroadcastOrganization(orgIDKey, notification)

	// Register the transaction in the hub to track confirmations.
	h.hub.RegisterPendingTransaction(orgIDKey, notification, org.Threshold)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "transaction initiated"})
}

func (h *Handler) ConfirmTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var confirmReq types.TransactionConfirmationRequest
	if err := json.NewDecoder(r.Body).Decode(&confirmReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Look up organization by name.
	org, err := h.crudHandler.GetOrganizationByName(confirmReq.OrganizationName)
	if err != nil {
		log.Error().Err(err).Msg("Organization not found")
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	orgIDKey := fmt.Sprintf("%d", org.ID)

	// Increment confirmation count.
	h.hub.mu.Lock()
	txState, exists := h.hub.pendingTransactions[orgIDKey]
	if !exists {
		h.hub.mu.Unlock()
		http.Error(w, "No pending transaction for this organization", http.StatusNotFound)
		return
	}
	txState.Confirmations++
	currentConf := txState.Confirmations
	h.hub.mu.Unlock()

	// Optionally, notify all users about the update.
	updateMsg := map[string]string{
		"update": fmt.Sprintf("Transaction confirmations: %d/%d", currentConf, txState.Threshold),
	}
	h.hub.BroadcastOrganization(orgIDKey, updateMsg)

	// If threshold is reached, send a final notification.
	if currentConf >= txState.Threshold {
		finalNotif := types.TransactionNotification{
			OrganizationID: org.ID,
			Initiator:      txState.TransactionNotification.Initiator,
			Details:        txState.TransactionNotification.Details,
			Message:        "Transaction confirmed by threshold",
		}
		h.hub.BroadcastOrganization(orgIDKey, finalNotif)

		// Remove the pending transaction after confirmation.
		h.hub.mu.Lock()
		delete(h.hub.pendingTransactions, orgIDKey)
		h.hub.mu.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"confirmations": currentConf})
}
