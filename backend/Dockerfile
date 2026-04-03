# syntax=docker/dockerfile:1
FROM debian:trixie-slim

ARG BUILDTAGS=sqlite_math_functions,sqlite_foreign_keys

RUN --mount=type=bind,source=./,target=/backend,rw <<EOF
set -e
# install basic packages
export DEBIAN_FRONTEND=noninteractive
cd /backend
apt update && apt upgrade -y
xargs -t -a .deb-requirements apt install -y
# install Go
cd /root
ARCH=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/)
wget https://go.dev/dl/go1.26.0.linux-$ARCH.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.26.0.linux-$ARCH.tar.gz
export PATH="$PATH:/usr/local/go/bin"
cd /backend
# install uv
curl -LsSf https://astral.sh/uv/install.sh | sh
# install server
make docs
go mod tidy -diff
go install -tags=$BUILDTAGS ./cmd/server
go clean -cache
go clean -modcache
EOF

RUN mkdir -p /backend

COPY <<EOT /backend/entrypoint.sh
#!/bin/sh
export PATH="$(go env GOPATH)/bin:$PATH"
server
EOT
RUN chmod +x /backend/entrypoint.sh

WORKDIR /backend
ENV PATH="/usr/local/go/bin:/root/.local/bin:${PATH}"
ENV GIN_MODE=release
EXPOSE 3824

CMD ["./entrypoint.sh"]
