package cache

import (
	"strings"
	"sync"
	"time"

	"github.com/jumpserver/koko/pkg/logger"
	"github.com/jumpserver/koko/pkg/srvconn"
)

type UserSSHClient struct {
	ID      string // userID_assetID_systemUserID_systemUsername
	clients map[*srvconn.SSHClient]int64
	sync.Mutex
}

func (u *UserSSHClient) AddClient(client *srvconn.SSHClient) {
	u.Lock()
	defer u.Unlock()
	u.clients[client] = time.Now().UnixNano()
	logger.Infof("Store new client(%s) remain %d", client, len(u.clients))
}

func (u *UserSSHClient) DeleteClient(client *srvconn.SSHClient) {
	u.Lock()
	defer u.Unlock()
	delete(u.clients, client)
	logger.Infof("Remove client(%s) remain %d", client, len(u.clients))
}

func (u *UserSSHClient) GetClient() *srvconn.SSHClient {
	u.Lock()
	defer u.Unlock()
	if len(u.clients) == 0 {
		return nil
	}

	var client *srvconn.SSHClient
	var latest int64
	for item, timestamp := range u.clients {
		if timestamp > latest {
			latest = timestamp
			client = item
		}
	}
	return client

}

func (u *UserSSHClient) count() int {
	u.Lock()
	defer u.Unlock()
	return len(u.clients)
}

type SSHManager struct {
	data map[string]*UserSSHClient
	mu   sync.Mutex
}

func (s *SSHManager) getClientFromCache(key string) (*srvconn.SSHClient, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if userClient, ok := s.data[key]; ok {
		client := userClient.GetClient()
		if client != nil {
			return client, true
		}
	}
	return nil, false
}

func (s *SSHManager) AddClientCache(key string, client *srvconn.SSHClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if userClient, ok := s.data[key]; ok {
		userClient.AddClient(client)
	} else {
		userClient = &UserSSHClient{
			ID:      key,
			clients: make(map[*srvconn.SSHClient]int64),
		}
		userClient.AddClient(client)
		s.data[key] = userClient
		logger.Infof("Add new user cache current count: %d", len(s.data))
	}
}

func (s *SSHManager) deleteClientFromCache(key string, client *srvconn.SSHClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if userClient, ok := s.data[key]; ok {
		userClient.DeleteClient(client)
		if userClient.count() == 0 {
			delete(s.data, key)
			logger.Infof("Delete user cache current count: %d", len(s.data))
		}
	}
}

func (s *SSHManager) searchSSHClientFromCache(prefixKey string) (client *srvconn.SSHClient, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, userClient := range s.data {
		if strings.HasPrefix(key, prefixKey) {
			client := userClient.GetClient()
			if client != nil {
				return client, true
			}
		}
	}
	return nil, false
}
