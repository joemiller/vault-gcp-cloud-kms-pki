vault-gcp-cloud-kms-pki
=======================

A fork of the @hashicorp Vault PKI backend with support for root and intermediate CA keys stored
in Google Cloud KMS.

Storing the CA keys in Cloud KMS, including the new hardware backed Cloud HSM modules, allows for
significantly improved security. The keys cannot be copied. The keys never leave the Google Cloud
APIs. Signing operations are sent to the API, signed by the protected key and the result returned.

The goal is not to maintain this project as a standalone plugin. Ideally this code can make its
way back to the core Vault PKI backend.

Usage
-----

TODO ...

Examples
--------

TODO...

notes
-----

```shell
# create a keyring:
gcloud beta kms keyrings create \
  my-keyring \
  --project my-project \
  --location us-west1


# creating asymmetric hardware-backed KMS key:
gcloud alpha kms keys create \
  root-ca \
  --keyring my-keyring \
  --purpose asymmetric-signing \
  --default-algorithm rsa-sign-pkcs1-2048-sha256 \
  --project my-project \
  --location us-west1 \
  --protection-level hsm

# get the full path of a key, in a format suitable for use with Vault pki "google_cloud_kms_key=" attribute
echo "$(gcloud alpha kms keys describe root-ca --location us-west1 --keyring my-keyring --project my-project --format="value(name)")/cryptoKeyVersions/1
```

TODO
----

- [x] rename project, hsm->kms (directory and import, and git repo)
- [ ] license
- [ ] readme docs
  - [x] intro
  - [ ] usage
  - [ ] examples
    - [ ] simple root CA
    - [ ] root CA + intermediate CA (on kms)
    - [ ] root CA (offline/yubi) + intermediate CA (on kms)
- [ ] circleci
  - [ ] deps, test.. goreleaser for releases? yeah sure