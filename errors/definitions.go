package errors

// How the Codes are generated?
// 1. "error" - like mandatory prefix for all error codes
// 2. (optional) {error group} - e.g. "configuration", "capabilities", "spv"
// 3. (optional) {subject} - name of model (with or without specific field) or some noun e.g. "body", "auth-header", "transaction", "paymail-address"
// 4. (optional) {reason} - what happened, e.g. "not-found", "missing", "invalid"

// CONFIG ERRORS
var (
	// ErrDomainMissing is the error for missing domain
	ErrDomainMissing = SPVError{Message: "domain is missing", StatusCode: 500, Code: "error-configuration-domain-missing"}

	// ErrPortMissing is when the port is not found
	ErrPortMissing = SPVError{Message: "missing a port", StatusCode: 500, Code: "error-configuration-port-missing"}

	// ErrServiceNameMissing is when the service name is not found
	ErrServiceNameMissing = SPVError{Message: "missing service name", StatusCode: 500, Code: "error-configuration-service-name-missing"}

	// ErrCapabilitiesMissing is when the capabilities struct is nil or not set
	ErrCapabilitiesMissing = SPVError{Message: "missing capabilities struct", StatusCode: 500, Code: "error-configuration-capabilities-missing"}

	// ErrBsvAliasMissing is when the bsv alias version is missing
	ErrBsvAliasMissing = SPVError{Message: "missing bsv alias version", StatusCode: 500, Code: "error-configuration-bsv-alias-missing"}

	// ErrServiceProviderNil is the error for having a nil service provider
	ErrServiceProviderNil = SPVError{Message: "service provider is nil", StatusCode: 500, Code: "error-configuration-service-provider-nil"}
)

// CAPABILITY ERRORS
var (
	// ErrPrefixOrDomainMissing is when the prefix or domain is missing
	ErrPrefixOrDomainMissing = SPVError{Message: "prefix or domain is missing", StatusCode: 400, Code: "error-capabilities-prefix-or-domain-missing"}

	// ErrDomainUnknown is when the domain is not in the list of allowed domains
	ErrDomainUnknown = SPVError{Message: "paymail domain is unknown", StatusCode: 400, Code: "error-capabilities-domain-unknown"}

	// ErrCastingNestedCapabilities is when the nested capabilities cannot be cast
	ErrCastingNestedCapabilities = SPVError{Message: "failed to cast nested capabilities", StatusCode: 500, Code: "error-capabilities-nested-capabilities-failed-to-cast"}
)

// PARSING ERRORS
var (
	// ErrCannotBindRequest is when request body cannot be bind into struct
	ErrCannotBindRequest = SPVError{Message: "cannot bind request body", StatusCode: 400, Code: "error-bind-body-invalid"}

	// ErrProcessingHex is when error occurred during processing hex
	ErrProcessingHex = SPVError{Message: "cannot process hex", StatusCode: 400, Code: "error-processing-hex"}

	// ErrProcessingBEEF is when error occurred during processing beef
	ErrProcessingBEEF = SPVError{Message: "cannot process beef", StatusCode: 400, Code: "error-processing-beef"}
)

// PAYMAIL ERRORS
var (
	// ErrCouldNotFindPaymail is when could not find paymail
	ErrCouldNotFindPaymail = SPVError{Message: "invalid paymail", StatusCode: 400, Code: "error-paymail-not-found"}
)

// INVALID FIELD ERRORS
var (
	// ErrInvalidPaymail is when the paymail is invalid
	ErrInvalidPaymail = SPVError{Message: "invalid paymail", StatusCode: 400, Code: "error-paymail-invalid"}

	// ErrInvalidPubKey is when the pubkey is invalid
	ErrInvalidPubKey = SPVError{Message: "invalid pubkey", StatusCode: 400, Code: "error-pubkey-invalid"}

	// ErrInvalidSignature is when the signature is invalid
	ErrInvalidSignature = SPVError{Message: "invalid signature", StatusCode: 400, Code: "error-signature-invalid"}

	// ErrInvalidScript is when the script is invalid
	ErrInvalidScript = SPVError{Message: "invalid script", StatusCode: 400, Code: "error-script-invalid"}

	// ErrInvalidTimestamp is when the timestamp is invalid
	ErrInvalidTimestamp = SPVError{Message: "invalid timestamp", StatusCode: 400, Code: "error-timestamp-invalid"}

	// ErrInvalidSenderHandle is when the sender handle is invalid
	ErrInvalidSenderHandle = SPVError{Message: "invalid sender handle", StatusCode: 400, Code: "error-sender-handle-invalid"}
)

