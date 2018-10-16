vault-gcp-cloud-kms-pki
=======================

A fork of the **@hashicorp** Vault PKI backend with support for CA keys stored securely in
Google Cloud KMS.

Storing the CA keys in Cloud KMS, including the new hardware backed Cloud HSM modules, allows for
significantly improved security. The keys cannot be copied. The keys never leave the Google Cloud
APIs.

The goal is not to maintain this project as a standalone plugin. Ideally this code can make its
way back to the builtin Vault PKI backend.

Code for accessing KMS keys is from [heptiolabs/google-kms-pgp](https://github.com/heptiolabs/google-kms-pgp)

Setup
-----

The setup guide assumes some familiarity with Vault and Vault's plugin
ecosystem. You must have a Vault server already running, unsealed, and
authenticated.

1. Download and decompress the latest plugin binary from the [releases](https://github.com/joemiller/vault-gcp-cloud-kms-pki/releases)
   tab on GitHub. Alternatively you can compile the plugin from source (`make build`).

2. Move the compiled plugin into Vault's configured `plugin_directory`:

    ```sh
    mv vault-gcp-cloud-kms-pki /etc/vault/plugins/vault-gcp-cloud-kms-pki
    ```

3. Calculate the SHA256 of the plugin and register it in Vault's plugin catalog.
   If you are downloading the pre-compiled binary, it is highly recommended that
   you use the published checksums to verify integrity.

    ```sh
    export SHA256=$(shasum -a 256 "/etc/vault/plugins/vault-gcp-cloud-kms-pki" | cut -d' ' -f1)

    vault write sys/plugins/catalog/vault-gcp-cloud-kms-pki \
      sha_256="${SHA256}" \
      command="vault-gcp-cloud-kms-pki" \
    ```

4. Mount the secrets backend:

    ```sh
    vault secrets enable \
      -path="pki" \
      -plugin-name="vault-gcp-cloud-kms-pki" plugin
    ```

5. To configure a new root CA backed by a KMS signing key:

    ```sh
    vault write \
        pki/root/generate/internal \
        common_name=rootCA \
        google_cloud_kms_key="projects/my-project/locations/us-west1/keyRings/my-keyring/cryptoKeys/root-ca/cryptoKeyVersions/1" \
        google_credentials=@service-account.json
    ```

6. Or, to configure a new intermediate CA backed by a KMS signing key:

    ```sh
    vault write \
        pki/intermediate/generate/internal \
        common_name=intCA \
        google_cloud_kms_key="projects/my-project/locations/us-west1/keyRings/my-keyring/cryptoKeys/root-ca/cryptoKeyVersions/1" \
        google_credentials=@service-account.json
    ```

The plugin follows the same rules as the builtin Vault PKI backend - only root or intermediate
can be used at a time by each instance of the plugin. Use additional pki mounts to support
multiple CA's.

The `google_credentials` attribute is optional. The plugin uses the official Google Cloud Golang SDK
which means it supports the [common ways of providing credentials](https://cloud.google.com/docs/authentication/production#providing_credentials_to_your_application)
to Google Cloud.

In addition to specifying credentials directly via Vault configuration, you can also get
configuration from the following values on the Vault server:

1. The environment variable `GOOGLE_APPLICATION_CREDENTIALS`. This is specified as the path to a
   Google Cloud credentials file, typically for a service account. If this environment variable
   is present, the resulting credentials are used. If the credentials are invalid, an error is
   returned.

2. Default instance credentials. When no environment variable is present, the default service
   account credentials are used. This is useful when running Vault on Google Compute Engine or
   Google Kubernetes Engine

Google Cloud KMS Keys
---------------------

In order to use a Google Cloud KMS key with the backend the keys must be created outside of Vault.

The keys must be asymmetric-signing keys.

1. First, create a keyring to hold the keys:

    ```sh
    gcloud beta kms keyrings create \
      my-keyring \
      --project my-project \
      --location us-west1
    ```

2. Next, create an **asymmetric-signing** key named `root-ca`. Only RSA PKCS#1 keys are currently
   supported:

    ```sh
    gcloud alpha kms keys create \
      root-ca \
      --keyring my-keyring \
      --purpose asymmetric-signing \
      --default-algorithm rsa-sign-pkcs1-2048-sha256 \
      --project my-project \
      --location us-west1 \
      --protection-level hsm
    ```

3. Get the full path of a key, in a format suitable for use with `"google_cloud_kms_key="` attribute:

    ```sh
    echo "$(gcloud alpha kms keys describe root-ca --location us-west1 --keyring my-keyring --project my-project --format="value(name)")/cryptoKeyVersions/1

    # projects/my-project/locations/us-west1/keyRings/my-keyring/cryptoKeys/root-ca/cryptoKeyVersions/1`
    ```

Google Service Account
----------------------

The credentials given to the service account used by Vault must include the permissions:

* cloudkms.cryptoKeyVersions.get
* cloudkms.cryptoKeyVersions.useToSign
* cloudkms.cryptoKeyVersions.viewPublicKey

The simplest approach is to assign the pre-defined roles `roles/viewer` and
roles/cloudkms.signerVerifier` to the service account. Alternatively, create
a custom role with only these permissions assigned.

Example:

1. Create service account:

    ```sh
    gcloud iam service-accounts create "vault-kms" --project "my-project"
    ```

2. Assign roles to the service account:

    ```sh
    gcloud projects add-iam-policy-binding "my-project" \
    --member serviceAccount:"vault-kms@my-project.iam.gserviceaccount.com" \
    --role roles/cloudkms.signerVerifier

    gcloud projects add-iam-policy-binding "my-project" \
    --member serviceAccount:"vault-kms@my-project.iam.gserviceaccount.com" \
    --role roles/viewer
    ```

3. Create credentials JSON file `service-account.json`:

    ```sh
    gcloud iam service-accounts keys create "service-account.json" \
    --iam-account="vault-kms@my-project.iam.gserviceaccount.com" \
    --project "my-project"
    ```

Example Scripts
---------------

The `./example` directory contains scripts that walk through the steps of setting up the plugin
for use with KMS keys and signing a single leaf certificate. The output of the CA certs and
the leaf certs will be in `./tmp-{rootca,intermediate}` for inspection.

* `example-rootca.sh`
* `example-intermediate.sh`

KMS keys will need to be created before running the scripts. Set the path to the keys
in the `ROOT_KEY` and `INTERMEDIATE_KEY` variables. Both scripts will use default current
Google credentials of your current installation.

Tests
-----

The original Vault PKI plugin uses unit tests. Run `make test` to execute these tests.

The Google KMS code is only run when acceptance tests are enabled. You will need to create two signing keys, one will
be used for the root CA and another for the intermediate CA.

In order to run the acceptance tests you must set the following environment variables:

* `VAULT_ACC`: set this to enable the Google KMS tests
* `TEST_GOOGLE_KMS_ROOT_KEY`: Set this to an asymmetric RSA key in Google KMS, including the key version, eg:
  `projects/my-project/locations/us-west1/keyRings/my-key-ring/cryptoKeys/root-ca/cryptoKeyVersions/1`
* `TEST_GOOGLE_KMS_INTERMEDIATE_KEY`: Set this to an asymmetric RSA key in Google KMS, including the key version, eg:
  `projects/my-project/locations/us-west1/keyRings/my-key-ring/cryptoKeys/intermediate-ca/cryptoKeyVersions/1`
* `TEST_GOOGLE_CREDENTIALS_FILE`: Set this to the path of a GCP service-account JSON key which has the required
  permissions to use the root and intermediate KMS keys for signing.

Run `make test` after setting the above env vars set to execute the Google KMS acceptance tests.

TODO
----

* [ ] add support for EC keys
* [x] setup cirlceci to run acceptance tests
* [ ] upstream into core vault...