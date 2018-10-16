#!/bin/bash
set -e

ROOT_KEY="${ROOT_KEY:-projects/joe-vault-kms-dev/locations/us-central1/keyRings/joe-hsm-testing/cryptoKeys/ca-signer/cryptoKeyVersions/1}"

export VAULT_CONFIG_PATH=/dev/null # disable token helper
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=root

# start vault server in the background
vault \
    server \
    -dev \
    -dev-root-token-id="root" \
    -dev-plugin-dir="$PWD/bin" &
vault_pid="$!"
shutdown() {
    kill "$vault_pid"
}
trap shutdown EXIT

sleep 0.2

# enable pki plugin at /pki
vault secrets enable -plugin-name=vault-gcp-cloud-kms-pki -path=pki plugin

wait