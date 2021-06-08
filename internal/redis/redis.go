package redis

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

//Connector provides access to redis
type Connector struct {
	URL      string        `yaml:"url"`
	Db       int           `yaml:"db"`
	Password string        `yaml:"password"`
	RedisTTL time.Duration `yaml:"ttl"`
}

func (r *Connector) getConnection() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     r.URL,
		Password: r.Password, // no password set
		DB:       r.Db,
	})
}

/*
WriteValue Write @value to @key in redis
*/
func (r *Connector) WriteValue(key string, value string, TTL ...time.Duration) (err error) {
	_TTL := r.RedisTTL
	if len(TTL) > 0 {
		_TTL = TTL[0]
	}

	client := r.getConnection()
	defer client.Close()

	key = "flyvo-" + key
	logrus.Info("Saving " + value + " to cache.")
	redisStatus := client.Set(key, value, _TTL)
	if redisStatus.Err() != nil && redisStatus.Err().Error() != "" {
		logrus.Errorf("Could not write: %s", redisStatus.Err().Error())
		return redisStatus.Err()
	}

	return nil
}

/*
GetValue returns the value from redis of given key if exists
*/
func (r *Connector) GetValue(key string) ([]byte, error) {
	client := r.getConnection()
	defer client.Close()

	key = "flyvo-" + key

	redisData := client.Get(key)
	logrus.Info(redis.Nil.Error())
	if redisData.Err() != nil && redisData.Err().Error() != "" &&
		redis.Nil.Error() != redisData.Err().Error() {
		logrus.Warn("Could not get data, error [" + redisData.Err().Error() + "]")
		return nil, redisData.Err()
	}

	byteData, err := redisData.Bytes()
	if err != nil && redis.Nil != err {
		logrus.Info("Error occured: " + err.Error())
		return nil, err
	} else if redisData.Val() == "" {
		logrus.Info("No value found")
		return nil, nil

	}

	//Reset TTL on retrieval, as data is still relevant.
	client.Expire(key, r.RedisTTL)

	return byteData, nil
}

/*
GetList returns a list of keys based on regex
*/
func (r *Connector) GetList(regex string) ([]string, error) {
	client := r.getConnection()
	defer client.Close()

	regex = "flyvo-" + regex

	cmd := client.Keys(regex)
	cmd.Val()
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

//GetStringValue - get string value of key
func (r *Connector) GetStringValue(key string) (string, error) {
	val, err := r.GetValue(key)
	if err != nil {
		return "", err
	}

	return string(val), nil
}

//DeleteRegex - deletes all keys/values with regex match
func (r *Connector) DeleteRegex(regex string) error {
	logrus.Debugf("Deleting regex '%s' from redis", regex)
	client := r.getConnection()
	defer client.Close()

	list, err := r.GetList(regex)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return nil
	}

	cmd := client.Del(list...)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	return nil
}
