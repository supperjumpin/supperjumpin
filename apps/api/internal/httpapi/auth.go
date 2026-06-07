package httpapi

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type StaticAuthVerifier map[string]AuthIdentity

func (v StaticAuthVerifier) Verify(token string) (AuthIdentity, bool) {
	identity, ok := v[token]
	return identity, ok
}

type AuthVerifierChain []AuthVerifier

func (chain AuthVerifierChain) Verify(token string) (AuthIdentity, bool) {
	for _, verifier := range chain {
		if verifier == nil {
			continue
		}
		identity, ok := verifier.Verify(token)
		if ok {
			return identity, true
		}
	}
	return AuthIdentity{}, false
}

type SupabaseJWTVerifier struct {
	Secret []byte
	Now    func() time.Time
}

func (v SupabaseJWTVerifier) Verify(token string) (AuthIdentity, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || len(v.Secret) == 0 {
		return AuthIdentity{}, false
	}

	var header struct {
		Algorithm string `json:"alg"`
	}
	if !decodeJWTPart(parts[0], &header) || header.Algorithm != "HS256" {
		return AuthIdentity{}, false
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return AuthIdentity{}, false
	}
	expected := hmac.New(sha256.New, v.Secret)
	expected.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(signature, expected.Sum(nil)) {
		return AuthIdentity{}, false
	}

	var claims struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Expiry  int64  `json:"exp"`
	}
	if !decodeJWTPart(parts[1], &claims) || claims.Subject == "" {
		return AuthIdentity{}, false
	}
	if claims.Expiry > 0 {
		now := time.Now
		if v.Now != nil {
			now = v.Now
		}
		if !now().Before(time.Unix(claims.Expiry, 0)) {
			return AuthIdentity{}, false
		}
	}

	return AuthIdentity{Provider: "supabase", Subject: claims.Subject, Email: claims.Email}, true
}

type SupabaseJWKSVerifier struct {
	Keys map[string]*ecdsa.PublicKey
	Now  func() time.Time
}

func NewSupabaseJWKSVerifier(jwksURL string, now func() time.Time) (SupabaseJWKSVerifier, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(jwksURL)
	if err != nil {
		return SupabaseJWKSVerifier{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return SupabaseJWKSVerifier{}, errors.New("fetch Supabase JWKS")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SupabaseJWKSVerifier{}, err
	}
	return ParseSupabaseJWKS(body, now)
}

func ParseSupabaseJWKS(body []byte, now func() time.Time) (SupabaseJWKSVerifier, error) {
	var jwks struct {
		Keys []struct {
			KeyID     string `json:"kid"`
			KeyType   string `json:"kty"`
			Algorithm string `json:"alg"`
			Curve     string `json:"crv"`
			X         string `json:"x"`
			Y         string `json:"y"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwks); err != nil {
		return SupabaseJWKSVerifier{}, err
	}

	keys := map[string]*ecdsa.PublicKey{}
	for _, key := range jwks.Keys {
		if key.KeyID == "" || key.KeyType != "EC" || key.Algorithm != "ES256" || key.Curve != "P-256" {
			continue
		}
		x, err := base64.RawURLEncoding.DecodeString(key.X)
		if err != nil {
			continue
		}
		y, err := base64.RawURLEncoding.DecodeString(key.Y)
		if err != nil {
			continue
		}
		keys[key.KeyID] = &ecdsa.PublicKey{Curve: elliptic.P256(), X: new(big.Int).SetBytes(x), Y: new(big.Int).SetBytes(y)}
	}

	if len(keys) == 0 {
		return SupabaseJWKSVerifier{}, errors.New("Supabase JWKS has no supported ES256 keys")
	}
	return SupabaseJWKSVerifier{Keys: keys, Now: now}, nil
}

func (v SupabaseJWKSVerifier) Verify(token string) (AuthIdentity, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || len(v.Keys) == 0 {
		return AuthIdentity{}, false
	}

	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}
	if !decodeJWTPart(parts[0], &header) || header.Algorithm != "ES256" || header.KeyID == "" {
		return AuthIdentity{}, false
	}
	key, ok := v.Keys[header.KeyID]
	if !ok {
		return AuthIdentity{}, false
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(signature) != 64 {
		return AuthIdentity{}, false
	}
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if !ecdsa.Verify(key, hash[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:])) {
		return AuthIdentity{}, false
	}

	claims, ok := supabaseClaims(parts[1], v.Now)
	if !ok {
		return AuthIdentity{}, false
	}
	return AuthIdentity{Provider: "supabase", Subject: claims.Subject, Email: claims.Email}, true
}

type jwtClaims struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Expiry  int64  `json:"exp"`
}

func supabaseClaims(part string, nowFunc func() time.Time) (jwtClaims, bool) {
	var claims jwtClaims
	if !decodeJWTPart(part, &claims) || claims.Subject == "" {
		return jwtClaims{}, false
	}
	if claims.Expiry > 0 {
		now := time.Now
		if nowFunc != nil {
			now = nowFunc
		}
		if !now().Before(time.Unix(claims.Expiry, 0)) {
			return jwtClaims{}, false
		}
	}
	return claims, true
}

func decodeJWTPart(part string, destination any) bool {
	data, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, destination) == nil
}
