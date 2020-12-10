package kvdb

import (
    "github.com/syndtr/goleveldb/leveldb"
)

type DBi interface {
}

type DB struct {
    db    *leveldb.DB
    cache Cache
}

type Option struct {
    MaxCacheItem   uint64
    MaxCacheMemory uint64
}

func defaultOption() *Option {
    return &Option{
        MaxCacheItem:   1000,
        MaxCacheMemory: 64 * 1024 * 1024, // 64MB
    }
}
func OpenDB(dbDir string, opt *Option) (*DB, error) {
    if opt == nil {
        opt = defaultOption()
    }
    db, err := leveldb.OpenFile(dbDir, nil)
    if err != nil {
        return nil, err
    }
    d := &DB{
        db:    db,
        cache: NewCache(opt.MaxCacheItem, opt.MaxCacheMemory),
    }
    return d, nil
}

func (d *DB) Put(key, value []byte) error {
    d.cache.Add(key, value)
    return d.db.Put(key, value, nil)
}
func (d *DB) Get(key []byte) ([]byte, error) {
    val, ok := d.cache.Get(key)
    if ok {
        return val, nil
    }
    val, err := d.db.Get(key, nil)
    if err != nil {
        return nil, err
    }
    d.cache.Add(key, val)
    return val, nil
}

func (d *DB) Close() error {
    return d.db.Close()
}
