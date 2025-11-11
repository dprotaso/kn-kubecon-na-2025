#!/usr/bin/env bash

docker run --rm \
  -ti \
  --network kind \
  -e DEBUG=true \
  -e HOST=0.0.0.0 \
  -p 6274:6274 \
  -p 6277:6277 \
  -e MCP_AUTO_OPEN_ENABLED=false \
  -e DANGEROUSLY_OMIT_AUTH=true \
  -e ALLOWED_ORIGINS=http://0.0.0.0:6274,http://localhost:6274,http://localhost:8080 \
  ghcr.io/modelcontextprotocol/inspector:latest
