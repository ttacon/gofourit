package gofourit

import (
	"log"
	"time"

	"github.com/bsm/redislock"
)

// redisSource is our implementation for a Redis based RemoteSource.
type redisSource struct {
	locker *redislock.Client
}

// TODO(ttacon): should support log options, e.g. for connection failures

// NewRedisRemoteSource returns a new RemoteSource based on the given redis
// client.
func NewRedisRemoteSource(client redislock.RedisClient) RemoteSource {
	return &redisSource{
		locker: redislock.New(client),
	}
}

// TODO(ttacon): should support options to GrabLock, e.g.:
//  - keepRenewingLock until released (this way we run a go routine renewing the
//    lock until it is released)
//  - TTL for lock
//  - option key prefixes for namespacing

// GrabLock acquires a unique lock based on the given key.
func (r *redisSource) GrabLock(key string) Lock {
	lock, err := r.locker.Obtain(key, 15*time.Second, nil)
	if err == redislock.ErrNotObtained {
		return nil
	} else if err != nil {
		// TODO(ttacon): replace with a call to the provided logger,
		// if it exists.
		log.Println(err)
		return nil
	}

	return &redisLock{
		lock: lock,
	}
}

// redisLock is our wrapping implementation to allow us to release the
// generated redislock Lock.
type redisLock struct {
	lock *redislock.Lock
}

// Release releases the given lock.
func (r *redisLock) Release() {
	r.lock.Release()
}
