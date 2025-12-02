package paymail

import "testing"

// FuzzSanitizeEmail tests the SanitizeEmail function with arbitrary string inputs.
// It verifies that the function does not panic on any input.
func FuzzSanitizeEmail(f *testing.F) {
	// Seed corpus with representative examples
	seeds := []string{
		"test@domain.com",
		"TEST@DOMAIN.COM",
		"mailto:user@example.com",
		"MAILTO:USER@EXAMPLE.COM",
		"  spaced@email.com  ",
		"special!#$%&'*+-/=?^_`{|}~chars@domain.com",
		"",
		"noatsign",
		"multiple@@signs@domain.com",
		"unicode\u6D4B\u8BD5@domain.com",
		"tab\t@domain.com",
		"newline\n@domain.com",
		"null\x00@domain.com",
		"<script>alert('xss')</script>@evil.com",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Function should not panic on any input
		_ = SanitizeEmail(input)
	})
}
