package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/harshjoeyit/goredis/core"
)

func RunSyncTCPServer() {
	server := core.NewServer()

	addr, err := net.ResolveTCPAddr("tcp", ":7379")
	if err != nil {
		log.Printf("Error resolving TCP address: %v\n", err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("Error starting server: %v\n", err)
	}
	defer listener.Close()

	log.Println("TCP server listening at", addr)

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		server.IncrConnClients()

		handleConnection(server, c)
	}
}

func handleConnection(s *core.Server, c net.Conn) {
	defer c.Close()
	clientAddr := c.RemoteAddr().String()

	fmt.Printf("client connected: %s, concurrent clients: %d\n", clientAddr, s.ConnClients())

	for {
		// decode command
		cmd, err := readCommand(c)
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.DecrConnClients()
				log.Printf("connection closed by client: %s, concurrent clients: %d\n", clientAddr, s.ConnClients())
			} else {
				log.Printf("unexpected error: %v", err)
			}
			break
		}
		respond(s, cmd, c)
	}
}

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf)
	if err != nil {
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  tokens[0],
		Args: tokens[1:],
	}, nil
}

func respond(s *core.Server, cmd *core.RedisCmd, c io.ReadWriter) {
	err := s.EvalAndRespond(cmd, c)
	if err != nil {
		respondError(err, c)
	}
}

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}
