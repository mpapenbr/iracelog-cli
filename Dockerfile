# This file is used by goreleaser
FROM scratch
ENTRYPOINT ["/iracelog-cli"]
COPY iracelog-cli /
