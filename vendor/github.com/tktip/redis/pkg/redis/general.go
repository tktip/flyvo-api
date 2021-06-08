package redis

import (
	"github.com/go-redis/redis"
	"time"
)

const (
	Nil = redis.Nil
)

// GetKeysForRegex returns the value from redis of given key if exists
func (r *DefaultHandler) GetKeysForRegex(regex string) ([]string, error) {
	client := r.getConnection()

	regex = r.appendKeyPrefix(regex)

	cmd := client.Keys(regex)
	cmd.Val()
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

func (r *DefaultHandler) ExireAfterTTL(key string, newTTL time.Duration) error {
	client := r.getConnection()

	key = r.appendKeyPrefix(key)
	cmd := client.Expire(key, newTTL)
	return cmd.Err()
}

func (r *DefaultHandler) Expire(key string) error {
	client := r.getConnection()

	key = r.appendKeyPrefix(key)
	cmd := client.Dump(key)
	return cmd.Err()
}

func (r *DefaultHandler) ExpireAt(key string, dateOfExpiry time.Time) error {
	client := r.getConnection()

	key = r.appendKeyPrefix(key)
	cmd := client.ExpireAt(key, dateOfExpiry)
	return cmd.Err()
}

func (r *DefaultHandler) Ping() error {
	client := r.getConnection()

	res := client.Ping()
	return res.Err()
}

//Only do this on shutdown.
//Redis client is supposed to be used by multiple
//functions and is 'safe for concurrent use'
func (r *DefaultHandler) Close() error {
	//TODO: Necessary to check err here? Can anything stall or is it safe to ignore?
	if r.client != nil {
		err := r.client.Close()
		if err != nil {
			return err
		}
	}

	r.client = nil
	return nil
}
