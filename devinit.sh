#!/bin/bash
echo "devinit.sh called"


go install github.com/goreleaser/goreleaser/v2@latest
# go install github.com/caarlos0/svu@latest

if [ -f setuplinks.sh ]; then
    . ./setuplinks.sh
fi
