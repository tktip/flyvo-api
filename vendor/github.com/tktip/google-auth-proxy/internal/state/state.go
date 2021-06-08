package state

import (
	"net/http"
	"time"

	redis "github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tktip/google-auth-proxy/internal/httperr"
	tipRedis "github.com/tktip/redis/pkg/redis"
)

const (
	blacklistPrefix = "blacklist-"
	oidcStatePrefix = "oidc-state-"
)

// Handler - redis handler
type Handler struct {
	Redis                tipRedis.DefaultHandler `yaml:"redis"`
	StateKeyValidMinutes time.Duration           `yaml:"StateKeyValidMinutes"`
}

// BlacklistKey blacklists a given sessionID for the given time in minutes
func (r *Handler) BlacklistKey(jti string, expTimeout time.Duration) httperr.Error {
	jti = blacklistPrefix + jti
	_, err := r.Redis.GetBytes(jti)
	if err != nil && err != redis.Nil {
		return httperr.New(http.StatusInternalServerError, err).
			WithCall("redis get to check if jti exists")
	}

	if err != redis.Nil {
		logrus.Infof("client tried to blacklist already blacklisted jti %s", jti)
		return nil
	}

	err = r.Redis.Write(jti, true, expTimeout)
	if err != nil {
		return httperr.New(http.StatusInternalServerError, err).WithCall("redis write")
	}
	logrus.Debugf("blacklisted sid %s for duration of '%f'", jti, expTimeout.Seconds())
	return nil
}

// IsBlacklisted checks if a given sessionID is blacklisted
func (r *Handler) IsBlacklisted(jti string) (bool, httperr.Error) {
	v, err := r.Redis.GetKeysForRegex(blacklistPrefix + jti)
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false,
			httperr.New(http.StatusInternalServerError, err).WithCall("redis get keys for regex")
	}

	if len(v) > 0 {
		logrus.Debugf("jti %s is blacklisted", jti)
		return true, nil
	}
	return false, nil
}

// CreateRedirState puts the given redirect uri in the configured redis cluster and returns the key
// it was stored on
func (r *Handler) CreateRedirState(redir string) (key string, err error) {
	key = uuid.New().String()

	redisKey := oidcStatePrefix + key

	// err = r.Redis.Write(redisKey, redir, time.Minute*time.Duration(r.StateKeyValidMinutes))
	err = r.Redis.Write(redisKey, redir, r.StateKeyValidMinutes)
	if err != nil {
		return key, err
	}

	return key, err
}

// GetRedirState fetches and returns a state-object from the configured redis-cluster stored on the
// given key
func (r *Handler) GetRedirState(key string) (redir string, exists bool, err error) {
	redisKey := oidcStatePrefix + key

	redir, err = r.Redis.GetString(redisKey)
	if err == redis.Nil {
		err = nil
		exists = false
		return
	} else if err != nil {
		return
	}
	exists = true
	return
}

// PopState deletes a state-object stored in the configured redis-cluster
func (r *Handler) PopState(key string) error {
	redisKey := oidcStatePrefix + key
	err := r.Redis.Expire(redisKey)
	if err != nil {
		logrus.Error(err)
		return err
	}
	logrus.Debugf("StateKey %s popped", redisKey)
	return nil
}
