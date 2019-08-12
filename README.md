# max-udp
Robust protocol that works on the connection with packet losses.

## Server
Server has a port flag.
```
$ go build server.go
$ ./server -port 33333
```

## Client
Client has a host, port, and file flag.
```
$ go build client.go
$ ./client -file dummy_small -port 33333
```
