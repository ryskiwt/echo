## implemented protocols
- WebSocket over TLS
- QUIC (with TLS)
- TLS over TCP
- TCP over IPv4 / IPv6
- UDP over IPv4 / IPv6

## build

```sh
$ cd xxx # xxx=wss, quic, tls, tcp, udp
$ go build
```

## usage

```
-c <server address>
  	client mode
-s <server address>
  	server mode
-i <interval>
  	sending interval [ms] *client mode only (default 1000)
-m <message>
  	message to send (default "Hello XXX !")
-v <version>
    IP version *TCP/UDP only (default "IPv4")
-n <stream number>
    number of parallel streams *QUIC only (default 3)
```

## WebSocket over TLS
### server

```sh
$ ./wss -s 127.0.0.1:18433                                                              +[master]
[echo-wss]2018/03/02 Listening on 127.0.0.1:18433 ...
[echo-wss]2018/03/02 Remote peer 127.0.0.1:50401 connected
[echo-wss]2018/03/02 Server: Got 'Hello WebSocket !'
[echo-wss]2018/03/02 Server: Got 'Hello WebSocket !'
[echo-wss]2018/03/02 Server: Got 'Hello WebSocket !'
[echo-wss]2018/03/02 Remote peer 127.0.0.1:50401 disconnected
```

### client

```sh
$ ./wss -c 127.0.0.1:18433                                                              +[master]
[echo-wss]2018/03/02 Client: Sending 'Hello WebSocket !'
[echo-wss]2018/03/02 Client: Got 'Hello WebSocket !'
[echo-wss]2018/03/02 Client: Sending 'Hello WebSocket !'
[echo-wss]2018/03/02 Client: Got 'Hello WebSocket !'
[echo-wss]2018/03/02 Client: Sending 'Hello WebSocket !'
[echo-wss]2018/03/02 Client: Got 'Hello WebSocket !'
```

## QUIC
### server

```sh
$ ./quic -s 127.0.0.1:4242                                                             +[master]
[echo-quic]2018/03/02 Listening on 127.0.0.1:4242 ...
[echo-quic]2018/03/02 Remote peer 127.0.0.1:57466 connected
[echo-quic]2018/03/02 Stream 3 on remote peer 127.0.0.1:57466 opened
[echo-quic]2018/03/02 Stream 5 on remote peer 127.0.0.1:57466 opened
[echo-quic]2018/03/02 Server: Got 'Hello QUIC !' on stream 5
[echo-quic]2018/03/02 Server: Got 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Server: Got 'Hello QUIC !' on stream 5
[echo-quic]2018/03/02 Server: Got 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Server: Got 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Server: Got 'Hello QUIC !' on stream 5
[echo-quic]2018/03/02 Stream 3 on remote peer 127.0.0.1:57466 closed
[echo-quic]2018/03/02 Stream 5 on remote peer 127.0.0.1:57466 closed
[echo-quic]2018/03/02 Remote peer 127.0.0.1:57466 disconnected
```

### client

```sh
$ ./quic -c 127.0.0.1:4242 -n 2                                                        +[master]
[echo-quic]2018/03/02 Client: Sending 'Hello QUIC !' on stream 5
[echo-quic]2018/03/02 Client: Sending 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Client: Got 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Client: Got 'Hello QUIC !' on stream 5
[echo-quic]2018/03/02 Client: Sending 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Client: Sending 'Hello QUIC !' on stream 5
[echo-quic]2018/03/02 Client: Got 'Hello QUIC !' on stream 5
[echo-quic]2018/03/02 Client: Got 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Client: Sending 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Client: Sending 'Hello QUIC !' on stream 5
[echo-quic]2018/03/02 Client: Got 'Hello QUIC !' on stream 3
[echo-quic]2018/03/02 Client: Got 'Hello QUIC !' on stream 5
```

## TLS
### server

```sh
$ ./tls -s :18433 -v IPv4                                                               +[master]
[echo-tls]2018/03/02 Listening on :18433 ...
[echo-tls]2018/03/02 Remote peer 127.0.0.1:50449 connected
[echo-tls]2018/03/02 Server: Got 'Hello TLS !'
[echo-tls]2018/03/02 Server: Got 'Hello TLS !'
[echo-tls]2018/03/02 Server: Got 'Hello TLS !'
[echo-tls]2018/03/02 Remote peer 127.0.0.1:50449 disconnected
```

### client

```sh
$ ./tls -c :18433 -v IPv4                                                               +[master]
[echo-tls]2018/03/02 Client: Sending 'Hello TLS !'
[echo-tls]2018/03/02 Client: Got 'Hello TLS !'
[echo-tls]2018/03/02 Client: Sending 'Hello TLS !'
[echo-tls]2018/03/02 Client: Got 'Hello TLS !'
[echo-tls]2018/03/02 Client: Sending 'Hello TLS !'
[echo-tls]2018/03/02 Client: Got 'Hello TLS !'
```

## TCP
### server

```sh
$ ./tcp -s [::]:8080 -v IPv6                                                            +[master]
[echo-tcp]2018/03/02 Listening on [::]:8080 ...
[echo-tcp]2018/03/02 Remote peer [::1]:50460 connected
[echo-tcp]2018/03/02 Server: Got 'Hello TCP !'
[echo-tcp]2018/03/02 Server: Got 'Hello TCP !'
[echo-tcp]2018/03/02 Server: Got 'Hello TCP !'
[echo-tcp]2018/03/02 Remote peer [::1]:50460 disconnected
```

### client

```sh
$ ./tcp -c [::]:8080 -v IPv6                                                            +[master]
[echo-tcp]2018/03/02 Client: Sending 'Hello TCP !'
[echo-tcp]2018/03/02 Client: Got 'Hello TCP !'
[echo-tcp]2018/03/02 Client: Sending 'Hello TCP !'
[echo-tcp]2018/03/02 Client: Got 'Hello TCP !'
[echo-tcp]2018/03/02 Client: Sending 'Hello TCP !'
[echo-tcp]2018/03/02 Client: Got 'Hello TCP !'
```

## UDP
### server

```sh
$ ./udp -s :8080 -v IPv6                                                                +[master]
[echo-udp]2018/03/02 Listening on :8080 ...
[echo-udp]2018/03/02 Server: Got 'Hello UDP !' from remote peer [::1]:63635
[echo-udp]2018/03/02 Server: Got 'Hello UDP !' from remote peer [::1]:63635
[echo-udp]2018/03/02 Server: Got 'Hello UDP !' from remote peer [::1]:63635
```

### client

```sh
$ ./udp -c :8080 -v IPv6                                                                +[master]
[echo-udp]2018/03/02 Client: Sending 'Hello UDP !'
[echo-udp]2018/03/02 Client: Got 'Hello UDP !'
[echo-udp]2018/03/02 Client: Sending 'Hello UDP !'
[echo-udp]2018/03/02 Client: Got 'Hello UDP !'
[echo-udp]2018/03/02 Client: Sending 'Hello UDP !'
[echo-udp]2018/03/02 Client: Got 'Hello UDP !'
```

## memo
### cert creation

```sh
$ openssl genrsa 2048 > server.key
$ openssl req -new -key server.key > server.csr
$ openssl x509 -days 3650 -req -signkey server.key < server.csr > server.crt
```
