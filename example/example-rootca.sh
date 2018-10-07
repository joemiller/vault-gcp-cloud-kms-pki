#!/bin/bash
set -e

ROOT_KEY="${ROOT_KEY:-projects/joe-vault-kms-dev/locations/us-central1/keyRings/joe-hsm-testing/cryptoKeys/ca-signer/cryptoKeyVersions/1}"

export VAULT_CONFIG_PATH=/dev/null # disable token helper
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=root

tmpdir="./tmp-rootca"
mkdir -p "$tmpdir"
rm -f "$tmpdir/*"

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

# generate self-signed cert for the root CA
root_crt="$tmpdir/rootCA.crt"
echo "==> Generating self-signed rootCA cert: $root_crt"
vault write \
    -field=certificate \
    pki/root/generate/internal \
    common_name=rootCA \
    google_cloud_kms_key="$ROOT_KEY" \
    > "$root_crt"

echo "==> Creating an end-entity key and cert signed by the intermediate CA"
vault write pki/roles/any allowed_domains=example.com allow_subdomains=true max_ttl=24h

vault write -format=json pki/issue/any common_name=foo.example.com ttl=1h >"$tmpdir/endpoint.bundle.json"

if command -v jq >/dev/null; then
    cat "$tmpdir/endpoint.bundle.json" | jq -r '.data.certificate' >"$tmpdir/endpoint.crt"
    cat "$tmpdir/endpoint.bundle.json" | jq -r '.data.private_key' >"$tmpdir/endpoint.key"
fi


echo "==> Verifying chain of trust: $tmpdir/rootCA.crt -> $tmpdir/endpoint.crt"
openssl verify -verbose -CAfile "$tmpdir/rootCA.crt" "$tmpdir/endpoint.crt"

echo "==> Created files in $tmpdir:"
ls -l "$tmpdir"

echo "==> Success"