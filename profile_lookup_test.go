package paymail

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/bsv-blockchain/go-sdk/auth/certificates"
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/transaction"
	"github.com/bsv-blockchain/go-sdk/wallet"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The BRC-235 test vector, which names example.com
const (
	profileTestDomain = "example.com"
	profileVectorKey  = "025ad43a22ac38d0bc1f8bacaabb323b5d634703b7a774c4268f6a09e4ddf79097"
	profileVectorJSON = `{
  "type": "SbatVXXssDW3AO0J9bxIljkHGbPBGCVAXg94gXFf0cE=",
  "serialNumber": "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA=",
  "subject": "025ad43a22ac38d0bc1f8bacaabb323b5d634703b7a774c4268f6a09e4ddf79097",
  "certifier": "025ad43a22ac38d0bc1f8bacaabb323b5d634703b7a774c4268f6a09e4ddf79097",
  "revocationOutpoint": "0000000000000000000000000000000000000000000000000000000000000000.0",
  "fields": {
    "displayName": "Alice",
    "issuedAt": "2026-10-29T12:00:00.000Z",
    "paymail": "alice@example.com"
  },
  "signature": "304402201433492b6713a2f210bbd75c5d57de1085b92193f868e1add7785d35e55fc602022032ee8493f6e7d2273164c22c9ae8d8e59419e56d7ba0445663a17a9513c84c19"
}`
	profileLookupURL  = "https://paymail.example.com/v1/bsvalias/handle/{query}"
	profileReverseURL = "https://paymail.example.com/v1/bsvalias/identityKey/{pubkey}"
)

// profileVectorWith returns the test vector with fields replaced or added (the signature no longer matches)
func profileVectorWith(t *testing.T, mutate func(m map[string]any)) []byte {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(profileVectorJSON), &m))
	mutate(m)
	b, err := json.Marshal(m)
	require.NoError(t, err)
	return b
}

// signProfile signs a fresh profile certificate for paymail with key
func signProfile(ctx context.Context, t *testing.T, keyHex, paymail string) []byte {
	t.Helper()
	priv, err := ec.PrivateKeyFromHex(keyHex)
	require.NoError(t, err)
	outpoint, err := transaction.OutpointFromString(ProfileRevocationOutpoint)
	require.NoError(t, err)
	serial := sha256.Sum256([]byte(paymail))
	cert := &certificates.Certificate{
		Type:               ProfileCertificateType,
		SerialNumber:       wallet.StringBase64(base64.StdEncoding.EncodeToString(serial[:])),
		Subject:            *priv.PubKey(),
		Certifier:          *priv.PubKey(),
		RevocationOutpoint: outpoint,
		Fields: map[wallet.CertificateFieldNameUnder50Bytes]wallet.StringBase64{
			"paymail":  wallet.StringBase64(paymail),
			"issuedAt": "2026-10-29T12:00:00Z",
		},
	}
	protoWallet, err := wallet.NewCompletedProtoWallet(priv)
	require.NoError(t, err)
	require.NoError(t, cert.Sign(ctx, protoWallet))
	b, err := json.Marshal(cert)
	require.NoError(t, err)
	return b
}

func TestProfileLookupBRFCs(t *testing.T) {
	for title, id := range map[string]string{
		"public profile lookup":         BRFCProfileLookup,
		"public profile reverse lookup": BRFCProfileReverseLookup,
	} {
		spec := &BRFCSpec{Title: title, Author: "Deggen", Version: "v1.0.0"}
		require.NoError(t, spec.Generate())
		assert.Equal(t, id, spec.ID)
	}
	typeHash := sha256.Sum256([]byte("public profile lookup"))
	assert.Equal(t, ProfileCertificateType, base64.StdEncoding.EncodeToString(typeHash[:]))
}

