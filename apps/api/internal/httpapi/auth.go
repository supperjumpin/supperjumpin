package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
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

func decodeJWTPart(part string, destination any) bool {
	data, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, destination) == nil
}
