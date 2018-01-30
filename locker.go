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
)

var (
	ErrAlreadyLocked = errors.New("lock already taken")
)

type Locker struct {
	redis *redis.Client
	locked bool
	key string
	lock string
}

func NewLocker(client *redis.Client, key string) *Locker {
	return &Locker{
		redis : client,
		locked: false,
		key: key,
	}
}

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
			ok, err = l.redis.SetNX(l.key+LOCK_SUFFIX, lock, time.Duration(20)*time.Second).Result()
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
