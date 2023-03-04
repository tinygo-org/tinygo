This is a port of Go's "net" package.  The port offers a subset of Go's "net"
package.  The subset maintains Go 1 compatiblity guarantee.

The "net" package is modified to use netdev, TinyGo's network device driver interface.
Netdev replaces the OS syscall interface for I/O access to the networking
device.

#### Table of Contents

- ["net" Package](#net-package)
- [Netdev and Netlink](#netdev-and-netlink)
- [Using "net" and "net/http" Packages](#using-net-and-nethttp-packages)
 
## "net" Package

The "net" package is ported from Go 1.19.3.  The tree listings below shows the
files copied.  If the file is marked with an '\*', it is copied _and_ modified
to work with netdev.  If the file is marked with an '+', the file is new.  If
there is no mark, it is a straight copy.

```
src/net
├── dial.go			*
├── http
│   ├── client.go		*
│   ├── clone.go
│   ├── cookie.go
│   ├── fs.go
│   ├── header.go		*
│   ├── http.go
│   ├── internal
│   │   ├── ascii
│   │   │   ├── print.go
│   │   │   └── print_test.go
│   │   ├── chunked.go
│   │   └── chunked_test.go
│   ├── jar.go
│   ├── method.go
│   ├── request.go		*
│   ├── response.go		*
│   ├── server.go		*
│   ├── sniff.go
│   ├── status.go
│   ├── transfer.go		*
│   └── transport.go		*
├── ip.go
├── iprawsock.go		*
├── ipsock.go			*
├── mac.go
├── mac_test.go
├── netdev.go			+
├── net.go			*
├── parse.go
├── pipe.go
├── README.md
├── tcpsock.go			*
├── tlssock.go			+
└── udpsock.go			*

src/crypto/tls/
├── common.go			*
└── tls.go			*
```

The modifications to "net" are to basically wrap TCPConn, UDPConn, and TLSConn
around netdev socket calls.  In Go, these net.Conns call out to OS syscalls for
the socket operations.  In TinyGo, the OS syscalls aren't available, so netdev
socket calls are substituted.

The modifications to "net/http" are on the client and the server side.  On the
client side, the TinyGo code changes remove the back-end round-tripper code and
replaces it with direct calls to TCPConns/TLSConns.  All of Go's http
request/response handling code is intact and operational in TinyGo.  Same holds
true for the server side.  The server side supports the normal server features
like ServeMux and Hijacker (for websockets).

### Maintaining "net"

As Go progresses, changes to the "net" package need to be periodically
back-ported to TinyGo's "net" package.  This is to pick up any upstream bug
fixes or security fixes.

Changes "net" package files are marked with // TINYGO comments.

The files that are marked modified * may contain only a subset of the original
file.  Basically only the parts necessary to compile and run the example/net
examples are copied (and maybe modified).

## Netdev and Netlink

Netdev is TinyGo's network device driver model.  Network drivers implement the
netdever interface, providing a common network I/O interface to TinyGo's "net"
package.  The interface is modeled after the BSD socket interface.  net.Conn
implementations (TCPConn, UDPConn, and TLSConn) use the netdev interface for
device I/O access.

Network drivers also (optionally) implement the Netlinker interface.  This
interface is not used by TinyGo's "net" package, but rather provides the TinyGo
application direct access to the network device for common settings and control
that fall outside of netdev's socket interface.

See the README-net.md in drivers repo for more details on netdev and netlink.

## Using "net" and "net/http" Packages

See README-net.md in drivers repo to more details on using "net" and "net/http"
packages in a TinyGo application.
