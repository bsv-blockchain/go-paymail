package paymail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bsv-blockchain/go-sdk/auth/certificates"
)

// BRC-235 public profile lookup: handle@domain to identity key and back, answered
// with self-signed BRC-52 profile certificates.
//
// Specs: https://github.com/bsv-blockchain/BRCs/blob/master/payments/0235.md

const (
	// ProfileCertificateType is base64(SHA-256("public profile lookup")).
	ProfileCertificateType = "SbatVXXssDW3AO0J9bxIljkHGbPBGCVAXg94gXFf0cE="
	// ProfileRevocationOutpoint is the zero outpoint: profiles are replaced or tombstoned, never revoked on chain.
	ProfileRevocationOutpoint = "0000000000000000000000000000000000000000000000000000000000000000.0"
	// ProfileLookupMaxResults is the most results a host returns and a client processes.
	ProfileLookupMaxResults = 10
	// ProfileLookupMinQueryLength is the shortest handle query a host searches for.
	ProfileLookupMinQueryLength = 2
	// ProfileLookupMaxQueryLength is the longest handle query a host searches for.
	ProfileLookupMaxQueryLength = 32
	// ProfileQueryTemplate is the placeholder in the lookup capability URL.
	ProfileQueryTemplate = "{query}"
	// ProfileIdentityKeyTemplate is the placeholder in the reverse lookup capability URL.
	ProfileIdentityKeyTemplate = "{pubkey}"
	// ProfileMaxCertificateBytes bounds the serialized certificate.
	ProfileMaxCertificateBytes = 16384

	profileMaxFields        = 32
	profileMaxFieldName     = 50 // exclusive
	profileMaxFieldValue    = 1024
	profileSerialNumberSize = 32
)