// MISSING FIELD ERRORS

var (
	// ErrMissingFieldReference is when the reference field is required but missing
	ErrMissingFieldReference = SPVError{Message: "missing required field: reference", StatusCode: 400, Code: "error-missing-field-reference"}

	// ErrMissingFieldHex is when the hex field is required but missing
	ErrMissingFieldHex = SPVError{Message: "missing required field: hex", StatusCode: 400, Code: "error-missing-field-hex"}

	// ErrMissingFieldBEEF is when the beef field is required but missing
	ErrMissingFieldBEEF = SPVError{Message: "missing required field: beef", StatusCode: 400, Code: "error-missing-field-beef"}

	// ErrMissingFieldSignature is when the signature field is required but missing
	ErrMissingFieldSignature = SPVError{Message: "missing required field: signature", StatusCode: 400, Code: "error-missing-field-signature"}

	// ErrMissingFieldPubKey is when the pubkey field is required but missing
	ErrMissingFieldPubKey = SPVError{Message: "missing required field: pubkey", StatusCode: 400, Code: "error-missing-field-pubkey"}

	// ErrMissingFieldSatoshis is when the satoshis field is required but missing
	ErrMissingFieldSatoshis = SPVError{Message: "missing required field: satoshis", StatusCode: 400, Code: "error-missing-field-satoshis"}
)

// EMPTY FIELDS ERRORS
var (
	// ErrSenderHandleEmpty is when the server handle is empty
	ErrSenderHandleEmpty = SPVError{Message: "empty sender handle", StatusCode: 400, Code: "error-sender-handle-empty"}

	// ErrDtEmpty is when the dt field is empty
	ErrDtEmpty = SPVError{Message: "empty dt", StatusCode: 400, Code: "error-dt-empty"}
)

// SPV ERRORS
var (
	// ErrNoOutputs is when there are no outputs
	ErrNoOutputs = SPVError{Message: "invalid output, no outputs", StatusCode: 417, Code: "error-spv-no-outputs"}

	// ErrNoInputs is when there are no inputs
	ErrNoInputs = SPVError{Message: "invalid input, no inputs", StatusCode: 417, Code: "error-spv-no-inputs"}

	// ErrInvalidParentTransactions is when the parent transactions are invalid
	ErrInvalidParentTransactions = SPVError{Message: "invalid parent transactions, no matching transactions for input", StatusCode: 417, Code: "error-spv-parent-tx-invalid"}

	// ErrLockTimeAndSequence is when the locktime and sequence are invalid
	ErrLockTimeAndSequence = SPVError{Message: "nLocktime is set and nSequence is not max, therefore this could be a non-final tx which is not currently supported", StatusCode: 417, Code: "error-spv-locktime-sequence-invalid"}

	// ErrOutputValueTooHigh is when the satoshis output is too high on a transaction
	ErrOutputValueTooHigh = SPVError{Message: "invalid input and output sum, outputs can not be larger than inputs", StatusCode: 417, Code: "error-spv-output-value-too-high"}

	// ErrBUMPAncestorNotPresent is when the input mined ancestor is not present in BUMPs
	ErrBUMPAncestorNotPresent = SPVError{Message: "invalid BUMP - input mined ancestor is not present in BUMPs", StatusCode: 417, Code: "error-spv-bump-ancestor-not-present"}

	// ErrBUMPCouldNotFindMinedParent is when the mined parent for input could not be found
	ErrBUMPCouldNotFindMinedParent = SPVError{Message: "invalid BUMP - cannot find mined parent for input", StatusCode: 417, Code: "error-spv-bump-mined-parent-not-found"}

	// ErrNoMatchingTransactionsForInput is when no matching transaction for input can be found
	ErrNoMatchingTransactionsForInput = SPVError{Message: "invalid parent transactions, no matching transactions for input", StatusCode: 417, Code: "error-spv-bump-ancestor-not-present"}

	// ErrSPVFailed is when the SPV returns an error
	ErrSPVFailed = SPVError{Message: "simplified payment verification has failed", StatusCode: 417, Code: "error-spv-failed"}
)
