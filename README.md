# go-tun2socks

A tun2socks implementation written in Go.

Tested and worked on macOS, Linux and iOS (as a library).

## Overview

```
                                      lwip.Setup()
                                           +
                                           |
                                           |
                                           |
                                           |                TCP/UDP             lwip.RegisterTCPConnectionHandler()
                                           |
                          lwip.Input()     |           tun2socks.Connection     lwip.RegisterUDPConnectionHandler()
                                           v
Application +------> TUN +-----------> lwIP stack +------------------------------> tun2socks.ConnectionHandler +-------> SOCKS5 server +--> Destination


                         <-----------+
                    lwip.RegisterOutputFn()

```

## Features

- Support both TCP and UDP (only IPv4 for now)
- SOCKS5 proxing (both TCP and UDP)
- UDP direct relaying (no proxy)
- ICMP local echoing

## Build

```sh
go get github.com/eycorsican/go-tun2socks
cd $GOPATH/src/github.com/eycorsican/go-tun2socks
go get -d ./...
make clean && make build
./build/tun2socks -h
```

## Run

Note that the name `tun1` maybe not available, make sure use `ifconfig` or `ip addr` to check it out.
```sh
./build/tun2socks -tunName tun1 -tunAddr 240.0.0.2 -tunGw 240.0.0.1 -proxyType socks -proxyServer 1.2.3.4:1086
```

## Configure Routing Table

Suppose your original gateway is 192.168.0.1. The proxy server address is 1.2.3.4.

The following commands will need root permissions.

### macOS

The program will automatically create a tun device for you on macOS. To show the created tun device, use ifconfig.

Delete original gateway:

```sh
route delete default
```

Add our tun interface as the default gateway:

```sh
route add default 240.0.0.2
```

Add a route for your proxy server to bypass the tun interface:

```sh
route add 1.2.3.4/32 192.168.0.1
```

### Linux

The program will not create the tun device for you on Linux. You need to create the tun device by yourself:

```sh
ip tuntap add mode tun dev tun1
ip addr add 240.0.0.2 dev tun1
ip link set dev tun1 up
```

Delete original gateway:

```sh
ip route del default
```

Add our tun interface as the default gateway:

```sh
ip route add default via 240.0.0.2
```

Add a route for your proxy server to bypass the tun interface:

```sh
ip route add 1.2.3.4/32 via 192.168.0.1
```

## What happened to lwIP?

Take a look at this repo: https://github.com/eycorsican/lwip

## Acknowledgements
- https://github.com/zhuhaow/tun2socks
- https://github.com/yinghuocho/gotun2socks
- https://github.com/shadowsocks/go-shadowsocks2
- https://github.com/nadoo/glider
