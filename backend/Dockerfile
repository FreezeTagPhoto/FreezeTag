# syntax=docker/dockerfile:1
FROM alpine:latest

COPY api /backend/api
COPY cmd /backend/cmd
COPY pkg /backend/pkg
COPY .apk-requirements /backend/.apk-requirements
COPY go.mod /backend/go.mod
COPY go.sum /backend/go.sum
COPY Makefile /backend/Makefile

WORKDIR /backend

ENV GIN_MODE=release
ENV PATH=".go/bin:$PATH"

RUN mkdir -p /backend/.apk-cache
RUN apk update --cache-dir /backend/.apk-cache
RUN xargs -t -a .apk-requirements apk add --cache-dir /backend/.apk-cache
RUN make tidy

EXPOSE 3824

CMD ["make", "run"]