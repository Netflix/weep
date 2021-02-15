package cache

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/netflix/weep/creds"
	"github.com/netflix/weep/errors"
	log "github.com/sirupsen/logrus"
)

const BUCKET = "credentials"

type FileDB struct {
	db *bolt.DB
}

func NewFileCache() (*FileDB, error) {
	db, err := bolt.Open("weep.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	fdb := &FileDB{
		db: db,
	}
	err = fdb.setup()
	if err != nil {
		return nil, err
	}

	return fdb, nil
}

func (f *FileDB) setup() error {
	err := f.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BUCKET))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (f *FileDB) Get(role string, assumeChain []string) (*creds.RefreshableProvider, error) {
	log.WithFields(log.Fields{
		"role":        role,
		"assumeChain": assumeChain,
		"cacheType":   "file",
	}).Info("retrieving credentials")
	c, err := f.get(getSlug(role, assumeChain))
	if err != nil {
		return nil, errors.NoCredentialsFoundInCache
	}
	return c, nil
}

func (f *FileDB) GetOrSet(client *creds.Client, role string, region string, assumeChain []string) (*creds.RefreshableProvider, error) {
	c, err := f.Get(role, assumeChain)
	if err == nil {
		return c, nil
	}
	log.Debugf("no credentials for %s in cache, creating", role)

	c, err = f.set(client, role, region, assumeChain)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (f *FileDB) SetDefault(client *creds.Client, role string, region string, assumeChain []string) error {
	// TODO
	return nil
}

func (f *FileDB) GetDefault() (*creds.RefreshableProvider, error) {
	// TODO
	return nil, nil
}

func (f *FileDB) get(slug string) (*creds.RefreshableProvider, error) {
	credentials := &creds.RefreshableProvider{}
	err := f.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET))
		result := b.Get([]byte(slug))
		err := json.Unmarshal(result, credentials)
		if err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		return credentials, err
	}
	err = credentials.EnsureRefreshed()
	if err != nil {
		return credentials, err
	}
	return credentials, nil
}

func (f *FileDB) set(client *creds.Client, role, region string, assumeChain []string) (*creds.RefreshableProvider, error) {
	c, err := creds.NewRefreshableProvider(client, role, region, assumeChain, false)
	if err != nil {
		return nil, fmt.Errorf("could not generate creds: %w", err)
	}
	data, err := json.Marshal(c)
	slug := getSlug(role, assumeChain)
	if err != nil {
		return nil, fmt.Errorf("could not marshal creds: %w", err)
	}
	err = f.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET))
		err := b.Put([]byte(slug), data)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}
