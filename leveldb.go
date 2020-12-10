package kvdb

import (
    "github.com/syndtr/goleveldb/leveldb"
)

type wrapperLevelDB struct {
    db    *leveldb.DB
    cache Cache
}

func defaultOption() *Option {
    return &Option{
        MaxCacheItem:   1000,
        MaxCacheMemory: 64 * 1024 * 1024, // 64MB
    }
}
func OpenDB(dbDir string, opt *Option) (DB, error) {
    if opt == nil {
        opt = defaultOption()
    }
    db, err := leveldb.OpenFile(dbDir, nil)
    if err != nil {
        return nil, err
    }
    d := &wrapperLevelDB{
        db:    db,
        cache: NewCache(opt.MaxCacheItem, opt.MaxCacheMemory),
    }
    return d, nil
}

func (d *wrapperLevelDB) Put(key, value []byte) error {
    err := d.db.Put(key, value, nil)
    if err == nil {
        d.cache.Remove(key)
    }
    return err
}
func (d *wrapperLevelDB) Get(key []byte) ([]byte, error) {
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

func (d *wrapperLevelDB) Close() error {
    return d.db.Close()
}
