package server

import (
	"fmt"
	"path"

	"net/http"

	"regexp"
	"strings"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/stugotech/goconfig"
	"github.com/stugotech/golog"
)

const (
	kvPrefix   = "challenges"
	pathRegexp = "^/%s/([a-zA-Z0-9_-]+)$"
)

// The following consts define config keys for this module
const (
	ListenKey      = "listen"
	StoreKey       = "store"
	StoreNodesKey  = "store-nodes"
	StorePrefixKey = "store-prefix"
	PathPrefixKey  = "path-prefix"
)

// Config describes the configuration settings for the server
type Config struct {
	Store       string
	StoreNodes  []string
	StorePrefix string
	Listen      string
	PathPrefix  string
}

type serverInfo struct {
	config    *Config
	store     store.Store
	validPath *regexp.Regexp
}

type serverInfoHandler func(s *serverInfo, response http.ResponseWriter, request *http.Request)

// Server describes an ACME challenge server
type Server interface {
	Listen() error
}

var logger = golog.NewPackageLogger()

// ReadConfig reads a configuration from a config provider
func ReadConfig(config goconfig.Config) *Config {
	return &Config{
		Store:       config.GetString(StoreKey),
		StoreNodes:  config.GetStringSlice(StoreNodesKey),
		StorePrefix: config.GetString(StorePrefixKey),
		Listen:      config.GetString(ListenKey),
		PathPrefix:  config.GetString(PathPrefixKey),
	}
}

// NewServer creates a new server
func NewServer(config *Config) (Server, error) {
	etcd.Register()
	consul.Register()
	boltdb.Register()
	zookeeper.Register()

	logger.Info("creating new server",
		golog.String("store", config.Store),
		golog.Strings("store-nodes", config.StoreNodes),
		golog.String("store-prefix", config.StorePrefix),
		golog.String("listen", config.Listen),
		golog.String("path-prefix", config.PathPrefix),
	)

	storeConfig := &store.Config{}
	s, err := libkv.NewStore(store.Backend(config.Store), config.StoreNodes, storeConfig)

	if err != nil {
		return nil, logger.Errore(err)
	}

	config.PathPrefix = strings.Trim(config.PathPrefix, "/")
	validPath := regexp.MustCompile(fmt.Sprintf(pathRegexp, config.PathPrefix))

	return &serverInfo{
		config:    config,
		store:     s,
		validPath: validPath,
	}, nil
}

// Listen starts the server listening for connections
func (s *serverInfo) Listen() error {
	http.HandleFunc("/", s.makeHandler(challengeHandler))
	logger.Info("server listening", golog.String("interface", s.config.Listen))
	err := http.ListenAndServe(s.config.Listen, nil)

	if err != nil {
		return logger.Errore(err)
	}

	return nil
}

func challengeHandler(s *serverInfo, response http.ResponseWriter, request *http.Request) {
	match := s.validPath.FindStringSubmatch(request.URL.Path)
	if match == nil {
		logger.Error("invalid URL format", golog.String("path", request.URL.Path))
		http.NotFound(response, request)
		return
	}

	key := match[1]
	value, err := s.getValue(key)

	if err != nil {
		logger.Error("error getting value", golog.String("url", request.URL.Path), golog.String("key", key))
		logger.Errore(err)
		http.NotFound(response, request)
		return
	}

	response.Write(value)

	if err = s.deleteKey(key); err != nil {
		logger.Error("error deleting key", golog.String("key", key))
		logger.Errore(err)
	}
}

func (s *serverInfo) makeHandler(fn serverInfoHandler) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		fn(s, response, request)
	}
}

func (s *serverInfo) getValue(key string) ([]byte, error) {
	key = s.getKeyName(key)
	logger.Debug("looking up KV store for value", golog.String("key", key))
	kv, err := s.store.Get(key)

	if err != nil {
		return nil, logger.Errore(err)
	}

	return kv.Value, nil
}

func (s *serverInfo) deleteKey(key string) error {
	key = s.getKeyName(key)
	logger.Debug("deleting key from KV store", golog.String("key", key))
	err := s.store.Delete(key)

	if err != nil {
		return logger.Errore(err)
	}

	return nil
}

func (s *serverInfo) getKeyName(key string) string {
	return path.Join(s.config.StorePrefix, kvPrefix, key)
}
