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
	sync.Map
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
	c := &tokenCache{}
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
	c.Range(func(key, value interface{}) bool {
		if attr, ok := value.(*tokenAttributes); ok {
			if attr.Expiration.Before(time.Now()) {
				logging.Log.Debugf("deleting token with expiration %v", attr.Expiration)
				c.Delete(key)
			}
		} else {
			logging.Log.Debugf("deleting token with invalid attributes %v", value)
			c.Delete(key)
		}
		return true
	})
}

func (c *tokenCache) generateToken(role string, ttlSeconds int) string {
	token := randomString(64)
	c.Set(token, role, ttlSeconds)
	return token
}

func (c *tokenCache) checkToken(token string) (bool, int) {
	attr, err := c.Get(token)
	if err != nil {
		logging.Log.Warningf("invalid session token: %v", err)
		return false, 0
	}
	if attr.Expiration.Before(time.Now()) {
		logging.Log.Warning("session token is expired")
		return false, 0
	}
	remainingTtl := attr.Expiration.Sub(time.Now())
	return true, int(remainingTtl.Seconds())
}

func (c *tokenCache) Set(token, role string, ttl int) {
	expiration := time.Now().Add(time.Duration(ttl) * time.Second)
	attr := tokenAttributes{
		InitialTtl: ttl,
		Expiration: expiration,
		Role:       role,
	}
	c.Store(token, &attr)
}

func (c *tokenCache) Get(token string) (*tokenAttributes, error) {
	var value interface{}
	var ok bool
	var attr *tokenAttributes
	value, ok = c.Load(token)
	if !ok {
		return nil, errors.NoTokenFoundInCache
	}
	if attr, ok = value.(*tokenAttributes); ok {
		return attr, nil
	}
	return nil, errors.InvalidTokenFoundInCache
}
