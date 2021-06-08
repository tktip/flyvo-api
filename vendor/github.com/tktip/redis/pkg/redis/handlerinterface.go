package redis

import (
	"encoding"
	"time"
)

type Handler interface {
	//GET
	GetAndUnmarshalJSON(key string, target interface{}) error
	GetAndUnmarshalYAML(key string, target interface{}) error
	GetAndUnmarshalBinary(key string, target encoding.BinaryUnmarshaler) error
	GetAndScanTo(key string, target interface{}) error
	GetBytes(key string) ([]byte, error)
	GetString(key string) (string, error)

	//SET
	WriteJSONObject(key string, value interface{}, TTL ...time.Duration) error
	WriteYAMLObject(key string, value interface{}, TTL ...time.Duration) error
	WriteAsBinary(key string, value encoding.BinaryMarshaler, TTL ...time.Duration) error
	Write(key string, value interface{}, TTL ...time.Duration) error

	//UTIL
	Close() error
	GetKeysForRegex(regex string) ([]string, error)
	ExireAfterTTL(key string, newTTL time.Duration) error
	ExpireAt(key string, dateOfExpiry time.Time) error
	Ping() error
}
