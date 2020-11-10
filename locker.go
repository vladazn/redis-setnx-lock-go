package locker

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisLocker struct {
	Client *redis.Client
	Ctx    context.Context
	Key    string
	Value  string
}

func (r RedisLocker) lock() error {
	timeLimit := make(chan bool)
	time.AfterFunc(10*time.Second, func() {
		timeLimit <- true
	})
	for {
		select {
		case <-timeLimit:
			return errors.New("time limit exceeded")
		default:
			set, err := r.Client.SetNX(r.Ctx, "lock_"+r.Key, "locked", 10*time.Second).Result()
			if set == true {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}

}

func (r RedisLocker) unlock() error {
	return r.Client.Del(r.Ctx, "lock_"+r.Key).Err()
}

func (r RedisLocker) sync() (string, error) {
	val, err := r.Client.Get(r.Ctx, r.Key).Result()
	if err != nil {
		return "", err
	}
	r.Value = val
	return val, nil
}

func (r RedisLocker) save() error {
	return r.Client.Set(r.Ctx, r.Key, r.Value, 2*time.Hour).Err()
}

func (r RedisLocker) LockAndSync() (string, error) {
	err := r.lock()
	if err != nil {
		return "", err
	}
	val, _ := r.sync()
	return val, nil
}

func (r RedisLocker) UpdateAndUnlock() error {
	err := r.save()
	if err != nil {
		return err
	}
	return r.unlock()
}

