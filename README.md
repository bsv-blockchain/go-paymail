# go-paymail
> Paymail client & server library for Golang

## Table of Contents
- [go-paymail](#go-paymail)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
  - [Documentation](#documentation)
    - [Features](#features)
  - [Examples \& Tests](#examples--tests)
  - [Benchmarks](#benchmarks)
  - [Code Standards](#code-standards)
  - [Usage](#usage)

<br/>

## Installation

**go-paymail** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).
```shell script
go get -u github.com/bsv-blockchain/go-paymail
```

<br/>

## Documentation
View the generated [documentation](https://pkg.go.dev/github.com/bsv-blockchain/go-paymail)

[![GoDoc](https://godoc.org/github.com/bsv-blockchain/go-paymail?status.svg&style=flat&v=4)](https://pkg.go.dev/github.com/bsv-blockchain/go-paymail)

### Features
- [Paymail Client](client.go) (outgoing requests to other providers)
    - Use your own custom [Resty HTTP client](https://github.com/go-resty/resty)
    - Customize the [client options](client.go)
    - Use your own custom [net.Resolver](srv_test.go)
    - Full network support: [`mainnet`, `testnet`, `STN`](networks.go)
    - [Get & Validate SRV records](srv.go)
    - [Check SSL Certificates](ssl.go)
    - [Check & Validate DNSSEC](dns_sec.go)
    - [Generate, Validate & Load Additional BRFC Specifications](brfc.go)
    - [Fetch, Get and Has Capabilities](capabilities.go)
    - [Get Public Key Information - PKI](pki.go)
    - [Basic Address Resolution](resolve_address.go)
    - [Verify PubKey & Handle](verify_pubkey.go)
    - [Get Public Profile](public_profile.go)
    - [P2P Payment Destination](p2p_payment_destination.go)
    - [P2P Send Transaction](p2p_send_transaction.go)
- [Paymail Server](server) (basic example for hosting your own paymail server)
    - [Example Showing Capabilities](server/capabilities.go) 
    - [Example Showing PKI](server/pki.go)
    - [Example Verifying a PubKey](server/verify.go)
    - [Example Address Resolution](server/resolve_address.go)
    - [Example Getting a P2P Payment Destination](server/p2p_payment_destination.go)
    - [Example Receiving a P2P Transaction](server/p2p_receive_transaction.go)
- [Paymail Utilities](utilities.go) (handy methods)
    - [Sanitize & Validate Paymail Addresses](utilities.go)
    - [Sign & Verify Sender Request](sender_request.go)
    
<details>
<summary><strong><code>Package Dependencies</code></strong></summary>
<br/>

Client Packages:
- [BitcoinSchema/go-bitcoin](https://github.com/BitcoinSchema/go-bitcoin)
- [go-resty/resty](https://github.com/go-resty/resty/v2)
- [jarcoal/httpmock](https://github.com/jarcoal/httpmock)
- [libsv/go-bk](https://github.com/libsv/go-bk)
- [libsv/go-bt](https://github.com/libsv/go-bt)
- [miekg/dns](https://github.com/miekg/dns)

Server Packages:
- [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)
</details>

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

[goreleaser](https://github.com/goreleaser/goreleaser) for easy binary or library deployment to GitHub and can be installed via: `brew install goreleaser`.

The [.goreleaser.yml](.goreleaser.yml) file is used to configure [goreleaser](https://github.com/goreleaser/goreleaser).

Use `make release-snap` to create a snapshot version of the release, and finally `make release` to ship to production.
</details>

<details>
<summary><strong><code>Makefile Commands</code></strong></summary>
<br/>

View all `makefile` commands
```shell script
make help
```

List of all current commands:
```text
all                   Runs multiple commands
clean                 Remove previous builds and any test cache data
clean-mods            Remove all the Go mod cache
coverage              Shows the test coverage
diff                  Show the git diff
generate              Runs the go generate command in the base of the repo
godocs                Sync the latest tag with GoDocs
help                  Show this help message
install               Install the application
install-go            Install the application (Using Native Go)
install-releaser      Install the GoReleaser application
lint                  Run the golangci-lint application (install if not found)
release               Full production release (creates release in GitHub)
release               Runs common.release then runs godocs
release-snap          Test the full release (build binaries)
release-test          Full production test release (everything except deploy)
replace-version       Replaces the version in HTML/JS (pre-deploy)
tag                   Generate a new tag and push (tag version=0.0.0)
tag-remove            Remove a tag if found (tag-remove version=0.0.0)
tag-update            Update an existing tag to current commit (tag-update version=0.0.0)
test                  Runs lint and ALL tests
test-ci               Runs all tests via CI (exports coverage)
test-ci-no-race       Runs all tests via CI (no race) (exports coverage)
test-ci-short         Runs unit tests via CI (exports coverage)
test-no-lint          Runs just tests
test-short            Runs vet, lint and tests (excludes integration tests)
test-unit             Runs tests and outputs coverage
uninstall             Uninstall the application (and remove files)
update-linter         Update the golangci-lint package (macOS only)
vet                   Run the Go vet application
```
</details>

<br/>

## Examples & Tests
All unit tests and [examples](examples) run via [GitHub Actions](https://github.com/bsv-blockchain/go-paymail/actions) and
uses [Go version 1.21](https://golang.org/doc/go1.21). View the [configuration file](.github/workflows/run-tests.yml).

Run all tests (including integration tests)
```shell script
make test
```

Run tests (excluding integration tests)
```shell script
make test-short
```

<br/>

## Benchmarks
Run the Go benchmarks:
```shell script
make bench
```

<br/>

## Code Standards
Read more about this Go project's [code standards](.github/CODE_STANDARDS.md).

<br/>

## Usage
Checkout all the [client examples](examples/client) or [server examples](examples/server)!
