groupadd -g "$SETGID" groupname
useradd -d "$PWD" -u "$SETUID" -g "$SETGID" -G sudo username 
echo '===> Installing cert store *inside the container*'
sudo -EH -u username mkcert -install && \
  sudo -EH -u username mkcert -ecdsa -cert-file /etc/ldnsd/server.pem -key-file /etc/ldnsd/server.key localhost 127.0.0.1 && \
  sudo -EH -u username mkcert -ecdsa -client -cert-file /etc/ldnsd/client.pem -key-file /etc/ldnsd/client.key localhost 127.0.0.1
exec sudo -EH -u username "$@"
