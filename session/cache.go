package session

import (
	"crypto/rand"
	"github.com/netflix/weep/errors"
	"math/big"
	"sync"
	"time"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

type Cache struct {
	sync.RWMutex
	TokenMap
}

type TokenAttributes struct {
	InitialTtl int
	Expiration time.Time
	Role       string
}

type TokenMap map[string]*TokenAttributes

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

func CreateCache() *Cache {
	c := &Cache{
		TokenMap: make(map[string]*TokenAttributes),
	}
	go c.startWatcher()
	return c
}

func (c *Cache) startWatcher() {
	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case _ = <-ticker.C:
			c.clean()
		}
	}
}

func (c *Cache) clean() {
	for token, attr := range c.TokenMap {
		if attr.Expiration.Before(time.Now()) {
			log.Debugf("deleting token with expiration %v", attr.Expiration)
			c.delete(token)
		}
	}
}

func (c *Cache) generateToken(role string, ttlSeconds int) string {
	token := randomString(64)
	c.Set(token, role, ttlSeconds)
	return token
}

func (c *Cache) checkToken(token string) (bool, int) {
	attr, err := sessions.Get(token)
	if err != nil {
		log.Warning("invalid session token")
		return false, 0
	}
	if attr.Expiration.Before(time.Now()) {
		log.Warning("session token is expired")
		return false, 0
	}
	remainingTtl := time.Now().Sub(attr.Expiration)
	return true, int(remainingTtl.Seconds())
}

func (c *Cache) delete(token string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.TokenMap[token]; ok {
		delete(c.TokenMap, token)
	}
}

func (c *Cache) Set(token, role string, ttl int) {
	expiration := time.Now().Add(time.Duration(ttl) * time.Second)
	attr := TokenAttributes{
		InitialTtl: ttl,
		Expiration: expiration,
		Role:       role,
	}
	c.Lock()
	defer c.Unlock()
	c.TokenMap[token] = &attr
}

func (c *Cache) Get(token string) (*TokenAttributes, error) {
	c.RLock()
	defer c.RUnlock()
	attr, ok := c.TokenMap[token]
	if !ok {
		return nil, errors.NoTokenFoundInCache
	}
	return attr, nil
}
