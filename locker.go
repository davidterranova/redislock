// Created by davidterranova on 30/01/2018.

package redis_lock

import (
	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"time"
	"context"
	"errors"
)

const (
	LOCK_SUFFIX = ".rdlock"
	DEFAULT_LOCK_TTL = time.Duration(20)*time.Second
)

var (
	ErrAlreadyLocked = errors.New("lock already taken")
)

// Locker struct
type Locker struct {
	redis *redis.Client
	locked bool
	key string
	lock string
	lockTTL time.Duration
}

// NewLocker : create a lock on a key with a *redis.Client connection
func NewLocker(client *redis.Client, key string) *Locker {
	return &Locker{
		redis : client,
		locked: false,
		key: key,
		lockTTL: DEFAULT_LOCK_TTL,
	}
}

// Lock : acquire lock, context aware
func (l *Locker) Lock(ctx context.Context) error {
	if l.locked {
		return ErrAlreadyLocked
	}
	var lock = uuid.NewV4().String()
	log.WithFields(log.Fields{
		"key":  l.key + LOCK_SUFFIX,
		"lock":  lock,
	}).Debug("locking")
	var ok bool = false
	var err error
	for !ok {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ok, err = l.redis.SetNX(l.key+LOCK_SUFFIX, lock, l.lockTTL).Result()
			if err != nil {
				return err
			}
		}
	}
	l.locked = true
	l.lock = lock
	log.WithFields(log.Fields{
		"key":  l.key + LOCK_SUFFIX,
		"lock":  lock,
	}).Debug("acquired")
	return err
}

// Unlock : release lock
func (l *Locker) Unlock() error {
	var err error
	if l.locked {
		log.WithFields(log.Fields{
			"key":   l.key + LOCK_SUFFIX,
			"lock":  l.lock,
		}).Debug("unlocking")
		var unlock = redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
      return redis.call("del", KEYS[1])
    else
      return 0
    end
	`)
		_, err = unlock.Run(l.redis, []string{l.key + LOCK_SUFFIX}, l.lock).Result()
		l.locked = false
	}
	return err
}

// SetLockTTL : specify lock TTL
func (l *Locker) SetLockTTL(d time.Duration) *Locker {
	l.lockTTL = d
	return l
}