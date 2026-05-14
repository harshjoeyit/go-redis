package core

import (
	"errors"
	"fmt"
)

// Simple strings
func decodeSimpleString(data []byte) (any, int, error) {
	pos := 1
	for ; data[pos] != '\r' && pos < len(data); pos++ {
	}

	return string(data[1:pos]), pos + 2, nil
}

func decodeSimpleError(data []byte) (any, int, error) {
	return decodeSimpleString(data)
}

func decodeInt64(data []byte) (any, int, error) {
	pos := 1
	var val int64
	for ; data[pos] != '\r' && pos < len(data); pos++ {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return val, pos + 2, errors.New("not a number")
		}
		val = val*10 + int64(b-'0')
	}
	return val, pos + 2, nil
}

func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}
	return 0, 0
}

func decodeBulkString(data []byte) (any, int, error) {
	pos := 1

	// reading the length and forwarding the pos by
	// the length of the integer
	len, delta := readLength(data[pos:])
	pos += delta // pos is where string begins

	return string(data[pos:(pos + len)]), pos + len + 2, nil
}

func decodeArray(data []byte) (any, int, error) {
	pos := 1

	// reading the length and forwarding the pos by
	// the length of the integer
	len, delta := readLength(data[pos:])
	pos += delta

	elems := make([]any, len)
	for i := range len {
		elem, delta, err := decodeOne(data[pos:])
		if err != nil {
			return elems, pos, err
		}
		elems[i] = elem
		pos += delta
	}
	return elems, pos, nil
}

// decodeOne return decoded data, index in byte array 'data'
// from which further decoding can start if needed
func decodeOne(data []byte) (any, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}

	switch data[0] {
	case '+':
		return decodeSimpleString(data)
	case '-':
		return decodeSimpleError(data)
	case ':':
		return decodeInt64(data)
	case '$':
		return decodeBulkString(data)
	case '*':
		return decodeArray(data)
	default:
		return nil, 0, fmt.Errorf("unrecognized data type: %c", data[0])
	}
}

func DecodeOne(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}
	value, _, err := decodeOne(data)
	return value, err
}
