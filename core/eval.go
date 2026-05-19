package core

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Server struct {
	MaxConnClients int
	connClients    int
	dataStore      DataStore
}

func NewServer() *Server {
	return &Server{
		MaxConnClients: 20000,
		dataStore:      NewInMemDataStore(),
	}
}

func (s *Server) IncrConnClients() {
	s.connClients++
}

func (s *Server) DecrConnClients() {
	s.connClients++
}

func (s *Server) ConnClients() int {
	return s.connClients
}

func (s *Server) DataStoreSnapshot() DataStore {
	return s.dataStore.(InMemDataStore).Snapshot()
}

func (s *Server) evalPING(args []string, c io.ReadWriter) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := c.Write(b)
	return err
}

func (s *Server) evalSET(args []string, c io.ReadWriter) error {
	// loop in pair (arg name, arg value)
	if len(args) <= 1 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}

	var key, value string
	var expDurationMs int64 = -1

	key, value = args[0], args[1]

	for i := 2; i < len(args); i = i + 2 {
		if i+1 == len(args) {
			return errors.New("ERR syntax error")
		}
		switch args[i] {
		case "EX", "ex":
			expDurationSec, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return errors.New("ERR value is not an integer or out of range")
			}
			expDurationMs = expDurationSec * 1000
		default:
			return fmt.Errorf("ERR syntax error")
		}
	}

	s.dataStore.Set(key, value, expDurationMs)

	c.Write([]byte("+OK\r\n"))

	return nil
}

func (s *Server) evalGET(args []string, c io.ReadWriter) error {
	_ = args
	_ = c
	return nil
}

func (s *Server) evalTTL(args []string, c io.ReadWriter) error {
	_ = args
	_ = c
	return nil
}

func (s *Server) EvalAndRespond(cmd *RedisCmd, c io.ReadWriter) error {
	switch cmd.Cmd {
	case "PING":
		return s.evalPING(cmd.Args, c)
	case "SET":
		return s.evalSET(cmd.Args, c)
	case "GET":
		return s.evalGET(cmd.Args, c)
	case "TTL":
		return s.evalTTL(cmd.Args, c)
	default:
		return fmt.Errorf("unsupported command %s", cmd.Cmd)
	}
}