func TestVerifyProfileCertificate(t *testing.T) {
	ctx := context.Background()

	t.Run("verifies the BRC-235 test vector", func(t *testing.T) {
		profile, err := VerifyProfileCertificate(ctx, []byte(profileVectorJSON), ProfileExpectation{
			Domain: profileTestDomain, Paymail: "Alice@Example.com", IdentityKey: strings.ToUpper(profileVectorKey),
		})
		require.NoError(t, err)
		assert.Equal(t, profileVectorKey, profile.IdentityKey)
		assert.Equal(t, "alice@example.com", profile.Paymail)
		assert.Equal(t, "alice", profile.Handle)
		assert.Equal(t, profileTestDomain, profile.Domain)
		assert.Equal(t, "Alice", profile.DisplayName)
		assert.Equal(t, "2026-10-29T12:00:00.000Z", profile.IssuedAt)
	})

	t.Run("verifies a certificate signed here", func(t *testing.T) {
		raw := signProfile(ctx, t, strings.Repeat("0", 63)+"2", "bob@example.com")
		profile, err := VerifyProfileCertificate(ctx, raw, ProfileExpectation{Domain: profileTestDomain})
		require.NoError(t, err)
		assert.Equal(t, "bob@example.com", profile.Paymail)
	})

	cases := map[string]struct {
		raw    []byte
		expect ProfileExpectation
	}{
		"another domain":        {[]byte(profileVectorJSON), ProfileExpectation{Domain: "example.org"}},
		"another exact paymail": {[]byte(profileVectorJSON), ProfileExpectation{Domain: profileTestDomain, Paymail: "bob@example.com"}},
		"another reverse key":   {[]byte(profileVectorJSON), ProfileExpectation{Domain: profileTestDomain, IdentityKey: "03" + strings.Repeat("1", 64)}},
		"no domain":             {[]byte(profileVectorJSON), ProfileExpectation{}},
		"a changed field": {profileVectorWith(t, func(m map[string]any) {
			m["fields"].(map[string]any)["displayName"] = "Mallory"
		}), ProfileExpectation{Domain: profileTestDomain}},
		"a tombstone": {profileVectorWith(t, func(m map[string]any) {
			m["fields"].(map[string]any)["released"] = "true"
		}), ProfileExpectation{Domain: profileTestDomain}},
		"a different type": {profileVectorWith(t, func(m map[string]any) {
			m["type"] = "TbatVXXssDW3AO0J9bxIljkHGbPBGCVAXg94gXFf0cE="
		}), ProfileExpectation{Domain: profileTestDomain}},
		"subject not certifier": {profileVectorWith(t, func(m map[string]any) {
			m["certifier"] = "03" + strings.Repeat("1", 64)
		}), ProfileExpectation{Domain: profileTestDomain}},
		"a revocation outpoint": {profileVectorWith(t, func(m map[string]any) {
			m["revocationOutpoint"] = strings.Repeat("1", 64) + ".0"
		}), ProfileExpectation{Domain: profileTestDomain}},
		"a short serial": {profileVectorWith(t, func(m map[string]any) {
			m["serialNumber"] = "AQID"
		}), ProfileExpectation{Domain: profileTestDomain}},
		"no issuedAt": {profileVectorWith(t, func(m map[string]any) {
			delete(m["fields"].(map[string]any), "issuedAt")
		}), ProfileExpectation{Domain: profileTestDomain}},
		"an uppercase paymail": {profileVectorWith(t, func(m map[string]any) {
			m["fields"].(map[string]any)["paymail"] = "Alice@example.com"
		}), ProfileExpectation{Domain: profileTestDomain}},
		"a non-string field": {profileVectorWith(t, func(m map[string]any) {
			m["fields"].(map[string]any)["bio"] = 1
		}), ProfileExpectation{Domain: profileTestDomain}},
		"an oversized field": {profileVectorWith(t, func(m map[string]any) {
			m["fields"].(map[string]any)["bio"] = strings.Repeat("x", 1025)
		}), ProfileExpectation{Domain: profileTestDomain}},
		"no fields":  {profileVectorWith(t, func(m map[string]any) { delete(m, "fields") }), ProfileExpectation{Domain: profileTestDomain}},
		"an array":   {[]byte("[" + profileVectorJSON + "]"), ProfileExpectation{Domain: profileTestDomain}},
		"not json":   {[]byte("nope"), ProfileExpectation{Domain: profileTestDomain}},
		"too large":  {[]byte(strings.Repeat(" ", ProfileMaxCertificateBytes+1)), ProfileExpectation{Domain: profileTestDomain}},
		"no subject": {profileVectorWith(t, func(m map[string]any) { delete(m, "subject") }), ProfileExpectation{Domain: profileTestDomain}},
	}
	for name, tc := range cases {
		t.Run("rejects "+name, func(t *testing.T) {
			profile, err := VerifyProfileCertificate(ctx, tc.raw, tc.expect)
			require.ErrorIs(t, err, ErrProfileCertificateInvalid)
			assert.Nil(t, profile)
		})
	}
}

