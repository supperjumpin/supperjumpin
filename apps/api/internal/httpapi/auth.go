package httpapi

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
