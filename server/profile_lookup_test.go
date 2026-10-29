package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bsv-blockchain/go-paymail"
)

// The BRC-235 test vector names example.com, so these servers serve example.com.
const (
	profileTestDomain    = "example.com"
	profileVectorKey     = "025ad43a22ac38d0bc1f8bacaabb323b5d634703b7a774c4268f6a09e4ddf79097"
	profileVectorJSON    = `{"type":"SbatVXXssDW3AO0J9bxIljkHGbPBGCVAXg94gXFf0cE=","serialNumber":"AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA=","subject":"025ad43a22ac38d0bc1f8bacaabb323b5d634703b7a774c4268f6a09e4ddf79097","certifier":"025ad43a22ac38d0bc1f8bacaabb323b5d634703b7a774c4268f6a09e4ddf79097","revocationOutpoint":"0000000000000000000000000000000000000000000000000000000000000000.0","fields":{"displayName":"Alice","issuedAt":"2026-10-29T12:00:00.000Z","paymail":"alice@example.com"},"signature":"304402201433492b6713a2f210bbd75c5d57de1085b92193f868e1add7785d35e55fc602022032ee8493f6e7d2273164c22c9ae8d8e59419e56d7ba0445663a17a9513c84c19"}`
	profileTombstoneJSON = `{"type":"SbatVXXssDW3AO0J9bxIljkHGbPBGCVAXg94gXFf0cE=","serialNumber":"AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA=","subject":"025ad43a22ac38d0bc1f8bacaabb323b5d634703b7a774c4268f6a09e4ddf79097","certifier":"025ad43a22ac38d0bc1f8bacaabb323b5d634703b7a774c4268f6a09e4ddf79097","revocationOutpoint":"0000000000000000000000000000000000000000000000000000000000000000.0","fields":{"issuedAt":"2026-10-29T12:00:00.000Z","paymail":"alice@example.com","released":"true"},"signature":"00"}`
)

type mockProfileLookupProvider struct {
	lookup   []json.RawMessage
	reverse  json.RawMessage
	err      error
	queries  []string
	domains  []string
	reverses []string
}

func (m *mockProfileLookupProvider) LookupProfiles(_ context.Context, domain, query string) ([]json.RawMessage, error) {
	m.domains = append(m.domains, domain)
	m.queries = append(m.queries, query)
	return m.lookup, m.err
}

func (m *mockProfileLookupProvider) ReverseLookupProfile(_ context.Context, domain, identityKey string) (json.RawMessage, error) {
	m.domains = append(m.domains, domain)
	m.reverses = append(m.reverses, identityKey)
	if identityKey != profileVectorKey {
		return nil, m.err
	}
	return m.reverse, m.err
}

// profileOnlyServer hosts the two BRC-235 capabilities and nothing else; no
// PaymailServiceProvider is registered.
func profileOnlyServer(t *testing.T, provider *mockProfileLookupProvider) http.Handler {
	t.Helper()
	locator := new(PaymailServiceLocator)
	locator.RegisterProfileLookupService(provider)
	config, err := NewConfig(locator,
		WithDomain(profileTestDomain), WithoutGenericCapabilities(), WithProfileLookupCapabilities())
	require.NoError(t, err)
	return Handlers(config)
}

var errRegistryDown = errors.New("registry down")

