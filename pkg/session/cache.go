package session

import (
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/netflix/weep/pkg/logging"

	"github.com/netflix/weep/pkg/errors"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

type tokenCache struct {
	sync.RWMutex
	TokenMap
}

type tokenAttributes struct {
	InitialTtl int
	Expiration time.Time
	Role       string
}

type TokenMap map[string]*tokenAttributes

func randomString(n int) string {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			continue
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}

func createCache() *tokenCache {
	c := &tokenCache{
		TokenMap: make(map[string]*tokenAttributes),
	}
	go c.startWatcher()
	return c
}

func (c *tokenCache) startWatcher() {
	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case _ = <-ticker.C:
			c.clean()
		}
	}
}

func (c *tokenCache) clean() {
	for token, attr := range c.TokenMap {
		if attr.Expiration.Before(time.Now()) {
			logging.Log.Debugf("deleting token with expiration %v", attr.Expiration)
			c.delete(token)
		}
	}
}

func (c *tokenCache) generateToken(role string, ttlSeconds int) string {
	token := randomString(64)
	c.Set(token, role, ttlSeconds)
	return token
}

func (c *tokenCache) checkToken(token string) (bool, int) {
	attr, err := sessions.Get(token)
	if err != nil {
		logging.Log.Warning("invalid session token")
		return false, 0
	}
	if attr.Expiration.Before(time.Now()) {
		logging.Log.Warning("session token is expired")
		return false, 0
	}
	remainingTtl := time.Now().Sub(attr.Expiration)
	return true, int(remainingTtl.Seconds())
}

func (c *tokenCache) delete(token string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.TokenMap[token]; ok {
		delete(c.TokenMap, token)
	}
}

func (c *tokenCache) Set(token, role string, ttl int) {
	expiration := time.Now().Add(time.Duration(ttl) * time.Second)
	attr := tokenAttributes{
		InitialTtl: ttl,
		Expiration: expiration,
		Role:       role,
	}
	c.Lock()
	defer c.Unlock()
	c.TokenMap[token] = &attr
}

func (c *tokenCache) Get(token string) (*tokenAttributes, error) {
	c.RLock()
	defer c.RUnlock()
	attr, ok := c.TokenMap[token]
	if !ok {
		return nil, errors.NoTokenFoundInCache
	}
	return attr, nil
}
