#!/bin/sh
set -e
export VAULT_CONFIG_PATH=./nofile
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=root

vault \
    server \
    -dev \
    -dev-root-token-id="root" \
    -dev-plugin-dir=$PWD/bin &

vault secrets enable -plugin-name=vault-cloud-kms-pki -path=pki plugin
vault write pki/root/generate/internal common_name=rootCA
vault write pki/roles/any  allowed_domains=example.com allow_subdomains=true max_ttl=24h
vault write pki/issue/any common_name=foo.example.com ttl=1h
wait

# TODO exit trap to stop vault
