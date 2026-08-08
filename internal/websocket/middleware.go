package websocket

import (
	"context"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type PeerRole string

const (
	PeerRoleBank   PeerRole = "bank"
	PeerRoleClient PeerRole = "client"
)

type Peer struct {
	BankID    string
	Role      PeerRole
	SessionID string
	Conn      *websocket.Conn
}

type Message struct {
	SessionID           string `json:"session_id,omitempty"`
	Type                string `json:"type"`
	EncapsulationKeyB64 string `json:"ek,omitempty"`
	PrivateKeyID        string `json:"pkid,omitempty"`
	CiphertextB64       string `json:"ct,omitempty"`
	SignatureB64        string `json:"sig,omitempty"`
	NonceB64            string `json:"nonce,omitempty"`
	MacB64              string `json:"mac,omitempty"`
}

type Manager struct {
	mu      sync.RWMutex
	banks   map[string]*Peer
	clients map[string]map[string]*Peer
}

func NewWebSocketManager() *Manager {
	return &Manager{
		banks: make(map[string]*Peer),
		clients: make(
			map[string]map[string]*Peer,
		),
	}
}

func (m *Manager) Add(peer *Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if peer.Role == PeerRoleBank {
		old := m.banks[peer.BankID]

		if old != nil {
			_ = old.Conn.Close(
				websocket.StatusNormalClosure,
				"replaced by new connection",
			)
		}

		m.banks[peer.BankID] = peer
		return
	}

	if _, ok := m.clients[peer.BankID]; !ok {
		m.clients[peer.BankID] = make(map[string]*Peer)
	}

	m.clients[peer.BankID][peer.SessionID] = peer
}

func (m *Manager) Remove(peer *Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if peer.Role == PeerRoleBank {
		if current, ok := m.banks[peer.BankID]; ok {
			if current == peer {
				delete(
					m.banks,
					peer.BankID,
				)
			}
		}
		return
	}

	if clients, ok := m.clients[peer.BankID]; ok {
		if current, ok := clients[peer.SessionID]; ok {
			if current == peer {
				delete(
					clients,
					peer.SessionID,
				)
			}
		}

		if len(clients) == 0 {
			delete(
				m.clients,
				peer.BankID,
			)
		}
	}
}

func (m *Manager) Forward(sender *Peer, message Message) {
	var target *Peer

	m.mu.RLock()

	if sender.Role == PeerRoleClient {
		target = m.banks[sender.BankID]
		message.SessionID = sender.SessionID
	} else {
		if clients, ok := m.clients[sender.BankID]; ok {
			target = clients[message.SessionID]
			message.SessionID = ""
		}
	}

	m.mu.RUnlock()

	if target == nil {
		return
	}

	_ = wsjson.Write(
		context.Background(),
		target.Conn,
		message,
	)
}
