package core

import (
	"errors"
	"slices"
	"testing"
	"time"
)

type TestReaderWriter struct {
	store []byte
}

func (trw *TestReaderWriter) Read(data []byte) (n int, err error) {
	copy(data, trw.store)
	return len(trw.store), nil
}

func (trw *TestReaderWriter) Write(data []byte) (n int, err error) {
	trw.store = make([]byte, len(data))
	copy(trw.store, data)
	return len(data), nil
}

func TestEvalAndRespondPING(t *testing.T) {
	tests := map[string]struct {
		rcmd     *RedisCmd
		response []byte
		err      error
	}{
		"Simple PING": {
			rcmd: &RedisCmd{
				Cmd:  "PING",
				Args: []string{},
			},
			response: []byte("+PONG\r\n"),
			err:      nil,
		},
		"PING with one arg": {
			rcmd: &RedisCmd{
				Cmd:  "PING",
				Args: []string{"hello"},
			},
			response: []byte("$5\r\nhello\r\n"),
			err:      nil,
		},
		"PING with two args": {
			rcmd: &RedisCmd{
				Cmd:  "PING",
				Args: []string{"hello", "world"},
			},
			response: nil,
			err:      errors.New("ERR wrong number of arguments for 'ping' command"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := NewServer()
			trw := &TestReaderWriter{}
			err := s.EvalAndRespond(tt.rcmd, trw)

			// error assertions
			if tt.err != nil {
				if err == nil {
					t.Fatalf("expected error: %v, got nil", tt.err)
				}
				if err.Error() != tt.err.Error() {
					t.Fatalf("expected error: %v, got: %v", tt.err, err)
				}
				return
			}

			// success assertions
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(trw.store, tt.response) {
				t.Errorf("response mismatch/expected: %s, got: %s", string(tt.response), string(trw.store))
			}
		})
	}
}

func TestEvalAndRespondSET(t *testing.T) {
	tests := map[string]struct {
		rcmd     *RedisCmd
		response []byte
		key      string
		setValue string
		ttlSec   int64
		err      error
	}{
		"SET without expiry": {
			rcmd: &RedisCmd{
				Cmd:  "SET",
				Args: []string{"k1", "v1"},
			},
			response: []byte("+OK\r\n"),
			key:      "k1",
			setValue: "v1",
			ttlSec:   -1,
		},
		"SET with TTL": {
			rcmd: &RedisCmd{
				Cmd:  "SET",
				Args: []string{"k2", "v2", "EX", "10"},
			},
			response: []byte("+OK\r\n"),
			key:      "k2",
			setValue: "v2",
			ttlSec:   10,
		},
		"SET unsupported arg": {
			rcmd: &RedisCmd{
				Cmd:  "SET",
				Args: []string{"k3", "v3", "XX"},
			},
			err: errors.New("ERR syntax error"),
		},
		"SET value for expiry not provided": {
			rcmd: &RedisCmd{
				Cmd:  "SET",
				Args: []string{"k3", "v3", "EX"},
			},
			err: errors.New("ERR syntax error"),
		},
		"SET value for expiry is not an integer": {
			rcmd: &RedisCmd{
				Cmd:  "SET",
				Args: []string{"k3", "v3", "EX", "X"},
			},
			err: errors.New("ERR value is not an integer or out of range"),
		},
		"SET without key": {
			rcmd: &RedisCmd{
				Cmd:  "SET",
				Args: []string{},
			},
			err: errors.New("ERR wrong number of arguments for 'set' command"),
		},
		"SET without value": {
			rcmd: &RedisCmd{
				Cmd:  "SET",
				Args: []string{"k4"},
			},
			err: errors.New("ERR wrong number of arguments for 'set' command"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := NewServer()
			trw := &TestReaderWriter{}

			err := s.EvalAndRespond(tt.rcmd, trw)

			// error assertions
			if tt.err != nil {
				if err == nil {
					t.Fatalf("expected error: %v, got nil", tt.err)
				}
				if err.Error() != tt.err.Error() {
					t.Fatalf("expected error: %v, got: %v", tt.err, err)
				}
				return
			}

			// success assertions
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(trw.store, tt.response) {
				t.Fatalf(
					"response mismatch\nexpected: %q\ngot: %q",
					tt.response,
					trw.store,
				)
			}

			ss := s.DataStoreSnapshot().(InMemDataStore)

			obj, ok := ss[tt.key]
			if !ok {
				t.Fatalf("key %q not found in datastore", tt.key)
			}

			gotValue := obj.value.(string)

			if gotValue != tt.setValue {
				t.Fatalf(
					"value mismatch\nexpected: %s\ngot: %s",
					tt.setValue,
					gotValue,
				)
			}

			// TTL assertions
			if tt.ttlSec == -1 {
				if obj.expiresAt != -1 {
					t.Fatalf("expected no expiry, got: %d", obj.expiresAt)
				}
				return
			}

			gotTTL := (obj.expiresAt - time.Now().UnixMilli()) / 1000

			// avoid exact equality for time-based tests
			if gotTTL < tt.ttlSec-1 || gotTTL > tt.ttlSec {
				t.Fatalf(
					"ttl mismatch\nexpected around: %d\ngot: %d",
					tt.ttlSec,
					gotTTL,
				)
			}
		})
	}
}
