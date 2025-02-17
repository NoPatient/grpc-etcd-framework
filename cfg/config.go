package cfg

import (
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"sync"
)

const (
	ETCDEndpoint   = "127.0.0.1:2379"
	ETCDConfigPath = "/config/server"
)

const (
	ServiceName             = "MyService"
	ServiceLimiterThreshold = 10
)

type ServerConfig struct {
	// etcdctl put /config/server '{"rate_limit": 5}'
	RateLimit int `json:"rate_limit"`
}

type ConfigManager struct {
	mu     sync.RWMutex
	config ServerConfig
}

func NewConfigManager(initialConfig ServerConfig) *ConfigManager {
	return &ConfigManager{
		config: initialConfig,
	}
}

func (cm *ConfigManager) UpdateConfig(newConfig ServerConfig) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = newConfig
	log.Printf("Configuration updated: %+v", cm.config)
}

func (cm *ConfigManager) GetConfig() ServerConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

func WatchConfigChanges(ctx context.Context, etcdCli *clientv3.Client, key string, updates chan<- ServerConfig) {
	research := etcdCli.Watch(ctx, key)
	for watchResponse := range research {
		for _, ev := range watchResponse.Events {
			var cfg ServerConfig
			err := json.Unmarshal(ev.Kv.Value, &cfg)
			if err != nil {
				log.Printf("Failed to unmarshal config: %v", err)
			}
			updates <- cfg
		}
	}
}
