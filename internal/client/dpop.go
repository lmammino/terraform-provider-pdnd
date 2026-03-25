package client

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// DPoPProofGenerator generates DPoP proof JWTs for PDND API authentication.
type DPoPProofGenerator struct {
	privateKey crypto.Signer
	keyID      string
	algorithm  string
	publicJWK  map[string]interface{}
	clock      func() time.Time
}

// NewDPoPProofGenerator parses a PEM-encoded private key and creates a generator.
// Returns error if PEM is invalid or key type is unsupported.
func NewDPoPProofGenerator(pemKey []byte, keyID string) (*DPoPProofGenerator, error) {
	if len(pemKey) == 0 {
		return nil, fmt.Errorf("PEM key data is empty")
	}

	block, _ := pem.Decode(pemKey)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	var (
		signer    crypto.Signer
		algorithm string
		jwk       map[string]interface{}
	)

	switch block.Type {
	case "PRIVATE KEY": // PKCS8
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
		s, ok := key.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("parsed key does not implement crypto.Signer")
		}
		signer = s

	case "RSA PRIVATE KEY": // PKCS1
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 RSA private key: %w", err)
		}
		signer = key

	case "EC PRIVATE KEY": // SEC1
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse EC private key: %w", err)
		}
		signer = key

	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	switch k := signer.(type) {
	case *rsa.PrivateKey:
		algorithm = "RS256"
		jwk = rsaPublicJWK(&k.PublicKey)
	case *ecdsa.PrivateKey:
		switch k.Curve {
		case elliptic.P256():
			algorithm = "ES256"
		case elliptic.P384():
			algorithm = "ES384"
		default:
			return nil, fmt.Errorf("unsupported EC curve: %v", k.Curve.Params().Name)
		}
		jwk = ecPublicJWK(&k.PublicKey)
	default:
		return nil, fmt.Errorf("unsupported key type: %T", signer)
	}

	return &DPoPProofGenerator{
		privateKey: signer,
		keyID:      keyID,
		algorithm:  algorithm,
		publicJWK:  jwk,
		clock:      time.Now,
	}, nil
}

// GenerateProof creates a signed DPoP proof JWT for the given HTTP method and URL.
// The URL is stripped of query string and fragment per RFC 9449.
func (g *DPoPProofGenerator) GenerateProof(method, rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	htu := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	now := g.clock()

	claims := jwt.MapClaims{
		"jti": uuid.New().String(),
		"iat": now.Unix(),
		"htm": method,
		"htu": htu,
	}

	var signingMethod jwt.SigningMethod
	switch g.algorithm {
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	case "ES256":
		signingMethod = jwt.SigningMethodES256
	case "ES384":
		signingMethod = jwt.SigningMethodES384
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", g.algorithm)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["typ"] = "dpop+jwt"
	token.Header["jwk"] = g.publicJWK

	return token.SignedString(g.privateKey)
}

func rsaPublicJWK(pub *rsa.PublicKey) map[string]interface{} {
	return map[string]interface{}{
		"kty": "RSA",
		"n":   base64URLEncodeBigInt(pub.N),
		"e":   base64URLEncodeBigInt(big.NewInt(int64(pub.E))),
	}
}

func ecPublicJWK(pub *ecdsa.PublicKey) map[string]interface{} {
	var crv string
	switch pub.Curve {
	case elliptic.P256():
		crv = "P-256"
	case elliptic.P384():
		crv = "P-384"
	}

	byteLen := (pub.Curve.Params().BitSize + 7) / 8

	return map[string]interface{}{
		"kty": "EC",
		"crv": crv,
		"x":   base64URLEncodeBigIntPadded(pub.X, byteLen),
		"y":   base64URLEncodeBigIntPadded(pub.Y, byteLen),
	}
}

func base64URLEncodeBigInt(n *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(n.Bytes())
}

func base64URLEncodeBigIntPadded(n *big.Int, size int) string {
	b := n.Bytes()
	if len(b) < size {
		padded := make([]byte, size)
		copy(padded[size-len(b):], b)
		b = padded
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
