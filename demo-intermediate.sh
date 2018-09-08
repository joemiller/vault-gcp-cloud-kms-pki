#!/bin/bash
set -e

ROOT_KEY="${ROOT_KEY:-projects/joe-vault-kms-dev/locations/us-central1/keyRings/joe-hsm-testing/cryptoKeys/ca-signer/cryptoKeyVersions/1}"
INTERMEDIATE_KEY="${INTERMEDIATE_KEY:-projects/joe-vault-kms-dev/locations/us-central1/keyRings/joe-hsm-testing/cryptoKeys/ca-signer/cryptoKeyVersions/1}"

tmpdir="./tmp-intermediate"
mkdir -p "$tmpdir"
rm -f "$tmpdir/*"

export VAULT_CONFIG_PATH=/dev/null
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

# enable two pki backends: /pki-root (root ca), /pki-intermediate (intermediate ca)
vault secrets enable -plugin-name=vault-gcp-cloud-kms-pki -path=pki-root         plugin
vault secrets enable -plugin-name=vault-gcp-cloud-kms-pki -path=pki-intermediate plugin

# generate self-signed cert for the root CA
root_crt="$tmpdir/rootCA.crt"
echo "==> Generating self-signed rootCA cert: $root_crt"
vault write \
    -field=certificate \
    pki-root/root/generate/internal \
    common_name=rootCA \
    google_cloud_kms_key="$ROOT_KEY" \
    > "$root_crt"

# generate a CSR from the intermediate CA
int_csr="$tmpdir/intCA.csr"
echo "==> Generating Intermediate CA CSR: $int_csr"
vault write -field=csr \
    pki-intermediate/intermediate/generate/internal \
    common_name=intermediateCA \
    google_cloud_kms_key="$INTERMEDIATE_KEY" \
    > "$int_csr"

# sign the intermediate CSR with the root CA. The result is the intermediate CA's cert
int_crt="$tmpdir/intCA.crt"
echo "==> Signining Intermediate CA crt: $int_crt"
cat "$tmpdir/intCA.csr" | \
    vault write -field=certificate pki-root/root/sign-intermediate csr=- format=pem_bundle \
    > "$int_crt"

# set the signed intermediate Cert into the pki-intermediate backend to be used for signing
echo "==> Pushing signed intermediate CA crt into Vault"
cat "$tmpdir/intCA.crt" | \
    vault write pki-intermediate/intermediate/set-signed certificate=-

# sign/create a new cert from the interemediate CA
echo "==> Creating an end-entity key and cert signed by the intermediate CA"
vault write pki-intermediate/roles/any allowed_domains=example.com allow_subdomains=true max_ttl=24h

vault write -format=json pki-intermediate/issue/any common_name=foo.example.com ttl=1h >"$tmpdir/endpoint.bundle.json"

if command -v jq >/dev/null; then
    cat "$tmpdir/endpoint.bundle.json" | jq -r '.data.certificate' >"$tmpdir/endpoint.crt"
    cat "$tmpdir/endpoint.bundle.json" | jq -r '.data.private_key' >"$tmpdir/endpoint.key"
fi

echo "==> Created files in $tmpdir:"
ls -l "$tmpdir"

echo "==> Verifying chain of trust: $tmpdir/rootCA.crt -> $tmpdir/intCA.crt"
openssl verify -verbose -CAfile "$tmpdir/rootCA.crt" "$tmpdir/intCA.crt"

echo "==> Verifying chain of trust: $tmpdir/rootCA.crt -> $tmpdir/intCA.crt -> $tmpdir/endpoint.crt"
openssl verify -verbose -CAfile <(cat "$tmpdir/intCA.crt"; echo; cat "$tmpdir/rootCA.crt") "$tmpdir/endpoint.crt"

echo "==> Success"