func TestNormalizeProfileLookupQuery(t *testing.T) {
	for query, expected := range map[string]string{
		"al": "al", " Alice ": "alice", "alice@example.com": "alice", "alice@exa": "alice", "alice@": "alice",
		"alice@example.org": "", "a": "", strings.Repeat("x", 33): "", "al ice": "", "..": "", "a@b@example.com": "",
	} {
		normalized, ok := NormalizeProfileLookupQuery(query, profileTestDomain)
		assert.Equal(t, expected, normalized, query)
		assert.Equal(t, expected != "", ok, query)
	}
}

func TestClient_LookupProfiles(t *testing.T) {
	// t.Parallel() (Cannot run in parallel - issues with overriding the mock client)
	searchURL := strings.ReplaceAll(profileLookupURL, ProfileQueryTemplate, "")

	t.Run("returns verified suggestions, dropping failures", func(t *testing.T) {
		client := newTestClient(t)
		httpmock.Reset()
		tampered := profileVectorWith(t, func(m map[string]any) { m["fields"].(map[string]any)["displayName"] = "Mallory" })
		httpmock.RegisterResponder(http.MethodGet, searchURL+"al",
			httpmock.NewStringResponder(http.StatusOK, "["+string(tampered)+","+profileVectorJSON+"]"))

		response, err := client.LookupProfiles(profileLookupURL, profileTestDomain, "Al")
		require.NoError(t, err)
		require.Len(t, response.Profiles, 1)
		assert.Equal(t, "alice@example.com", response.Profiles[0].Paymail)
	})

	t.Run("returns only the exact paymail for a complete one", func(t *testing.T) {
		client := newTestClient(t)
		httpmock.Reset()
		other := signProfile(context.Background(), t, strings.Repeat("0", 63)+"4", "alicea@example.com")
		httpmock.RegisterResponder(http.MethodGet, searchURL+"alice",
			httpmock.NewStringResponder(http.StatusOK, "["+string(other)+","+profileVectorJSON+"]"))

		exact, err := client.LookupProfiles(profileLookupURL, profileTestDomain, "alice@example.com")
		require.NoError(t, err)
		require.Len(t, exact.Profiles, 1)
		assert.Equal(t, profileVectorKey, exact.Profiles[0].IdentityKey)

		partial, err := client.LookupProfiles(profileLookupURL, profileTestDomain, "alice")
		require.NoError(t, err)
		assert.Len(t, partial.Profiles, 2)
	})

	t.Run("processes at most ten results", func(t *testing.T) {
		client := newTestClient(t)
		httpmock.Reset()
		many := strings.TrimSuffix(strings.Repeat(profileVectorJSON+",", 25), ",")
		httpmock.RegisterResponder(http.MethodGet, searchURL+"al", httpmock.NewStringResponder(http.StatusOK, "["+many+"]"))

		response, err := client.LookupProfiles(profileLookupURL, profileTestDomain, "al")
		require.NoError(t, err)
		assert.Len(t, response.Profiles, ProfileLookupMaxResults)
	})

	t.Run("does not ask for a query the host would not search", func(t *testing.T) {
		client := newTestClient(t)
		httpmock.Reset()
		response, err := client.LookupProfiles(profileLookupURL, profileTestDomain, "a")
		require.NoError(t, err)
		assert.Empty(t, response.Profiles)
		assert.Equal(t, 0, httpmock.GetTotalCallCount())
	})

	t.Run("rejects an unusable template", func(t *testing.T) {
		client := newTestClient(t)
		_, err := client.LookupProfiles("https://paymail.example.com/handle", profileTestDomain, "al")
		require.ErrorIs(t, err, ErrProfileLookupInvalidURL)
		_, err = client.LookupProfiles("http://paymail.example.com/handle/{query}", profileTestDomain, "al")
		require.ErrorIs(t, err, ErrProfileLookupInvalidURL)
	})

	t.Run("rejects a body that is not an array and a host error", func(t *testing.T) {
		client := newTestClient(t)
		httpmock.Reset()
		httpmock.RegisterResponder(http.MethodGet, searchURL+"al", httpmock.NewStringResponder(http.StatusOK, profileVectorJSON))
		_, err := client.LookupProfiles(profileLookupURL, profileTestDomain, "al")
		require.ErrorIs(t, err, ErrProfileLookupBadResponse)

		httpmock.RegisterResponder(http.MethodGet, searchURL+"al", httpmock.NewStringResponder(http.StatusTooManyRequests, `{"message":"slow down"}`))
		response, err := client.LookupProfiles(profileLookupURL, profileTestDomain, "al")
		require.ErrorIs(t, err, ErrProfileLookupBadResponse)
		assert.Equal(t, http.StatusTooManyRequests, response.StatusCode)
	})
}

