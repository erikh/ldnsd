#!/bin/sh

set -e

mkcert -install && \
  mkcert -ecdsa -cert-file /etc/ldnsd/server.pem -key-file /etc/ldnsd/server.key localhost 127.0.0.1 && \
  mkcert -ecdsa -client -cert-file /etc/ldnsd/client.pem -key-file /etc/ldnsd/client.key localhost 127.0.0.1
