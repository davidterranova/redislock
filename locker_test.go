// Created by davidterranova on 30/01/2018.

package redis_lock

import (
	"testing"
	"time"
	log "github.com/sirupsen/logrus"
	"os"
	"context"
	"github.com/go-redis/redis"
)

const(
	REDIS_URI = "localhost:6379"
)

func init() {
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
	log.Info("starting")
}

func TestLocker(t *testing.T) {
	client := redis.NewClient(
		&redis.Options{
			Addr:     REDIS_URI,
			Password: "",
			DB:       0,
		},
	)

	_, err := client.Ping().Result()
	if err != nil {
		t.Error(err)
	}

	var nb = 2

	var lockers = make([]*Locker, nb)
	for i := 0; i < len(lockers); i++ {
		lockers[i] = NewLocker(client, "key")
	}

	var unlocked = 0
	var ctx, _ = context.WithTimeout(context.Background(), time.Duration(3) * time.Second)

	var lu = func(l *Locker) {
		err := l.Lock(ctx)
		if err != nil {
			t.Error(err)
		}
		time.Sleep(2000 * time.Millisecond)
		err = l.Unlock()
		if err != nil {
			t.Error(err)
		}
		unlocked++
	}

	for i := 0; i < len(lockers); i++ {
		go lu(lockers[i])
	}

	// wait for all unlock
	for ; unlocked < len(lockers); {
		time.Sleep(10 * time.Millisecond)
	}
}