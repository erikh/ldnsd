#!/bin/sh

export CAROOT=/etc/ldnsd PATH=/usr/local/bin:${PATH}
if [ ! -f "${CAROOT}/rootCA.pem" ]
then
  mkcert -install
fi

if [ ! -f "${CAROOT}/server.pem" ]
then
  mkcert -cert-file /etc/ldnsd/server.pem -key-file /etc/ldnsd/server.key localhost 127.0.0.1
fi

if [ ! -f "${CAROOT}/client.pem" ]
then
  mkcert -client -cert-file /etc/ldnsd/client.pem -key-file /etc/ldnsd/client.key localhost 127.0.0.1
fi

exec "$@"
