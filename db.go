package kvdb

type DB interface {
    Put(key, value []byte) error
    Get(key []byte) ([]byte, error)
    Close() error
}
type Option struct {
    MaxCacheItem   uint64
    MaxCacheMemory uint64
}
