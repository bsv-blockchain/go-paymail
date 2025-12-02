package paymail

import "testing"

// FuzzDecodeSignature tests the DecodeSignature function with arbitrary string inputs.
// It verifies that the function does not panic on any input and handles errors gracefully.
func FuzzDecodeSignature(f *testing.F) {
	// Seed corpus with representative examples
	seeds := []string{
		"SGVsbG8gV29ybGQh",                 // Valid base64: "Hello World!"
		"dGVzdA==",                         // Valid base64: "test"
		"YWJjZGVm",                         // Valid base64: "abcdef"
		"",                                 // Empty string
		"!!!invalid!!!",                    // Invalid base64
		"SGVsbG8gV29ybGQh==",               // Incorrect padding
		"SGVsbG8=V29ybGQh",                 // Padding in wrong place
		"    ",                             // Only whitespace
		"YWJj\nZGVm",                       // Base64 with newline
		"YWJj ZGVm",                        // Base64 with space
		"../../../etc/passwd",              // Path traversal attempt
		"<script>alert('xss')</script>",    // XSS attempt
		string(make([]byte, 1000)),         // Long string
		"\x00\x01\x02\x03",                 // Binary data
		"+++///===",                        // Valid base64 chars
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", // Repeated pattern
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Function should not panic on any input
		// Error returns are expected for invalid base64
		_, _ = DecodeSignature(input)
	})
}
