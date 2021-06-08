package redis

import (
	"encoding"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"time"
)

func (r *DefaultHandler) write(key string, value interface{}, TTL ...time.Duration) (err error) {
	_TTL := r.DefaultTTL
	if len(TTL) > 0 {
		_TTL = TTL[0]
	}

	client := r.getConnection()

	key = r.appendKeyPrefix(key)

	redisStatus := client.Set(key, value, _TTL)
	if redisStatus.Err() != nil && redisStatus.Err().Error() != "" {
		return redisStatus.Err()
	}

	return redisStatus.Err()
}

func (r *DefaultHandler) WriteJSONObject(key string, value interface{}, TTL ...time.Duration) (err error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.write(key, bytes, TTL...)
}

func (r *DefaultHandler) WriteYAMLObject(key string, value interface{}, TTL ...time.Duration) (err error) {
	bytes, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return r.write(key, bytes, TTL...)
}

func (r *DefaultHandler) WriteAsBinary(key string, value encoding.BinaryMarshaler, TTL ...time.Duration) (err error) {
	return r.write(key, value, TTL...)
}

func (r *DefaultHandler) Write(key string, value interface{}, TTL ...time.Duration) error {
	if !supportedWriteDataType(value) {
		return ErrorUnsupportedDataType
	}
	return r.write(key, value, TTL...)
}
