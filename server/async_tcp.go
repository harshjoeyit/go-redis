package server

import (
	"errors"
	"io"
	"log"
	"syscall"

	"github.com/harshjoeyit/goredis/core"
)

func RunAsyncTCPServer() {
	s := core.NewServer()

	// create a kqueue instance - epoll_create1
	kq, err := syscall.Kqueue()
	if err != nil {
		panic(err)
	}

	// create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}

	// bind to attach socket to IP + port
	serverSockAddr := &syscall.SockaddrInet4{
		Port: 7379,
		Addr: [4]byte{0, 0, 0, 0},
	}
	err = syscall.Bind(serverFD, serverSockAddr)
	if err != nil {
		panic(err)
	}

	// listen for incoming conns
	err = syscall.Listen(serverFD, 10)

	log.Printf("async TCP server listening at :%+v\n", serverSockAddr)

	// register server with kqueue
	registerWithKQ(kq, serverFD, true)

	// event loop - epoll_wait
	for {
		events := make([]syscall.Kevent_t, s.MaxConnClients)

		log.Println("wating for events...")
		n, err := syscall.Kevent(kq, []syscall.Kevent_t{}, events, nil)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("events ready: ", n)

		// dispatch - iterrate over returned events
		for _, e := range events[:n] {
			if e.Ident == uint64(serverFD) {
				// accept new client connection
				clientFD, clientSockAddr, err := syscall.Accept(serverFD)
				if err != nil {
					log.Printf("client fd: %d, addr: %+v failed to connect\n", e.Ident, clientSockAddr)
				}

				s.IncrConnClients()
				log.Printf("new client connected fd: %d, addr: %+v, connected clients: %d\n", e.Ident, clientSockAddr, s.ConnClients())

				// register client with kqueue
				registerWithKQ(kq, clientFD, true)
			} else {
				// command on existing client connection
				cmd := core.Cmd{FD: int(e.Ident)}
				rcmd, err := readCommand(cmd)
				if err != nil {
					if errors.Is(err, io.EOF) || err.Error() == "no data" {
						// close connection and unregister
						syscall.Close(cmd.FD)
						registerWithKQ(kq, cmd.FD, false)

						s.DecrConnClients()
						log.Printf("connection closed by client: %+v, connected clients: %d\n", e.Ident, s.ConnClients())
					} else {
						log.Printf("unexpected error: %+v", err)
					}
					continue
				}

				// respond to client
				respond(s, rcmd, cmd)
			}
		}
	}
}

func registerWithKQ(kq, fd int, register bool) {
	ke := syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
	}
	if !register {
		ke.Flags = syscall.EV_DELETE
	}
	// canonical to epoll_ctl
	syscall.Kevent(kq, []syscall.Kevent_t{ke}, []syscall.Kevent_t{}, nil)
}
