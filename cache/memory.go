package cache

import (
	"fmt"
	"sync"

	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/errors"
	log "github.com/sirupsen/logrus"
)

type InMemory struct {
	RoleCredentials map[string]*creds.RefreshableProvider
	DefaultRole     string
	mu              sync.RWMutex
}

func NewMemoryCache() *InMemory {
	return &InMemory{
		RoleCredentials: make(map[string]*creds.RefreshableProvider),
	}
}

func (cc *InMemory) Get(role string, assumeChain []string) (*creds.RefreshableProvider, error) {
	log.WithFields(log.Fields{
		"role":        role,
		"assumeChain": assumeChain,
	}).Info("retrieving credentials")
	c, ok := cc.get(getSlug(role, assumeChain))
	if ok {
		log.Debugf("found credentials for %s in cache", role)
		return c, nil
	}
	return nil, errors.NoCredentialsFoundInCache
}

func (cc *InMemory) GetOrSet(client *creds.Client, role, region string, assumeChain []string) (*creds.RefreshableProvider, error) {
	c, err := cc.Get(role, assumeChain)
	if err == nil {
		return c, nil
	}
	log.Debugf("no credentials for %s in cache, creating", role)

	c, err = cc.set(client, role, region, assumeChain)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (cc *InMemory) SetDefault(client *creds.Client, role, region string, assumeChain []string) error {
	_, err := cc.set(client, role, region, assumeChain)
	if err != nil {
		return err
	}
	cc.DefaultRole = getSlug(role, assumeChain)
	return nil
}

func (cc *InMemory) GetDefault() (*creds.RefreshableProvider, error) {
	if cc.DefaultRole == "" {
		return nil, errors.NoDefaultRoleSet
	}
	c, ok := cc.get(cc.DefaultRole)
	if ok {
		return c, nil
	}
	return nil, errors.NoCredentialsFoundInCache
}

func (cc *InMemory) get(slug string) (*creds.RefreshableProvider, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	c, ok := cc.RoleCredentials[slug]
	return c, ok
}

func (cc *InMemory) set(client *creds.Client, role, region string, assumeChain []string) (*creds.RefreshableProvider, error) {
	c, err := creds.NewRefreshableProvider(client, role, region, assumeChain, false)
	if err != nil {
		return nil, fmt.Errorf("could not generate creds: %w", err)
	}
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.RoleCredentials[getSlug(role, assumeChain)] = c
	return c, nil
}
