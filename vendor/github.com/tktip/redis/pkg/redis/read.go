package redis

import (
	"encoding"
	"encoding/json"
	"github.com/go-redis/redis"
	"gopkg.in/yaml.v2"
)

type tipRedisError string

func (e tipRedisError) Error() string { return string(e) }

const (
	ErrorUnsupportedDataType = tipRedisError("unsupported target interface")
)

func (r *DefaultHandler) get(key string) *redis.StringCmd {
	client := r.getConnection()
	cmd := client.Get(r.appendKeyPrefix(key))
	return cmd
}

func (r *DefaultHandler) GetAndUnmarshalJSON(key string, target interface{}) error {
	val, err := r.GetBytes(key)
	if err != nil {
		return err
	}

	return json.Unmarshal(val, target)
}

func (r *DefaultHandler) GetAndUnmarshalYAML(key string, target interface{}) error {
	val, err := r.GetBytes(key)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(val, target)
}

func (r *DefaultHandler) GetAndUnmarshalBinary(key string, target encoding.BinaryUnmarshaler) error {
	return r.GetAndScanTo(key, target)
}

func (r *DefaultHandler) GetAndScanTo(key string, target interface{}) error {
	if !supportedReadDataType(target) {
		return ErrorUnsupportedDataType
	}

	cmd := r.get(key)
	if cmd.Err() != nil {
		return cmd.Err()
	}

	err := cmd.Scan(target)
	return err
}

func (r *DefaultHandler) GetBytes(key string) ([]byte, error) {
	cmd := r.get(key)
	err := cmd.Err()
	if err != nil {
		return nil, err
	}
	return cmd.Bytes()
}

func (r *DefaultHandler) GetString(key string) (string, error) {
	cmd := r.get(key)
	if cmd.Err() != nil {
		return "", cmd.Err()
	}

	return cmd.Val(), cmd.Err()
}
