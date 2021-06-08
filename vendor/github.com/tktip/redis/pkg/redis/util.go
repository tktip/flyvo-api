package redis

import "encoding"

func supportedReadDataType(target interface{}) bool {
	switch target.(type) {
	case nil:
	case *string:
	case *[]byte:
	case *int:
	case *int8:
	case *int16:
	case *int32:
	case *int64:
	case *uint:
	case *uint8:
	case *uint16:
	case *uint32:
	case *uint64:
	case *float32:
	case *float64:
	case *bool:
	case encoding.BinaryUnmarshaler:
	default:
		return false
	}
	return true
}

func supportedWriteDataType(target interface{}) bool {
	switch target.(type) {
	case nil:
	case string:
	case []byte:
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
	case float32:
	case float64:
	case bool:
	case encoding.BinaryMarshaler:
	default:
		return false
	}
	return true
}
