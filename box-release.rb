from "debian:latest"

MKCERT_VERSION = "1.4.1"
MKCERT_URL = "https://github.com/FiloSottile/mkcert/releases/download/v#{MKCERT_VERSION}/mkcert-v#{MKCERT_VERSION}-linux-amd64"

def go_get(name)
  run "go get -v -u #{name}"
end

def download(name, url)
  run "curl -sSL -o /#{name} '#{url}'"
  yield "/#{name}"
  run "rm -f /#{name}"
end

version = getenv("VERSION")
skip do
  run "apt update && apt install curl -y"
  download("mkcert", MKCERT_URL) do |path|
    run "chmod 0755 '#{path}'"
    run "mv '#{path}' /tmp/mkcert"
  end

  copy "ldnsd-#{version}.tar.gz", "/tmp/"
  inside "/tmp" do
    run "tar vxzf ldnsd-#{version}.tar.gz"
  end
end

inside "/tmp/ldnsd-#{version}" do
  run "mv -v ldnsd ldnsctl /usr/local/bin && mkdir -p /etc/ldnsd && mv -v example.conf /etc/ldnsd/ldnsd.conf"
end

run "mv /tmp/mkcert /usr/local/bin"
copy "release-entrypoint.sh", "/entrypoint.sh"
copy "VERSION", "/VERSION"
run "chmod 755 /entrypoint.sh"
set_exec entrypoint: ["/entrypoint.sh"], cmd: ["/usr/local/bin/ldnsd", "/etc/ldnsd/ldnsd.conf"]
