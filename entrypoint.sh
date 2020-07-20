groupadd -g "$SETGID" groupname
useradd -d "$PWD" -u "$SETUID" -g "$SETGID" -G sudo username 
echo '===> Installing cert store *inside the container*'
sudo -EH -u username bash do_mkcert.sh 
exec sudo -EH -u username "$@"
