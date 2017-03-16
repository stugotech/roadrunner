package server

import (
	"fmt"

	"net/http"

	"regexp"
	"strings"

	"github.com/docker/libkv/store"
	"github.com/stugotech/coyote/store"
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
	logger.Info("creating new server",
		golog.String("store", config.Store),
		golog.Strings("store-nodes", config.StoreNodes),
		golog.String("store-prefix", config.StorePrefix),
		golog.String("listen", config.Listen),
		golog.String("path-prefix", config.PathPrefix),
	)

	store, err := store.NewStore(config.Store, config.StoreNodes, config.StorePrefix)
	if err != nil {
		return nil, logger.Errore(err)
	}

	config.PathPrefix = strings.Trim(config.PathPrefix, "/")
	validPath := regexp.MustCompile(fmt.Sprintf(pathRegexp, config.PathPrefix))

	return &serverInfo{
		config:    config,
		store:     store,
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
	challenge, err := s.store.GetChallenge(key)

	if err != nil {
		logger.Error("error getting value", golog.String("url", request.URL.Path), golog.String("key", key))
		logger.Errore(err)
		http.NotFound(response, request)
		return
	}

	response.Write([]byte(challenge.Value))

	if err = s.store.DeleteChallenge(key); err != nil {
		logger.Error("error deleting key", golog.String("key", key))
		logger.Errore(err)
	}
}

func (s *serverInfo) makeHandler(fn serverInfoHandler) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		fn(s, response, request)
	}
}
