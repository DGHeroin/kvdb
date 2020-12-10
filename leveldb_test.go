package kvdb

import (
    "fmt"
    "os"
    "testing"
    "time"
)

func TestOpenDB(t *testing.T) {
    dir := fmt.Sprintf("%s%cdb", os.TempDir(), os.PathSeparator)
    t.Log("db dir path", dir)
    db, err := OpenDB(dir, nil)
    if err != nil {
        t.Error(err)
        return
    }
    defer db.Close()
    // 写 100k 次
    var (
        n           = 100 * 1000
        startTime   time.Time
        elapsedTime time.Duration
    )
    var (
        key = []byte("test-key")
        val = []byte("test-value")
    )
    startTime = time.Now()
    for i := 0; i < n; i++ {
        db.Put(key, val)
    }
    elapsedTime = time.Since(startTime)
    t.Logf("leveldb-lru put %v time %v %v", n, elapsedTime, int64(time.Second/(elapsedTime/time.Duration(n))))

    // 读
    startTime = time.Now()
    for i := 0; i < n; i++ {
        db.Get(key)
    }
    elapsedTime = time.Since(startTime)
    t.Logf("leveldb-lru get %v time %v %v", n, elapsedTime, int64(time.Second/(elapsedTime/time.Duration(n))))
}
func TestDB(t *testing.T)  {
    dir := fmt.Sprintf("%s%cdb", os.TempDir(), os.PathSeparator)
    t.Log("db dir path", dir)
    db, err := OpenDB(dir, nil)
    if err != nil {
        t.Error(err)
        return
    }
    defer db.Close()

    var (
        key = []byte("test-key")
        val = []byte("test-value")
    )
    db.Put(key, val)

    if data, err := db.Get(key); err == nil {
        t.Log("get value", string(data))
    } else {
        t.Error(err)
    }
}
