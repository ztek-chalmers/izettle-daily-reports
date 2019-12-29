#!/bin/bash
cd "$(dirname "$0")"
if [ ! -f server.key ]; then
openssl req \
       -x509 \
       -nodes \
       -newkey rsa:2048 \
       -keyout server.key \
       -out server.crt \
       -days 7300 \
       -subj "/C=SE/ST=Vastra Gotaland/L=Gothemburg/O=Chalmers Students/OU=Ztyret/CN=*"
fi
go run cmd/sync-report/main.go