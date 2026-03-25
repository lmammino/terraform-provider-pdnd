package client_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lmammino/terraform-provider-pdnd/internal/client"
)

func generateTestRSAKey() (*rsa.PrivateKey, []byte) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return key, pemBytes
}

func generateTestECKey() (*ecdsa.PrivateKey, []byte) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	ecBytes, _ := x509.MarshalECPrivateKey(key)
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: ecBytes,
	})
	return key, pemBytes
}

func TestNewDPoPProofGenerator_RSA(t *testing.T) {
	_, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gen == nil {
		t.Fatal("expected generator, got nil")
	}
}

func TestNewDPoPProofGenerator_EC(t *testing.T) {
	_, pemBytes := generateTestECKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gen == nil {
		t.Fatal("expected generator, got nil")
	}
}

func TestNewDPoPProofGenerator_InvalidPEM(t *testing.T) {
	_, err := client.NewDPoPProofGenerator([]byte("not a valid PEM"), "test-key-id")
	if err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
}

func TestNewDPoPProofGenerator_NilKey(t *testing.T) {
	_, err := client.NewDPoPProofGenerator(nil, "test-key-id")
	if err == nil {
		t.Fatal("expected error for nil key, got nil")
	}

	_, err = client.NewDPoPProofGenerator([]byte{}, "test-key-id")
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
}

func parseProofToken(t *testing.T, tokenString string, pubKey interface{}) *jwt.Token {
	t.Helper()
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return pubKey, nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	return token
}

func TestGenerateProof_ContainsCorrectMethod(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	methods := []string{"GET", "POST", "DELETE"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			proof, err := gen.GenerateProof(method, "https://api.example.com/v3/agreements")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			token := parseProofToken(t, proof, &key.PublicKey)
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				t.Fatal("unexpected claims type")
			}
			htm, ok := claims["htm"].(string)
			if !ok || htm != method {
				t.Errorf("expected htm=%q, got %v", method, claims["htm"])
			}
		})
	}
}

func TestGenerateProof_ContainsCorrectURL(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proof, err := gen.GenerateProof("GET", "https://api.example.com/v3/agreements?offset=0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token := parseProofToken(t, proof, &key.PublicKey)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("unexpected claims type")
	}
	htu, ok := claims["htu"].(string)
	if !ok {
		t.Fatalf("htu claim missing or not a string")
	}
	expected := "https://api.example.com/v3/agreements"
	if htu != expected {
		t.Errorf("expected htu=%q, got %q", expected, htu)
	}
}

func TestGenerateProof_UniqueJTI(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proof1, _ := gen.GenerateProof("GET", "https://api.example.com/v3/test")
	proof2, _ := gen.GenerateProof("GET", "https://api.example.com/v3/test")

	token1 := parseProofToken(t, proof1, &key.PublicKey)
	token2 := parseProofToken(t, proof2, &key.PublicKey)

	claims1, _ := token1.Claims.(jwt.MapClaims)
	claims2, _ := token2.Claims.(jwt.MapClaims)
	jti1, _ := claims1["jti"].(string)
	jti2, _ := claims2["jti"].(string)

	if jti1 == jti2 {
		t.Errorf("expected unique jti values, both were %q", jti1)
	}
}

func TestGenerateProof_IATIsRecent(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now()
	proof, _ := gen.GenerateProof("GET", "https://api.example.com/v3/test")
	token := parseProofToken(t, proof, &key.PublicKey)

	iatFloat, ok := token.Claims.(jwt.MapClaims)["iat"].(float64)
	if !ok {
		t.Fatal("iat claim missing or not a number")
	}
	iat := time.Unix(int64(iatFloat), 0)
	diff := math.Abs(float64(now.Unix() - iat.Unix()))
	if diff > 5 {
		t.Errorf("iat is not recent: diff=%v seconds", diff)
	}
}

func TestGenerateProof_HeaderTyp(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proof, _ := gen.GenerateProof("GET", "https://api.example.com/v3/test")
	token := parseProofToken(t, proof, &key.PublicKey)

	typ, ok := token.Header["typ"].(string)
	if !ok || typ != "dpop+jwt" {
		t.Errorf("expected typ=dpop+jwt, got %v", token.Header["typ"])
	}
}

func TestGenerateProof_HeaderJWK(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proof, _ := gen.GenerateProof("GET", "https://api.example.com/v3/test")
	token := parseProofToken(t, proof, &key.PublicKey)

	jwk, ok := token.Header["jwk"].(map[string]interface{})
	if !ok {
		t.Fatal("jwk header missing or not a map")
	}

	kty, ok := jwk["kty"].(string)
	if !ok || kty != "RSA" {
		t.Errorf("expected kty=RSA, got %v", jwk["kty"])
	}

	if _, ok := jwk["n"]; !ok {
		t.Error("jwk missing 'n' field")
	}
	if _, ok := jwk["e"]; !ok {
		t.Error("jwk missing 'e' field")
	}
}

func TestGenerateProof_HeaderJWK_EC(t *testing.T) {
	key, pemBytes := generateTestECKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proof, _ := gen.GenerateProof("GET", "https://api.example.com/v3/test")
	token := parseProofToken(t, proof, &key.PublicKey)

	jwk, ok := token.Header["jwk"].(map[string]interface{})
	if !ok {
		t.Fatal("jwk header missing or not a map")
	}

	if jwk["kty"] != "EC" {
		t.Errorf("expected kty=EC, got %v", jwk["kty"])
	}
	if jwk["crv"] != "P-256" {
		t.Errorf("expected crv=P-256, got %v", jwk["crv"])
	}
	if _, ok := jwk["x"]; !ok {
		t.Error("jwk missing 'x' field")
	}
	if _, ok := jwk["y"]; !ok {
		t.Error("jwk missing 'y' field")
	}
}

func TestGenerateProof_SignatureValid(t *testing.T) {
	key, pemBytes := generateTestRSAKey()
	gen, err := client.NewDPoPProofGenerator(pemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proof, _ := gen.GenerateProof("GET", "https://api.example.com/v3/test")

	// Parse with validation (signature check).
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	_, err = parser.Parse(proof, func(token *jwt.Token) (interface{}, error) {
		return &key.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("signature validation failed: %v", err)
	}

	// Also test with EC key.
	ecKey, ecPemBytes := generateTestECKey()
	ecGen, err := client.NewDPoPProofGenerator(ecPemBytes, "test-key-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ecProof, _ := ecGen.GenerateProof("POST", "https://api.example.com/v3/test")

	_, err = parser.Parse(ecProof, func(token *jwt.Token) (interface{}, error) {
		return &ecKey.PublicKey, nil
	})
	if err != nil {
		t.Fatalf("EC signature validation failed: %v", err)
	}
}
