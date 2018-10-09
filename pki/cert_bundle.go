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
	GoogleCredentials   string
}

type WrappedParsedCertBundle struct {
	certutil.ParsedCertBundle `mapstructure:",squash"`
	GoogleCloudKMSKey         string
	GoogleCredentials         string
}

type WrappedCSRBundle struct {
	certutil.CSRBundle `mapstructure:",squash"`
	GoogleCloudKMSKey  string
	GoogleCredentials  string
}

type WrappedParsedCSRBundle struct {
	certutil.ParsedCSRBundle `mapstructure:",squash"`
	GoogleCloudKMSKey        string
	GoogleCredentials        string
}

// GooglePrivateKeyContainer extends the upstream ParsedPrivateKeyContainer interface by adding
// a method for setting Google specific parameters on the WrappedParsed{Cert,CSR}Bundle objects
type GooglePrivateKeyContainer interface {
	certutil.ParsedPrivateKeyContainer
	SetGoogleParams(key, creds string)
}

// ToParsedCertBundle resolves any externally hosted signing keys then adds the parsed
// crypto.Signer on the PrivateKey attribute of the ParsedCertBundle
func (c *WrappedCertBundle) ToParsedCertBundle(ctx context.Context) (*WrappedParsedCertBundle, error) {
	// first, parse the embedded certBundle
	parsedBundle, err := c.CertBundle.ToParsedCertBundle()
	if err != nil {
		return nil, err
	}
	// wrap the embedded cert bundle
	wrapParsedBundle := &WrappedParsedCertBundle{
		ParsedCertBundle: *parsedBundle,
	}

	// if the bundle contains an externally hosted Google KMS key, parse it and set the values on the wrapped bundle
	if c.GoogleCloudKMSKey != "" {
		if err := parseGooglePrivateKey(ctx, c.GoogleCloudKMSKey, c.GoogleCredentials, wrapParsedBundle); err != nil {
			return nil, err
		}
	}
	return wrapParsedBundle, nil
}

func (c *WrappedParsedCertBundle) SetGoogleParams(key, creds string) {
	c.GoogleCloudKMSKey = key
	c.GoogleCredentials = creds
}

func (c *WrappedParsedCertBundle) ToCertBundle() (*WrappedCertBundle, error) {
	certBundle, err := c.ParsedCertBundle.ToCertBundle()
	if err != nil {
		return nil, err
	}
	wrapCertBundle := &WrappedCertBundle{
		CertBundle:        *certBundle,
		GoogleCloudKMSKey: c.GoogleCloudKMSKey,
		GoogleCredentials: c.GoogleCredentials,
	}
	return wrapCertBundle, nil
}

func (c *WrappedCSRBundle) ToParsedCSRBundle(ctx context.Context) (*WrappedParsedCSRBundle, error) {
	// first, parse the embedded csrBundle
	csrBundle, err := c.CSRBundle.ToParsedCSRBundle()
	if err != nil {
		return nil, err
	}
	// wrap the embedded cert bundle
	wrapParsedBundle := &WrappedParsedCSRBundle{
		ParsedCSRBundle: *csrBundle,
	}

	// if the bundle contains an externally hosted Google KMS key, parse it and set the values on the wrapped bundle
	if c.GoogleCloudKMSKey != "" {
		if err := parseGooglePrivateKey(ctx, c.GoogleCloudKMSKey, c.GoogleCredentials, wrapParsedBundle); err != nil {
			return nil, err
		}
	}
	return wrapParsedBundle, nil
}

func (c *WrappedParsedCSRBundle) SetGoogleParams(key, creds string) {
	c.GoogleCloudKMSKey = key
	c.GoogleCredentials = creds
}

func (c *WrappedParsedCSRBundle) ToCSRBundle() (*WrappedCSRBundle, error) {
	csrBundle, err := c.ParsedCSRBundle.ToCSRBundle()
	if err != nil {
		return nil, err
	}
	wrapCSRBundle := &WrappedCSRBundle{
		CSRBundle:         *csrBundle,
		GoogleCloudKMSKey: c.GoogleCloudKMSKey,
		GoogleCredentials: c.GoogleCredentials,
	}
	return wrapCSRBundle, nil
}

// parseGooglePrivateKey parses a Google KMS key identified by key path and optionally service-account
// credentials in JSON format into a crypto.Signer and adds the parsed values to the GooglePrivateKeyContainer
// object. If the credentials are empty the Google SDK Default Application Credentials methods will attempt to discover credentials to use.
func parseGooglePrivateKey(ctx context.Context, key, creds string, c GooglePrivateKeyContainer) error {
	kms, err := newGoogleKMSClient(ctx, creds)
	if err != nil {
		return err
	}
	signer, err := kmsSigner(kms, key)
	if err != nil {
		return fmt.Errorf("unable to lookup GCP Cloud KMS Key: %v", err)
	}
	keyType := gcpKMSKeyType(signer.Algorithm())

	c.SetParsedPrivateKey(signer, keyType, []byte(""))
	c.SetGoogleParams(key, creds)
	return nil
}
