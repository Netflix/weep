package cache

import (
	"github.com/boltdb/bolt"
	"github.com/netflix/weep/creds"
)

type FileDB struct {
	db *bolt.DB
}

func (f *FileDB) Get(role string, assumeChain []string) (*creds.RefreshableProvider, error) {
	return nil, nil
}

func (f *FileDB) GetOrSet(client *creds.Client, role string, region string, assumeChain []string) (*creds.RefreshableProvider, error) {
	return nil, nil
}

func (f *FileDB) SetDefault(client *creds.Client, role string, region string, assumeChain []string) error {
	return nil
}

func (f *FileDB) GetDefault() (*creds.RefreshableProvider, error) {
	return nil, nil
}
