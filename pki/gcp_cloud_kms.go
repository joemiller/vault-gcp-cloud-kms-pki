package pki

import (
	"context"

	"github.com/hashicorp/go-gcp-common/gcputil"
	"github.com/hashicorp/vault/helper/certutil"
	"github.com/joemiller/vault-gcp-cloud-kms-pki/kmssigner"
	"golang.org/x/oauth2"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// TODO(joe): retire this code if we're not going to store creds in Storage
// func fetchGoogleCreds(ctx context.Context, s logical.Storage) (string, error) {
// 	if s == nil {
// 		return "", nil
// 	}
// 	credsEntry, err := s.Get(ctx, "google_credentials")
// 	if err != nil {
// 		return "", err
// 	}
// 	credsJSON := ""
// 	if credsEntry != nil {
// 		credsJSON = string(credsEntry.Value)
// 	}
// 	return credsJSON, nil
// }

// TODO: document this func
func newGoogleKMSClient(ctx context.Context, credsJSON string) (*cloudkms.Service, error) {
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}

	_, tokenSource, err := gcputil.FindCredentials(credsJSON, ctx, scopes...)
	if err != nil {
		return nil, err
	}

	oauthClient := oauth2.NewClient(ctx, tokenSource)
	svc, err := cloudkms.New(oauthClient)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// TODO: document this func
func kmsSigner(svc *cloudkms.Service, key string) (kmssigner.Signer, error) {
	signer, err := kmssigner.New(svc, key)
	if err != nil {
		return nil, err
	}
	return signer, nil
}

// gcpKMSKeyType converts a GCP KMS key algorithm identifier to a certutil.PrivateKeyType
func gcpKMSKeyType(algo string) certutil.PrivateKeyType {
	switch algo {
	case "RSA_SIGN_PKCS1_2048_SHA256":
		return certutil.RSAPrivateKey
	case "RSA_SIGN_PKCS1_3072_SHA256":
		return certutil.RSAPrivateKey
	case "RSA_SIGN_PKCS1_4096_SHA256":
		return certutil.RSAPrivateKey
	case "EC_SIGN_P256_SHA256":
		return certutil.ECPrivateKey
	case "EC_SIGN_P384_SHA384":
		return certutil.ECPrivateKey
	}
	return certutil.UnknownPrivateKey
}