var (
	// ErrProfileLookupInvalidURL is returned when the capability URL is not an HTTPS template
	ErrProfileLookupInvalidURL = errors.New("invalid profile lookup url")
	// ErrProfileLookupInvalidKey is returned when an identity key is not a compressed public key
	ErrProfileLookupInvalidKey = errors.New("invalid identity key")
	// ErrProfileLookupBadResponse is returned when the paymail provider returns a bad response
	ErrProfileLookupBadResponse = errors.New("bad response from paymail provider")
	// ErrProfileCertificateInvalid is returned when a certificate fails a BRC-235 rule
	ErrProfileCertificateInvalid = errors.New("invalid profile certificate")

	profileHandleRegExp      = regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]{0,30}[a-z0-9])?$`)
	profileQueryRegExp       = regexp.MustCompile(`^[a-z0-9._-]+$`)
	profileIdentityKeyRegExp = regexp.MustCompile(`^0[23][0-9a-f]{64}$`)
)

// ProfileCertificate is the seven-member BRC-52 JSON form of a profile certificate
type ProfileCertificate struct {
	Type               string            `json:"type"`
	SerialNumber       string            `json:"serialNumber"`
	Subject            string            `json:"subject"`
	Certifier          string            `json:"certifier"`
	RevocationOutpoint string            `json:"revocationOutpoint"`
	Fields             map[string]string `json:"fields"`
	Signature          string            `json:"signature"`
}

// VerifiedProfile is a profile certificate that passed every BRC-235 client verification rule
type VerifiedProfile struct {
	IdentityKey string             `json:"identityKey"` // The holder's identity key: the BRC-29 payment recipient
	Paymail     string             `json:"paymail"`
	Handle      string             `json:"handle"`
	Domain      string             `json:"domain"`
	DisplayName string             `json:"displayName,omitempty"`
	IssuedAt    string             `json:"issuedAt"`
	Certificate ProfileCertificate `json:"certificate"`
}

// ProfileExpectation is what a verified certificate must also match
type ProfileExpectation struct {
	Domain      string // The domain that was resolved; fields.paymail must be at it
	Paymail     string // Optional: a complete paymail the user typed; fields.paymail must equal it
	IdentityKey string // Optional: the key a reverse lookup asked for; subject must equal it
}

// ProfileLookupResponse is the result returned from LookupProfiles()
type ProfileLookupResponse struct {
	StandardResponse

	Profiles []VerifiedProfile `json:"profiles"` // Verified suggestions in the host's order
}

// ProfileReverseLookupResponse is the result returned from ReverseLookupProfile()
type ProfileReverseLookupResponse struct {
	StandardResponse

	Profile *VerifiedProfile `json:"profile,omitempty"` // Nil when the key holds no handle at the host
}

// IsValidProfileHandle reports whether handle is a BRC-235 handle:
// 1-32 of a-z 0-9 . _ -, starting and ending with a letter or digit
func IsValidProfileHandle(handle string) bool {
	return profileHandleRegExp.MatchString(handle)
}

// IsProfileIdentityKey reports whether value is a compressed public key in lowercase hex
func IsProfileIdentityKey(value string) bool {
	return profileIdentityKeyRegExp.MatchString(value)
}

// NormalizeProfileLookupQuery returns the handle text a host searches for. It lowercases
// and trims the query and strips an @domain suffix when the suffix is domain or a prefix
// of it (a domain still being typed). ok is false when the query can match nothing at domain.
func NormalizeProfileLookupQuery(query, domain string) (normalized string, ok bool) {
	normalized = strings.ToLower(strings.TrimSpace(query))
	if local, suffix, found := strings.Cut(normalized, "@"); found {
		if strings.Contains(suffix, "@") || !strings.HasPrefix(strings.ToLower(domain), suffix) {
			return "", false
		}
		normalized = local
	}
	if len(normalized) < ProfileLookupMinQueryLength || len(normalized) > ProfileLookupMaxQueryLength ||
		!profileQueryRegExp.MatchString(normalized) || normalized == ".." {
		return "", false
	}
	return normalized, true
}

func invalidProfile(reason string) error {
	return fmt.Errorf("%w: %s", ErrProfileCertificateInvalid, reason)
}

// CheckProfileCertificate parses raw and checks every BRC-235 rule except the
// signature, so a host can screen what it serves cheaply. Tombstones fail.
func CheckProfileCertificate(raw []byte, domain string) (*ProfileCertificate, error) {
	if len(raw) > ProfileMaxCertificateBytes {
		return nil, invalidProfile("larger than the size limit")
	}
	var cert ProfileCertificate
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&cert); err != nil {
		return nil, invalidProfile("not a certificate object")
	}
	if cert.Fields == nil {
		return nil, invalidProfile("fields are required")
	}
	if cert.Type != ProfileCertificateType {
		return nil, invalidProfile("wrong certificate type")
	}
	if serial, err := base64.StdEncoding.DecodeString(cert.SerialNumber); err != nil ||
		len(serial) != profileSerialNumberSize || base64.StdEncoding.EncodeToString(serial) != cert.SerialNumber {
		return nil, invalidProfile("serialNumber must be 32 bytes in canonical base64")
	}
	subject := strings.ToLower(cert.Subject)
	if !IsProfileIdentityKey(subject) || subject != strings.ToLower(cert.Certifier) {
		return nil, invalidProfile("subject must be a compressed key equal to certifier")
	}
	if cert.RevocationOutpoint != ProfileRevocationOutpoint {
		return nil, invalidProfile("revocationOutpoint must be the zero outpoint")
	}
	if len(cert.Fields) > profileMaxFields {
		return nil, invalidProfile("too many fields")
	}
	for name, value := range cert.Fields {
		if len(name) >= profileMaxFieldName || len(value) > profileMaxFieldValue || !utf8.ValidString(value) {
			return nil, invalidProfile("field " + name + " exceeds the limits")
		}
	}
	if cert.Fields["released"] == "true" {
		return nil, invalidProfile("released")
	}
	if _, err := time.Parse(time.RFC3339Nano, cert.Fields["issuedAt"]); err != nil {
		return nil, invalidProfile("issuedAt must be RFC 3339")
	}
	handle, paymailDomain, found := strings.Cut(cert.Fields["paymail"], "@")
	if !found || !IsValidProfileHandle(handle) || domain == "" ||
		paymailDomain != strings.ToLower(strings.TrimSpace(domain)) {
		return nil, invalidProfile("paymail must be a lowercase handle at " + domain)
	}
	return &cert, nil
}

// VerifyProfileCertificate applies every BRC-235 client verification rule, signature
// included, and returns the verified profile
func VerifyProfileCertificate(ctx context.Context, raw []byte, expect ProfileExpectation) (*VerifiedProfile, error) {
	cert, err := CheckProfileCertificate(raw, expect.Domain)
	if err != nil {
		return nil, err
	}
	paymail := cert.Fields["paymail"]
	if expect.Paymail != "" && paymail != strings.ToLower(strings.TrimSpace(expect.Paymail)) {
		return nil, invalidProfile("paymail is not the one requested")
	}
	if expect.IdentityKey != "" && !strings.EqualFold(cert.Subject, expect.IdentityKey) {
		return nil, invalidProfile("subject is not the key requested")
	}
	var sdkCert certificates.Certificate
	if err = json.Unmarshal(raw, &sdkCert); err != nil {
		return nil, invalidProfile("not a BRC-52 certificate")
	}
	if err = sdkCert.Verify(ctx); err != nil {
		return nil, invalidProfile("signature does not verify")
	}
	handle, domain, _ := strings.Cut(paymail, "@")
	return &VerifiedProfile{
		IdentityKey: strings.ToLower(cert.Subject),
		Paymail:     paymail,
		Handle:      handle,
		Domain:      domain,
		DisplayName: cert.Fields["displayName"],
		IssuedAt:    cert.Fields["issuedAt"],
		Certificate: *cert,
	}, nil
}

func profileTemplateURL(template, placeholder string) error {
	if !strings.HasPrefix(template, "https://") || !strings.Contains(template, placeholder) {
		return fmt.Errorf("url %s: %w", template, ErrProfileLookupInvalidURL)
	}
	return nil
}

func profileServerError(resp StandardResponse) error {
	serverError := &ServerError{}
	if err := json.Unmarshal(resp.Body, serverError); err != nil {
		return fmt.Errorf("code %d: %w", resp.StatusCode, ErrProfileLookupBadResponse)
	}
	return fmt.Errorf("code %d, message: %s: %w", resp.StatusCode, serverError.Message, ErrProfileLookupBadResponse)
}

// LookupProfiles searches domain for handles matching query (a partial handle, a full
// handle, or a full paymail) through the BRC-235 public profile lookup capability.
// Every certificate is verified and any that fails is dropped. Results are suggestions:
// show each full paymail and let the user choose; never pick one by position. When
// query is a complete handle@domain at domain, only that exact paymail is returned.
//
// lookupURL is the capability template from GetCapabilities(), containing {query}.
func (c *Client) LookupProfiles(lookupURL, domain, query string) (response *ProfileLookupResponse, err error) {
	if err = profileTemplateURL(lookupURL, ProfileQueryTemplate); err != nil {
		return nil, err
	}
	domain = strings.ToLower(strings.TrimSpace(domain))
	handleQuery, ok := NormalizeProfileLookupQuery(query, domain)
	if !ok {
		return &ProfileLookupResponse{Profiles: []VerifiedProfile{}}, nil
	}
	exact := ""
	if typed := strings.ToLower(strings.TrimSpace(query)); typed == handleQuery+"@"+domain {
		exact = typed
	}

	var resp StandardResponse
	if resp, err = c.getRequest(strings.ReplaceAll(lookupURL, ProfileQueryTemplate, url.PathEscape(handleQuery))); err != nil {
		return nil, err
	}
	response = &ProfileLookupResponse{StandardResponse: resp, Profiles: []VerifiedProfile{}}
	if resp.StatusCode != http.StatusOK {
		return response, profileServerError(resp)
	}
	var certs []json.RawMessage
	if err = json.Unmarshal(resp.Body, &certs); err != nil {
		return response, fmt.Errorf("profile lookup must return an array: %w", ErrProfileLookupBadResponse)
	}
	if len(certs) > ProfileLookupMaxResults {
		certs = certs[:ProfileLookupMaxResults]
	}
	ctx := context.Background()
	for _, raw := range certs {
		profile, verifyErr := VerifyProfileCertificate(ctx, raw, ProfileExpectation{Domain: domain, Paymail: exact})
		if verifyErr == nil {
			response.Profiles = append(response.Profiles, *profile)
		}
	}
	return response, nil
}

// ReverseLookupProfile returns the verified profile identityKey holds at domain through
// the BRC-235 public profile reverse lookup capability. Profile is nil when the key holds
// no handle there or the host's certificate fails verification.
//
// reverseURL is the capability template from GetCapabilities(), containing {pubkey}.
func (c *Client) ReverseLookupProfile(reverseURL, domain, identityKey string) (response *ProfileReverseLookupResponse, err error) {
	if err = profileTemplateURL(reverseURL, ProfileIdentityKeyTemplate); err != nil {
		return nil, err
	}
	identityKey = strings.ToLower(identityKey)
	if !IsProfileIdentityKey(identityKey) {
		return nil, ErrProfileLookupInvalidKey
	}

	var resp StandardResponse
	if resp, err = c.getRequest(strings.ReplaceAll(reverseURL, ProfileIdentityKeyTemplate, identityKey)); err != nil {
		return nil, err
	}
	response = &ProfileReverseLookupResponse{StandardResponse: resp}
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return response, nil
	default:
		return response, profileServerError(resp)
	}
	profile, verifyErr := VerifyProfileCertificate(context.Background(), resp.Body, ProfileExpectation{
		Domain: strings.ToLower(strings.TrimSpace(domain)), IdentityKey: identityKey,
	})
	if verifyErr == nil {
		response.Profile = profile
	}
	return response, nil
}
