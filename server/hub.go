package server

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Connection represents a user websocket connection.
type Connection struct {
	Conn    *websocket.Conn
	Address string
}

// Hub manages active websocket connections and organization rooms.
type Hub struct {
	// mu protects the maps below.
	mu sync.RWMutex
	// mapping of user addresses to their websocket connections
	connections map[string]*Connection
	// mapping of organization IDs to a set of addresses connected to that room
	orgRooms map[string]map[string]*Connection
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*Connection),
		orgRooms:    make(map[string]map[string]*Connection),
	}
}

// RegisterConnection registers a connection by address.
func (h *Hub) RegisterConnection(address string, ws *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conn := &Connection{Conn: ws, Address: address}
	h.connections[address] = conn
	log.Printf("Registered connection for address: %s", address)
}

// UnregisterConnection removes a connection from the hub and all rooms.
func (h *Hub) UnregisterConnection(address string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, address)
	// Remove from all organization rooms
	for orgID, members := range h.orgRooms {
		delete(members, address)
		// remove empty rooms
		if len(members) == 0 {
			delete(h.orgRooms, orgID)
		}
	}
	log.Printf("Unregistered connection for address: %s", address)
}

// JoinOrganizationRoom adds a connection to an organization room.
func (h *Hub) JoinOrganizationRoom(orgID, address string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conn, ok := h.connections[address]
	if !ok {
		log.Printf("No connection for address: %s", address)
		return
	}
	if h.orgRooms[orgID] == nil {
		h.orgRooms[orgID] = make(map[string]*Connection)
	}
	h.orgRooms[orgID][address] = conn
	log.Printf("Address %s joined organization room: %s", address, orgID)
}

// LeaveOrganizationRoom removes a connection from an organization room.
func (h *Hub) LeaveOrganizationRoom(orgID, address string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if members, ok := h.orgRooms[orgID]; ok {
		delete(members, address)
		if len(members) == 0 {
			delete(h.orgRooms, orgID)
		}
		log.Printf("Address %s left organization room: %s", address, orgID)
	}
}

// NotifyUser sends a message to a specific user connection.
func (h *Hub) NotifyUser(address string, message interface{}) {
	h.mu.RLock()
	conn, ok := h.connections[address]
	h.mu.RUnlock()
	if !ok {
		log.Printf("No connection for address: %s", address)
		return
	}
	if err := conn.Conn.WriteJSON(message); err != nil {
		log.Printf("Error sending message to %s: %v", address, err)
	}
}

// BroadcastOrganization sends a message to all connections in an organization room.
func (h *Hub) BroadcastOrganization(orgID string, message interface{}) {
	h.mu.RLock()
	members, ok := h.orgRooms[orgID]
	h.mu.RUnlock()
	if !ok {
		log.Printf("No room found for organization: %s", orgID)
		return
	}
	for addr, conn := range members {
		if err := conn.Conn.WriteJSON(message); err != nil {
			log.Printf("Error sending message to %s in room %s: %v", addr, orgID, err)
		}
	}
}
