# ldnsd: Light DNSd: a small, 0 ttl A record store that is remotely programmable.

Light DNSd is largely designed for testing & small environments, providing an
easy to manage DNS service that serves the minimum necessary to deliver name
services centrally. Compare with a remotely programmable `/etc/hosts` file that
lives in a central location. ldnsd comes with `ldnsctl` which can handle the
programming, or use/generate GRPC-compliant clients from our protobuf
definitions to program it inside of your tools directly.

Light DNSd is backed by sqlite3 and provides very few features:

- No recursion
- No caching (although this may be added soon, see _Potential Issues_)
- No forwarding
- All records are 0 TTL

Since not all clients are very happy with how ldnsd sees the world (simply), it
is _strongly advised_ that you front it with a caching, recursive,
standards-compliant nameserver like coredns, bind, dqcache/dnscache, etc,
especially for the purposes of servicing client resolvers. Think of ldnsd more
as a companion to your DNS stack instead of replacing it.

## Installing

Installing a release is your best choice. Otherwise, you can still `go get github.com/erikh/ldnsd/...`
and get the desired result in your `$GOBIN` or `$GOPATH/bin`.

### Docker Image

If you wish to use Docker to power ldnsd, you can use our `erikh/ldnsd`
version-tagged images. Running with `--net=host` is advisable to avoid the UDP
proxy docker provides as it tends to drop packets under high load.

Example usage:

```bash
# start the service
$ docker run -it -d --name ldnsd --net=host erikh/ldnsd:0.1.0
# configure some hosts
$ docker exec -it ldnsd ldnsctl set myhost 1.2.3.4
$ dig myhost.internal. @127.0.0.1
```

### Manual Installation

If you'd like to build in a container or build the release version, just make
sure to have `docker` installed; [box](https://github.com/box-builder/box) will
be installed as root as a part of the process during the first run while
creating the shell image.

To create a release tarball:

```shell
$ make shell
# <inside the container>
$ make release # out comes the tarball in the CWD
```

### Make a CA

You'll need a CA. I strongly recommend installing
[mkcert](https://github.com/FiloSottile/mkcert) and trying this script to
generate the CA, one server cert, and one client cert in `/etc/ldnsd` (you may
need to be root to write to this directory):

This will only make the service available on `localhost/127.0.0.1` through this
cert. All other attempts will be rejected.

Note if you change the directories, you will need to adjust the configuration file, which is discussed below.

```shell
export CAROOT="/etc/ldnsd"
mkcert -install
mkcert -ecdsa -cert-file /etc/ldnsd/server.pem -key-file /etc/ldnsd/server.key localhost 127.0.0.1
mkcert -ecdsa -client -cert-file /etc/ldnsd/client.pem -key-file /etc/ldnsd/client.key localhost 127.0.0.1
```

### Configuration

`ldnsd` takes one argument, the configuration filename. It is a basic YAML
document that covers certificate management and network listening information.

Here is an example. If in doubt, all options have defaults:

```yaml
# vim: ft=yaml
---
certificate:
  ca: "/etc/ldnsd/rootCA.pem"
  cert: "/etc/ldnsd/server.pem"
  key: "/etc/ldnsd/server.key"
# grpc listening port
grpc: "localhost:7847"
# dns listening port (udp only!)
listen: "localhost:53"
# TLD for domains.
domain: "internal"
```

## Launching and Utilization

`ldnsd my.conf` to launch the service, it does not daemonize so be sure to run
it in the background if you need to. Also, since :53 is privileged port, you
will need to run this process as root.

`ldnsctl` can be used to query and manipulate the service. To resolve hosts, use DNS:

```shell
dig bar.internal. @127.0.0.1
```

"Set", "List" and "Delete" operations go through `ldnsctl`. To review how to
use those operations, please review the `ldnsctl help` command's output.

## Potential Issues

sqlite3 (and the way we use it) under a lot of contention could cause slow
responses, which could lead to dropped queries. There are benchmarks at the
root of the repository, if you are able to produce this behavior with them
please let us know.

On a 12 thread / 6 core intel 9xxx processor, the erikh/dnsserver package
delivers 7000ns/op for a similar test that ldnsd delivers in 30000ns/op,
suggesting that (understandably) sqlite3 is slower than map access. That said,
extended "burn-in" benchmarks have shown no delivery issues so far.

For most other bugs, please see the Issues pages.

## Author

Erik Hollensbe <erik+git@hollensbe.org>
