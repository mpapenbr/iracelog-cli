# This file is used by goreleaser
FROM alpine:3.20
ENTRYPOINT ["/iracelog-cli"]
COPY iracelog-cli /
COPY samples /
