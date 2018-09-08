package pki

import (
	"context"

	"github.com/hashicorp/vault/helper/certutil"
	"github.com/hashicorp/vault/helper/errutil"
	"github.com/joemiller/vault-gcp-cloud-kms-pki/kmssigner"
	"golang.org/x/oauth2/google"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

func googleKMSSigner(ctx context.Context, keyName string) (kmssigner.Signer, error) {
	oauthClient, err := google.DefaultClient(ctx, cloudkms.CloudPlatformScope)
	if err != nil {
		return nil, errutil.InternalError{Err: err.Error()}
	}
	svc, err := cloudkms.New(oauthClient)
	if err != nil {
		return nil, errutil.InternalError{Err: err.Error()}
	}
	signer, err := kmssigner.New(svc, keyName)
	if err != nil {
		return nil, errutil.InternalError{Err: err.Error()}
	}
	return signer, nil
}

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
