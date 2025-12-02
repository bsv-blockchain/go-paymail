package paymail

import "testing"

// FuzzIsValidEmail tests the IsValidEmail function with arbitrary string inputs.
// It verifies that the function does not panic on any input.
func FuzzIsValidEmail(f *testing.F) {
	// Seed corpus with representative examples
	seeds := []string{
		"test@domain.com",
		"user@example.org",
		"invalid-email",
		"@nodomain.com",
		"noat.com",
		"",
		"a@b.co",
		"very.long.email.address+tag@subdomain.domain.co.uk",
		"test@.com",
		"test@domain.",
		"<script>@evil.com",
		"test\x00@domain.com",
		"test@domain.com\n",
		"test\u6D4B\u8BD5@domain.com",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Function should not panic on any input
		_ = IsValidEmail(input)
	})
}

// FuzzIsValidDNSName tests the IsValidDNSName function with arbitrary string inputs.
// It verifies that the function does not panic on any input.
func FuzzIsValidDNSName(f *testing.F) {
	// Seed corpus with representative examples
	seeds := []string{
		"example.com",
		"sub.domain.org",
		"localhost",
		"",
		"192.168.1.1",
		"::1",
		"-invalid.com",
		"invalid-.com",
		"a.b.c.d.e.f.g",
		"xn--nxasmq5a.com",
		"test_domain.com",
		"UPPERCASE.COM",
		"mixed.Case.Domain",
		string(make([]byte, 300)),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Function should not panic on any input
		_ = IsValidDNSName(input)
	})
}

// FuzzIsValidIP tests the IsValidIP function with arbitrary string inputs.
// It verifies that the function does not panic on any input.
func FuzzIsValidIP(f *testing.F) {
	// Seed corpus with representative examples
	seeds := []string{
		"192.168.1.1",
		"10.0.0.1",
		"255.255.255.255",
		"0.0.0.0",
		"::1",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"fe80::1",
		"invalid",
		"",
		"256.256.256.256",
		"192.168.1",
		"192.168.1.1.1",
		"abc.def.ghi.jkl",
		"192.168.1.1:8080",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Function should not panic on any input
		_ = IsValidIP(input)
	})
}