func get(handler http.Handler, host, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+host+path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestProfileLookup_HostedAlone(t *testing.T) {
	provider := &mockProfileLookupProvider{lookup: []json.RawMessage{json.RawMessage(profileVectorJSON)}, reverse: json.RawMessage(profileVectorJSON)}
	handler := profileOnlyServer(t, provider)

	rec := get(handler, profileTestDomain, "/.well-known/bsvalias")
	require.Equal(t, http.StatusOK, rec.Code)
	var payload paymail.CapabilitiesPayload
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	assert.Equal(t, map[string]interface{}{
		paymail.BRFCProfileLookup:        "https://example.com/v1/bsvalias/handle/{query}",
		paymail.BRFCProfileReverseLookup: "https://example.com/v1/bsvalias/identityKey/{pubkey}",
	}, payload.Capabilities)

	// The legacy routes are not mounted
	assert.Equal(t, http.StatusNotFound, get(handler, profileTestDomain, "/v1/bsvalias/id/alice@example.com").Code)
}

func TestProfileLookup_Search(t *testing.T) {
	t.Run("serves servable certificates, capped at ten", func(t *testing.T) {
		lookup := []json.RawMessage{json.RawMessage(profileTombstoneJSON), json.RawMessage(`"junk"`)}
		for i := 0; i < 12; i++ {
			lookup = append(lookup, json.RawMessage(profileVectorJSON))
		}
		provider := &mockProfileLookupProvider{lookup: lookup}
		rec := get(profileOnlyServer(t, provider), profileTestDomain, "/v1/bsvalias/handle/Alice@example.com")
		require.Equal(t, http.StatusOK, rec.Code)
		var certs []json.RawMessage
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &certs))
		require.Len(t, certs, paymail.ProfileLookupMaxResults)
		assert.JSONEq(t, profileVectorJSON, string(certs[0]))
		assert.Equal(t, []string{"alice"}, provider.queries)
		assert.Equal(t, []string{profileTestDomain}, provider.domains)
	})

	t.Run("answers an empty array without asking the provider for unsearchable queries", func(t *testing.T) {
		provider := &mockProfileLookupProvider{lookup: []json.RawMessage{json.RawMessage(profileVectorJSON)}}
		handler := profileOnlyServer(t, provider)
		for _, query := range []string{"a", "alice@example.org", "al%20ice"} {
			rec := get(handler, profileTestDomain, "/v1/bsvalias/handle/"+query)
			require.Equal(t, http.StatusOK, rec.Code, query)
			assert.JSONEq(t, "[]", rec.Body.String(), query)
		}
		assert.Empty(t, provider.queries)
	})

	t.Run("uses the single configured domain when reached through another host", func(t *testing.T) {
		provider := &mockProfileLookupProvider{lookup: []json.RawMessage{json.RawMessage(profileVectorJSON)}}
		rec := get(profileOnlyServer(t, provider), "paymail.example.net", "/v1/bsvalias/handle/al")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, []string{profileTestDomain}, provider.domains)
	})

	t.Run("strips a port from the request host", func(t *testing.T) {
		provider := &mockProfileLookupProvider{lookup: []json.RawMessage{json.RawMessage(profileVectorJSON)}}
		locator := new(PaymailServiceLocator)
		locator.RegisterProfileLookupService(provider)
		config, err := NewConfig(locator, WithDomain(profileTestDomain), WithDomain("example.org"),
			WithoutGenericCapabilities(), WithProfileLookupCapabilities())
		require.NoError(t, err)
		rec := get(Handlers(config), profileTestDomain+":3000", "/v1/bsvalias/handle/al")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotEqual(t, "[]", rec.Body.String())
		assert.Equal(t, []string{profileTestDomain}, provider.domains)

		assert.Equal(t, http.StatusBadRequest, get(Handlers(config), "unknown.example", "/v1/bsvalias/handle/al").Code)
	})

	t.Run("reports provider errors", func(t *testing.T) {
		provider := &mockProfileLookupProvider{err: errRegistryDown}
		rec := get(profileOnlyServer(t, provider), profileTestDomain, "/v1/bsvalias/handle/al")
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestProfileLookup_Reverse(t *testing.T) {
	provider := &mockProfileLookupProvider{reverse: json.RawMessage(profileVectorJSON)}
	handler := profileOnlyServer(t, provider)

	t.Run("serves the certificate of the key", func(t *testing.T) {
		rec := get(handler, profileTestDomain, "/v1/bsvalias/identityKey/"+strings.ToUpper(profileVectorKey))
		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, profileVectorJSON, rec.Body.String())
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	})

	t.Run("answers 404 for an unknown key and 400 for a malformed one", func(t *testing.T) {
		rec := get(handler, profileTestDomain, "/v1/bsvalias/identityKey/03"+strings.Repeat("1", 64))
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "error-profile-not-found")

		rec = get(handler, profileTestDomain, "/v1/bsvalias/identityKey/nope")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("refuses to serve an unservable certificate", func(t *testing.T) {
		tombstoned := profileOnlyServer(t, &mockProfileLookupProvider{reverse: json.RawMessage(profileTombstoneJSON)})
		rec := get(tombstoned, profileTestDomain, "/v1/bsvalias/identityKey/"+profileVectorKey)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "error-profile-unservable")
	})
}

func TestProfileLookup_Configuration(t *testing.T) {
	t.Run("legacy capabilities keep requiring a paymail service", func(t *testing.T) {
		locator := new(PaymailServiceLocator)
		locator.RegisterProfileLookupService(&mockProfileLookupProvider{})
		assert.Panics(t, func() {
			_, _ = NewConfig(locator, WithDomain(profileTestDomain), WithProfileLookupCapabilities())
		})
	})

	t.Run("profile lookup requires its provider", func(t *testing.T) {
		locator := new(PaymailServiceLocator)
		locator.RegisterPaymailService(new(mockServiceProvider))
		assert.Panics(t, func() {
			_, _ = NewConfig(locator, WithDomain(profileTestDomain), WithProfileLookupCapabilities())
		})
	})

	t.Run("combines with the generic capabilities", func(t *testing.T) {
		locator := new(PaymailServiceLocator)
		locator.RegisterPaymailService(new(mockServiceProvider))
		locator.RegisterProfileLookupService(&mockProfileLookupProvider{})
		config, err := NewConfig(locator, WithDomain(profileTestDomain), WithProfileLookupCapabilities())
		require.NoError(t, err)
		payload, err := config.EnrichCapabilities(profileTestDomain)
		require.NoError(t, err)
		assert.Contains(t, payload.Capabilities, paymail.BRFCPki)
		assert.Contains(t, payload.Capabilities, paymail.BRFCSenderValidation)
		assert.Contains(t, payload.Capabilities, paymail.BRFCProfileLookup)
	})

	t.Run("a host with no capabilities is refused", func(t *testing.T) {
		_, err := NewConfig(new(PaymailServiceLocator), WithDomain(profileTestDomain), WithoutGenericCapabilities())
		require.Error(t, err)
	})
}
