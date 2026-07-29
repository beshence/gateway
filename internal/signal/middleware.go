package signal

import (
	"sync"

	"golang.org/x/net/websocket"
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
	SessionID     string  `json:"session_id,omitempty"`
	Type          string  `json:"type"`
	SDP           string  `json:"sdp,omitempty"`
	Candidate     string  `json:"candidate,omitempty"`
	SDPMid        *string `json:"sdpmid,omitempty"`
	SDPMLineIndex *uint16 `json:"sdpmlineindex,omitempty"`
}

type Manager struct {
	mu      sync.RWMutex
	banks   map[string]*Peer
	clients map[string]map[string]*Peer
}

func NewSignalManager() *Manager {
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
			old.Conn.Close()
		}

		m.banks[peer.BankID] = peer
		return
	}

	if _, ok := m.clients[peer.BankID]; !ok {
		m.clients[peer.BankID] =
			make(map[string]*Peer)
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

	_ = websocket.JSON.Send(
		target.Conn,
		message,
	)
}
