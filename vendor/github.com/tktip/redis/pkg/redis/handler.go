package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

//TODO: Use 'New(ctx context.Context) *DefaultHandler'? causes close on context end
//TODO: New could also support sigint.
type DefaultHandler struct {
	Debug      bool          `yaml:"debug"`
	Prefix     string        `yaml:"prefix"`
	Address    string        `yaml:"address"`
	Db         int           `yaml:"db"`
	Password   string        `yaml:"password"`
	DefaultTTL time.Duration `yaml:"defaultTtl"`
	client     *redis.Client
}

func (r *DefaultHandler) appendKeyPrefix(key string) string {
	return fmt.Sprintf("%s-%s", r.Prefix, key)
}

func (r *DefaultHandler) getConnection() *redis.Client {
	if r.client == nil {
		r.client = redis.NewClient(&redis.Options{
			Addr:     r.Address,
			Password: r.Password, // no password set
			DB:       r.Db,
		})
	}

	return r.client
}

