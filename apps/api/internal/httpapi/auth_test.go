package httpapi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

func TestSupabaseJWKSVerifierAcceptsValidES256Token(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	verifier, err := ParseSupabaseJWKS(testJWKS(t, "current-key", &privateKey.PublicKey), func() time.Time {
		return time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	})
	if err != nil {
		t.Fatalf("parse JWKS: %v", err)
	}
	token := signTestES256JWT(t, privateKey, "current-key", map[string]any{
		"sub":   "auth-user-123",
		"email": "player@example.com",
		"exp":   time.Date(2026, 6, 7, 13, 0, 0, 0, time.UTC).Unix(),
	})

	identity, ok := verifier.Verify(token)

	if !ok {
		t.Fatalf("expected token to verify")
	}
	if identity.Provider != "supabase" || identity.Subject != "auth-user-123" || identity.Email != "player@example.com" {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}

func TestSupabaseJWKSVerifierRejectsUnknownKeyID(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	verifier, err := ParseSupabaseJWKS(testJWKS(t, "current-key", &privateKey.PublicKey), func() time.Time {
		return time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	})
	if err != nil {
		t.Fatalf("parse JWKS: %v", err)
	}
	token := signTestES256JWT(t, privateKey, "standby-key", map[string]any{
		"sub": "auth-user-123",
		"exp": time.Date(2026, 6, 7, 13, 0, 0, 0, time.UTC).Unix(),
	})

	_, ok := verifier.Verify(token)

	if ok {
		t.Fatalf("expected unknown key id to be rejected")
	}
}

func TestSupabaseJWTVerifierAcceptsValidToken(t *testing.T) {
	secret := []byte("test-jwt-secret")
	token := signTestHS256JWT(t, secret, map[string]any{
		"sub":   "auth-user-123",
		"email": "player@example.com",
		"exp":   time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC).Add(time.Hour).Unix(),
	})

	identity, ok := SupabaseJWTVerifier{
		Secret: secret,
		Now:    func() time.Time { return time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC) },
	}.Verify(token)

	if !ok {
		t.Fatalf("expected token to verify")
	}
	if identity.Provider != "supabase" || identity.Subject != "auth-user-123" || identity.Email != "player@example.com" {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}

func TestSupabaseJWTVerifierRejectsInvalidSignature(t *testing.T) {
	token := signTestHS256JWT(t, []byte("correct-secret"), map[string]any{
		"sub": "auth-user-123",
		"exp": time.Date(2026, 6, 7, 13, 0, 0, 0, time.UTC).Unix(),
	})

	_, ok := SupabaseJWTVerifier{
		Secret: []byte("wrong-secret"),
		Now:    func() time.Time { return time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC) },
	}.Verify(token)

	if ok {
		t.Fatalf("expected invalid signature to be rejected")
	}
}

func TestSupabaseJWTVerifierRejectsExpiredToken(t *testing.T) {
	secret := []byte("test-jwt-secret")
	token := signTestHS256JWT(t, secret, map[string]any{
		"sub": "auth-user-123",
		"exp": time.Date(2026, 6, 7, 11, 0, 0, 0, time.UTC).Unix(),
	})

	_, ok := SupabaseJWTVerifier{
		Secret: secret,
		Now:    func() time.Time { return time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC) },
	}.Verify(token)

	if ok {
		t.Fatalf("expected expired token to be rejected")
	}
}

func TestAuthVerifierChainUsesFirstVerifierThatAcceptsToken(t *testing.T) {
	chain := AuthVerifierChain{
		StaticAuthVerifier{},
		StaticAuthVerifier{"valid-token": {Provider: "supabase", Subject: "subject", Email: "player@example.com"}},
	}

	identity, ok := chain.Verify("valid-token")
	if !ok {
		t.Fatalf("expected chain to verify token")
	}
	if identity.Subject != "subject" {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}

func signTestHS256JWT(t *testing.T, secret []byte, claims map[string]any) string {
	t.Helper()
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	signature := hmac.New(sha256.New, secret)
	signature.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil))
}

func signTestES256JWT(t *testing.T, privateKey *ecdsa.PrivateKey, keyID string, claims map[string]any) string {
	t.Helper()
	header := map[string]any{"alg": "ES256", "typ": "JWT", "kid": keyID}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	hash := sha256.Sum256([]byte(unsigned))
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	signature := append(padBigInt(r, 32), padBigInt(s, 32)...)
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func testJWKS(t *testing.T, keyID string, publicKey *ecdsa.PublicKey) []byte {
	t.Helper()
	body := map[string]any{
		"keys": []map[string]any{{
			"kid": keyID,
			"kty": "EC",
			"alg": "ES256",
			"crv": "P-256",
			"x":   base64.RawURLEncoding.EncodeToString(padBigInt(publicKey.X, 32)),
			"y":   base64.RawURLEncoding.EncodeToString(padBigInt(publicKey.Y, 32)),
		}},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal JWKS: %v", err)
	}
	return jsonBody
}

func padBigInt(value *big.Int, size int) []byte {
	bytes := value.Bytes()
	if len(bytes) >= size {
		return bytes
	}
	padding := make([]byte, size-len(bytes))
	return append(padding, bytes...)
}
