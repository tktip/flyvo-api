package redis

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"encoding/json"
	"github.com/go-redis/redis"
	"gopkg.in/yaml.v2"
	"regexp"
	"time"
)

//TODO: VERIFY THAT WORKS

type MockHandler struct {
	Handler
	Cache *map[string][]byte
}

func (d *MockHandler) get(key string) ([]byte, bool) {
	val, ok := (*d.Cache)[key]
	return val, ok
}

//GET
func (d *MockHandler) GetAndUnmarshalJSON(key string, target interface{}) error {
	v, ok := d.get(key)
	if !ok {
		return redis.Nil
	}
	return json.Unmarshal(v, target)
}

func (d *MockHandler) GetAndUnmarshalYAML(key string, target interface{}) error {
	v, ok := d.get(key)
	if !ok {
		return redis.Nil
	}
	return yaml.Unmarshal(v, target)
}

func (d *MockHandler) GetAndUnmarshalBinary(key string, target encoding.BinaryUnmarshaler) error {
	v, ok := d.get(key)
	if !ok {
		return redis.Nil
	}

	return target.UnmarshalBinary(v)
}

func (d *MockHandler) GetAndScanTo(key string, target interface{}) error {
	if !supportedReadDataType(target) {
		return ErrorUnsupportedDataType
	}
	v, ok := d.get(key)
	if !ok {
		return redis.Nil
	}

	buf := bytes.NewBuffer(v)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(target)
	if err != nil {
		return err
	}
	return nil
}

func (d *MockHandler) GetBytes(key string) ([]byte, error) {
	v, ok := d.get(key)
	if !ok {
		return nil, redis.Nil
	}
	return v, nil
}

func (d *MockHandler) GetString(key string) (string, error) {
	v, ok := d.get(key)
	if !ok {
		return "", redis.Nil
	}
	return string(v), nil
}

func (d *MockHandler) WriteJSONObject(key string, value interface{}, TTL ...time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	(*d.Cache)[key] = val
	return nil
}

func (d *MockHandler) WriteYAMLObject(key string, value interface{}, TTL ...time.Duration) error {
	val, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	(*d.Cache)[key] = val
	return nil
}

func (d *MockHandler) WriteAsBinary(key string, value encoding.BinaryMarshaler, TTL ...time.Duration) error {
	data, err := value.MarshalBinary()
	if err != nil {
		return err
	}
	(*d.Cache)[key] = data
	return nil
}

func (d *MockHandler) Write(key string, value interface{}, TTL ...time.Duration) error {
	if !supportedWriteDataType(value) {
		return ErrorUnsupportedDataType
	}

	switch v := value.(type) {
	case encoding.BinaryMarshaler:
		return d.WriteAsBinary(key, v)
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return nil
	}
	(*d.Cache)[key] = buf.Bytes()
	return nil
}

func (d *MockHandler) Close() error {
	return nil
}

func (d *MockHandler) GetKeysForRegex(r string) ([]string, error) {
	res := []string{}
	for key, _ := range *d.Cache {
		ok, err := regexp.Match(r, []byte(key))
		if err != nil {
			return nil, err
		} else if ok {
			res = append(res, key)
		}
	}
	return res, nil
}

func (d *MockHandler) ExireAfterTTL(key string, newTTL time.Duration) error {
	_, ok := d.get(key)
	if !ok {
		return redis.Nil
	}
	return nil
}

func (d *MockHandler) ExpireAt(key string, dateOfExpiry time.Time) error {
	_, ok := d.get(key)
	if !ok {
		return redis.Nil
	}
	return nil
}

func (d *MockHandler) Ping() error {
	return nil
}
