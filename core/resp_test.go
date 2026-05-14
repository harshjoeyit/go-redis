package core

import (
	"fmt"
	"testing"
)

type tc struct {
	encoded string
	decoded any
}

func TestDecodeSimpleString(t *testing.T) {
	tests := []tc{
		{
			encoded: "+OK\r\n",
			decoded: "OK",
		},
		{
			encoded: "+PING\r\n",
			decoded: "PING",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			got, err := DecodeOne([]byte(tt.encoded))
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			gotVal, ok := got.(string)
			if !ok {
				t.Errorf("Expected string, got: %v", got)
			} else if gotVal != tt.decoded {
				t.Errorf("Got: %v Expected:%v", gotVal, tt.decoded)
			}
		})
	}
}

func TestDecodeSimpleError(t *testing.T) {
	tests := []tc{
		{
			encoded: "-Error message\r\n",
			decoded: "Error message",
		},
		{
			encoded: "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
			decoded: "WRONGTYPE Operation against a key holding the wrong kind of value",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			got, err := DecodeOne([]byte(tt.encoded))
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			gotVal, ok := got.(string)
			if !ok {
				t.Errorf("Expected string, got: %v", got)
			} else if gotVal != tt.decoded {
				t.Errorf("Got: %v Expected:%v", gotVal, tt.decoded)
			}
		})
	}
}

func TestDecodeInteger(t *testing.T) {
	tests := []tc{
		{
			encoded: ":0\r\n",
			decoded: int64(0),
		},
		{
			encoded: ":1000\r\n",
			decoded: int64(1000),
		},
		{
			encoded: ":95213\r\n",
			decoded: int64(95213),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			got, err := DecodeOne([]byte(tt.encoded))
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			gotVal, ok := got.(int64)
			if !ok {
				t.Errorf("Expected int64, got: %v", got)
			} else if gotVal != tt.decoded {
				t.Errorf("Got: %v Expected:%v", gotVal, tt.decoded)
			}
		})
	}
}

func TestDecodeBulkString(t *testing.T) {
	tests := []tc{
		{
			encoded: "$5\r\nhello\r\n",
			decoded: "hello",
		},
		{
			encoded: "$0\r\n\r\n",
			decoded: "",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			got, err := DecodeOne([]byte(tt.encoded))
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			gotVal, ok := got.(string)
			if !ok {
				t.Errorf("Expected string, got: %v", got)
			} else if gotVal != tt.decoded {
				t.Errorf("Got: %v Expected:%v", gotVal, tt.decoded)
			}
		})
	}
}

func TestDecodeArray(t *testing.T) {
	tests := []tc{
		{
			encoded: "*0\r\n",
			decoded: []any{},
		},
		{
			encoded: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			decoded: []any{"hello", "world"},
		},
		{
			encoded: "*3\r\n:1\r\n:2\r\n:3\r\n",
			decoded: []any{1, 2, 3},
		},
		{
			encoded: "*5\r\n:1\r\n:2\r\n:3\r\n:4\r\n$5\r\nhello\r\n",
			decoded: []any{1, 2, 3, 4, "hello"},
		},
		{
			encoded: "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n-World\r\n",
			decoded: []any{[]any{1, 2, 3}, []any{"Hello", "World"}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			got, err := DecodeOne([]byte(tt.encoded))
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			gotVal, ok := got.([]any)
			if !ok {
				t.Errorf("Expected array, got: %v", got)
			} else {
				decVal := tt.decoded.([]any)
				if len(gotVal) != len(decVal) {
					t.Errorf("Expected len: %d, Got len: %d", len(decVal), len(gotVal))
				} else {
					for i := range gotVal {
						if fmt.Sprintf("%v", gotVal[i]) != fmt.Sprintf("%v", decVal[i]) {
							t.Errorf("Expected: %d, Got: %d", gotVal[i], decVal[i])
						}
					}
				}
			}
		})
	}
}
