package pki

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/helper/certutil"
)

// WrappedCertBundle wraps a vault certutil.CertBundle with support for externally
// managed private keys such as an asymmetric signing key stored in Google Cloud KMS.
type WrappedCertBundle struct {
	certutil.CertBundle `mapstructure:",squash"`
	GoogleCloudKMSKey   string
}

type WrappedParsedCertBundle struct {
	certutil.ParsedCertBundle `mapstructure:",squash"`
	GoogleCloudKMSKey         string
}

type WrappedCSRBundle struct {
	certutil.CSRBundle `mapstructure:",squash"`
	GoogleCloudKMSKey  string
}

type WrappedParsedCSRBundle struct {
	certutil.ParsedCSRBundle `mapstructure:",squash"`
	GoogleCloudKMSKey        string
}

// ToParsedCertBundle resolves any externally hosed signing keys and assigns the resolved
// crypto.Signer on the PrivateKey attribute of the ParsedCertBundle
func (c *WrappedCertBundle) ToParsedCertBundle(ctx context.Context) (*WrappedParsedCertBundle, error) {
	parsedBundle, err := c.CertBundle.ToParsedCertBundle()
	if err != nil {
		return nil, err
	}
	wrapParsedBundle := &WrappedParsedCertBundle{
		ParsedCertBundle:  *parsedBundle,
		GoogleCloudKMSKey: c.GoogleCloudKMSKey,
	}

	if c.GoogleCloudKMSKey != "" {
		signer, err := googleKMSSigner(ctx, c.GoogleCloudKMSKey)
		if err != nil {
			return nil, fmt.Errorf("unable to lookup GCP Cloud KMS Key: %v", err)
		}
		keyType := gcpKMSKeyType(signer.Algorithm())
		wrapParsedBundle.SetParsedPrivateKey(signer, keyType, []byte(""))
	}
	return wrapParsedBundle, nil
}

func (c *WrappedParsedCertBundle) ToCertBundle() (*WrappedCertBundle, error) {
	certBundle, err := c.ParsedCertBundle.ToCertBundle()
	if err != nil {
		return nil, err
	}
	wrapCertBundle := &WrappedCertBundle{
		CertBundle:        *certBundle,
		GoogleCloudKMSKey: c.GoogleCloudKMSKey,
	}
	return wrapCertBundle, nil
}

func (c *WrappedCSRBundle) ToParsedCSRBundle(ctx context.Context) (*WrappedParsedCSRBundle, error) {
	csrBundle, err := c.CSRBundle.ToParsedCSRBundle()
	if err != nil {
		return nil, err
	}
	wrapParsedBundle := &WrappedParsedCSRBundle{
		ParsedCSRBundle:   *csrBundle,
		GoogleCloudKMSKey: c.GoogleCloudKMSKey,
	}

	if c.GoogleCloudKMSKey != "" {
		signer, err := googleKMSSigner(ctx, c.GoogleCloudKMSKey)
		if err != nil {
			return nil, fmt.Errorf("unable to lookup GCP Cloud KMS Key: %v", err)
		}
		keyType := gcpKMSKeyType(signer.Algorithm())
		wrapParsedBundle.SetParsedPrivateKey(signer, keyType, []byte(""))
	}
	return wrapParsedBundle, nil
}

func (c *WrappedParsedCSRBundle) ToCSRBundle() (*WrappedCSRBundle, error) {
	csrBundle, err := c.ParsedCSRBundle.ToCSRBundle()
	if err != nil {
		return nil, err
	}
	wrapCSRBundle := &WrappedCSRBundle{
		CSRBundle:         *csrBundle,
		GoogleCloudKMSKey: c.GoogleCloudKMSKey,
	}
	wrapCSRBundle.PrivateKeyType = c.PrivateKeyType
	return wrapCSRBundle, nil
}