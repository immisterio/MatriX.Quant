package torr

import (
	"sync"
	"time"
)

var (
	serversMu sync.Mutex
	servers   = make(map[string]*BTServer)

	cleanupOnce sync.Once
)

const idleServerTimeout = 15 * time.Minute

func startCleanupLoop() {
	cleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				cleanupStaleServers()
			}
		}()
	})
}

func cleanupStaleServers() {
	now := time.Now()

	serversMu.Lock()
	stale := make([]*BTServer, 0)
	for user, srv := range servers {
		if srv.lastUsed.IsZero() {
			srv.lastUsed = now
			continue
		}

		if now.Sub(srv.lastUsed) <= idleServerTimeout {
			continue
		}

		stale = append(stale, srv)
		delete(servers, user)
	}
	serversMu.Unlock()

	for _, srv := range stale {
		srv.Disconnect()
	}
}

func normalizeUser(user string) string {
	if user == "" {
		return "base"
	}
	return user
}

func getOrCreateServer(user string) *BTServer {
	startCleanupLoop()

	key := normalizeUser(user)
	now := time.Now()

	serversMu.Lock()

	if srv, ok := servers[key]; ok {
		srv.lastUsed = now
		serversMu.Unlock()
		return srv
	}

	srv := NewBTS()
	srv.lastUsed = now
	servers[key] = srv
	serversMu.Unlock()

	return srv
}

func ConnectServer(user string) (*BTServer, error) {
	srv := getOrCreateServer(user)
	srv.mu.Lock()
	connected := srv.client != nil
	srv.mu.Unlock()
	if connected {
		return srv, nil
	}
	return srv, srv.Connect()
}

func DisconnectAllServers() {
	serversMu.Lock()
	current := make(map[string]*BTServer, len(servers))
	for user, srv := range servers {
		current[user] = srv
	}
	serversMu.Unlock()

	for user, srv := range current {
		srv.Disconnect()
		serversMu.Lock()
		delete(servers, user)
		serversMu.Unlock()
	}
}

func ForEachServer(fn func(string, *BTServer)) {
	serversMu.Lock()
	current := make(map[string]*BTServer, len(servers))
	for user, srv := range servers {
		current[user] = srv
	}
	serversMu.Unlock()

	for user, srv := range current {
		fn(user, srv)
	}
}
