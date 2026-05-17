# Implementation of Redis in Golang

Aim is to understand via implementation - 
- RESP protocal
- I/O multiplexing

## Concepts 

### I/O multiplexing
**Why do we need this?**

Can't we just spin up multiple thread (go-routines) per connection?
1. Multiple clients - can go upto 10k clients
    - Connection-pooled microservices architectures (each service pool = 10–50 connections)
    - Pub/Sub scenarios
2. Avoid multi-threading and thus avoid locking of in-mem data structures - reduces contention
3. Predictable latency - bottleneck is always I/O, no computation (thread scheduling, lock contention)

**Every event loop is simply this:**
```
loop:
  ready_fds = epoll_wait(all_registered_fds)  // blocks until something is ready
  for each fd in ready_fds:
    read and process command
```

**System calls:**
1. `epoll_create1` - Create a new EPoller
2. `epoll_ctl` - Register file descriptor (fd)
3. `epoll_wait` - Waits for updates on registered fds

### Benchmarks

```bash
./redis-benchmark -n 10000 -t ping_mbulk -c 200 -h localhost -p 7379
```