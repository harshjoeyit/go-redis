package core

import "syscall"

type RedisCmd struct {
	Cmd  string
	Args []string
}

type Cmd struct {
	FD int
}

func (c Cmd) Read(data []byte) (n int, err error) {
	return syscall.Read(int(c.FD), data)
}

func (c Cmd) Write(data []byte) (n int, err error) {
	return syscall.Write(c.FD, data)
}
