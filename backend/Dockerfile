# syntax=docker/dockerfile:1
FROM alpine:latest

ARG BUILDTAGS=sqlite_math_functions,sqlite_foreign_keys

RUN --mount=type=bind,source=./,target=/backend <<EOF
set -e
cd /backend
apk update
xargs -t -a .apk-requirements apk add
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
ENV GIN_MODE=release
EXPOSE 3824

CMD ["./entrypoint.sh"]
