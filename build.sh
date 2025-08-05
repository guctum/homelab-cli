#!/bin/sh

go build -o homelab ./cmd/

sudo cp homelab /usr/local/bin/

# Clean up
rm homelab