func TestClient_ReverseLookupProfile(t *testing.T) {
	// t.Parallel() (Cannot run in parallel - issues with overriding the mock client)
	keyURL := strings.ReplaceAll(profileReverseURL, ProfileIdentityKeyTemplate, "")
	other := "03" + strings.Repeat("1", 64)

	client := newTestClient(t)
	httpmock.Reset()
	httpmock.RegisterResponder(http.MethodGet, keyURL+profileVectorKey, httpmock.NewStringResponder(http.StatusOK, profileVectorJSON))
	httpmock.RegisterResponder(http.MethodGet, keyURL+other, httpmock.NewStringResponder(http.StatusOK, profileVectorJSON))
	httpmock.RegisterResponder(http.MethodGet, keyURL+"02"+strings.Repeat("2", 64),
		httpmock.NewStringResponder(http.StatusNotFound, `{"code":"error-profile-not-found","message":"no active handle"}`))
	httpmock.RegisterResponder(http.MethodGet, keyURL+"02"+strings.Repeat("3", 64),
		httpmock.NewStringResponder(http.StatusServiceUnavailable, `{"message":"busy"}`))

	t.Run("returns the verified profile", func(t *testing.T) {
		response, err := client.ReverseLookupProfile(profileReverseURL, profileTestDomain, strings.ToUpper(profileVectorKey))
		require.NoError(t, err)
		require.NotNil(t, response.Profile)
		assert.Equal(t, "alice@example.com", response.Profile.Paymail)
	})

	t.Run("returns no profile for a certificate of another key or a 404", func(t *testing.T) {
		response, err := client.ReverseLookupProfile(profileReverseURL, profileTestDomain, other)
		require.NoError(t, err)
		assert.Nil(t, response.Profile)

		response, err = client.ReverseLookupProfile(profileReverseURL, profileTestDomain, "02"+strings.Repeat("2", 64))
		require.NoError(t, err)
		assert.Nil(t, response.Profile)
		assert.Equal(t, http.StatusNotFound, response.StatusCode)
	})

	t.Run("reports host errors and bad input", func(t *testing.T) {
		_, err := client.ReverseLookupProfile(profileReverseURL, profileTestDomain, "02"+strings.Repeat("3", 64))
		require.ErrorIs(t, err, ErrProfileLookupBadResponse)
		_, err = client.ReverseLookupProfile(profileReverseURL, profileTestDomain, "nope")
		require.ErrorIs(t, err, ErrProfileLookupInvalidKey)
		_, err = client.ReverseLookupProfile("https://paymail.example.com/identityKey", profileTestDomain, profileVectorKey)
		require.ErrorIs(t, err, ErrProfileLookupInvalidURL)
	})
}
