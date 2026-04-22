<div align="center">

# ✉️&nbsp;&nbsp;go-paymail

**Paymail toolkit for Go with full‑stack client and server support.**

<br/>

<a href="https://github.com/bsv-blockchain/go-paymail/releases"><img src="https://img.shields.io/github/release-pre/bsv-blockchain/go-paymail?include_prereleases&style=flat-square&logo=github&color=black" alt="Release"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/github/go-mod/go-version/bsv-blockchain/go-paymail?style=flat-square&logo=go&color=00ADD8" alt="Go Version"></a>
<a href="https://github.com/bsv-blockchain/go-paymail/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-OpenBSV-blue?style=flat-square" alt="License"></a>

<br/>

<table align="center" border="0">
  <tr>
    <td align="right">
       <code>CI / CD</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/bsv-blockchain/go-paymail/actions"><img src="https://img.shields.io/github/actions/workflow/status/bsv-blockchain/go-paymail/fortress.yml?branch=main&label=build&logo=github&style=flat-square" alt="Build"></a>
       <a href="https://github.com/bsv-blockchain/go-paymail/actions"><img src="https://img.shields.io/github/last-commit/bsv-blockchain/go-paymail?style=flat-square&logo=git&logoColor=white&label=last%20update" alt="Last Commit"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Quality</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://goreportcard.com/report/github.com/bsv-blockchain/go-paymail"><img src="https://goreportcard.com/badge/github.com/bsv-blockchain/go-paymail?style=flat-square" alt="Go Report"></a>
       <a href="https://codecov.io/gh/bsv-blockchain/go-paymail"><img src="https://codecov.io/gh/bsv-blockchain/go-paymail/branch/main/graph/badge.svg?style=flat-square" alt="Coverage"></a>
    </td>
  </tr>

  <tr>
    <td align="right">
       <code>Security</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://scorecard.dev/viewer/?uri=github.com/bsv-blockchain/go-paymail"><img src="https://api.scorecard.dev/projects/github.com/bsv-blockchain/go-paymail/badge?style=flat-square" alt="Scorecard"></a>
       <a href=".github/SECURITY.md"><img src="https://img.shields.io/badge/policy-active-success?style=flat-square&logo=security&logoColor=white" alt="Security"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Community</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/bsv-blockchain/go-paymail/graphs/contributors"><img src="https://img.shields.io/github/contributors/bsv-blockchain/go-paymail?style=flat-square&color=orange" alt="Contributors"></a>
       <a href="https://github.com/sponsors/bsv-blockchain"><img src="https://img.shields.io/badge/sponsor-BSV-181717.svg?logo=github&style=flat-square" alt="Sponsor"></a>
    </td>
  </tr>
</table>

</div>

<br/>
<br/>

<div align="center">

### <code>Project Navigation</code>

</div>

<table align="center">
  <tr>
    <td align="center" width="33%">
       📦&nbsp;<a href="#-installation"><code>Installation</code></a>
    </td>
    <td align="center" width="33%">
       🧪&nbsp;<a href="#-examples--tests"><code>Examples&nbsp;&&nbsp;Tests</code></a>
    </td>
    <td align="center" width="33%">
       📚&nbsp;<a href="#-documentation"><code>Documentation</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
       🤝&nbsp;<a href="#-contributing"><code>Contributing</code></a>
    </td>
    <td align="center">
       🛠️&nbsp;<a href="#-code-standards"><code>Code&nbsp;Standards</code></a>
    </td>
    <td align="center">
       ⚡&nbsp;<a href="#-benchmarks"><code>Benchmarks</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
       🤖&nbsp;<a href="#-ai-usage--assistant-guidelines"><code>AI&nbsp;Usage</code></a>
    </td>
    <td align="center">
       📝&nbsp;<a href="#-license"><code>License</code></a>
    </td>
    <td align="center">
       👥&nbsp;<a href="#-maintainers"><code>Maintainers</code></a>
    </td>
  </tr>
</table>
<br/>

## 📦 Installation

**go-paymail** requires a [supported release of Go](https://golang.org/doc/devel/release.html#policy).
```shell script
go get -u github.com/bsv-blockchain/go-paymail
```

<br/>

## 📚 Documentation
- **API Reference** – Dive into the godocs at [pkg.go.dev/github.com/bsv-blockchain/go-paymail](https://pkg.go.dev/github.com/bsv-blockchain/go-paymail)
- **Usage Examples** – Browse practical patterns either the [examples directory](examples) or the example tests
- **Benchmarks** – Check the latest numbers in the [benchmark results](#benchmark-results)

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

<br/>

<details>
<summary><strong><code>Development Build Commands</code></strong></summary>
<br/>

Get the [MAGE-X](https://github.com/mrz1836/mage-x) build tool for development:
```shell script
go install github.com/mrz1836/mage-x/cmd/magex@latest
```

View all build commands

```bash script
magex help
```

</details>

<details>
<summary><strong>Repository Features</strong></summary>
<br/>

This repository includes 25+ built-in features covering CI/CD, security, code quality, developer experience, and community tooling.

**[View the full Repository Features list →](.github/docs/repository-features.md)**

</details>

<details>
<summary><strong><code>Library Deployment</code></strong></summary>
<br/>

This project uses [goreleaser](https://github.com/goreleaser/goreleaser) for streamlined binary and library deployment to GitHub. To get started, install it via:

```bash
brew install goreleaser
```

The release process is defined in the [.goreleaser.yml](.goreleaser.yml) configuration file.


Then create and push a new Git tag using:

```bash
magex version:bump push=true bump=patch branch=main
```

This process ensures consistent, repeatable releases with properly versioned artifacts and citation metadata.

</details>

<details>
<summary><strong><code>Pre-commit Hooks</code></strong></summary>
<br/>

Set up the Go-Pre-commit System to run the same formatting, linting, and tests defined in [AGENTS.md](.github/AGENTS.md) before every commit:

```bash
go install github.com/mrz1836/go-pre-commit/cmd/go-pre-commit@latest
go-pre-commit install
```

The system is configured via the [environment configuration](.github/env/README.md) and provides 17x faster execution than traditional Python-based pre-commit hooks. See the [complete documentation](http://github.com/mrz1836/go-pre-commit) for details.

</details>

<details>
<summary><strong>GitHub Workflows</strong></summary>
<br/>

All workflows are driven by modular configuration in [`.github/env/`](.github/env/README.md) — no YAML editing required.

**[View all workflows and the control center →](.github/docs/workflows.md)**

</details>

<details>
<summary><strong><code>Updating Dependencies</code></strong></summary>
<br/>

To update all dependencies (Go modules, linters, and related tools), run:

```bash
magex deps:update
```

This command ensures all dependencies are brought up to date in a single step, including Go modules and any tools managed by [MAGE-X](https://github.com/mrz1836/mage-x). It is the recommended way to keep your development environment and CI in sync with the latest versions.

</details>

<br/>

## 🧪 Examples & Tests

All unit tests and [examples](examples) run via [GitHub Actions](https://github.com/bsv-blockchain/go-paymail/actions) and use [Go version 1.25.x](https://go.dev/doc/go1.25). View the [configuration file](.github/workflows/fortress.yml).

Run all tests (fast):

```bash script
magex test
```

Run all tests with race detector (slower):
```bash script
magex test:race
```

<br/>

## ⚡ Benchmarks

Run the Go benchmarks:

```bash script
magex bench
```

<br/>

### Benchmark Results

| Benchmark                     | Iterations | ns/op       | B/op    | allocs/op |
|-------------------------------|------------|-------------|---------|-----------|
| Validate SRV Record           | 4,516,506  | 26.49       | 24      | 1         |
| Capabilities Has              | 3,920,778  | 30.30       | 0       | 0         |
| Capabilities Get Bool         | 3,938,580  | 30.27       | 0       | 0         |
| Capabilities Get String       | 3,416,118  | 34.53       | 0       | 0         |
| Convert Handle                | 1,938,082  | 61.95       | 24      | 2         |
| Validate Timestamp            | 984,312    | 118.1       | 0       | 0         |
| BRFC Spec Generate            | 568,420    | 196.9       | 144     | 4         |
| BRFC Spec Validate            | 615,147    | 200.6       | 144     | 4         |
| Get SRV Record                | 405,182    | 286.4       | 144     | 8         |
| Validate Paymail              | 394,790    | 306.2       | 0       | 0         |
| Validate Domain               | 375,542    | 307.7       | 113     | 3         |
| Sanitize Paymail              | 157,504    | 703.0       | 317     | 9         |
| Validate And Sanitize Paymail | 121,256    | 1,001       | 346     | 10        |
| Get Public Profile            | 26,770     | 4,378       | 4,944   | 54        |
| Get PKI                       | 27,373     | 4,487       | 4,928   | 55        |
| Add Invite Request            | 24,337     | 4,772       | 5,091   | 54        |
| Resolve Address               | 22,069     | 5,352       | 4,922   | 64        |
| Get Outputs Template          | 21,830     | 5,367       | 4,971   | 63        |
| Get Capabilities              | 22,478     | 5,414       | 5,421   | 70        |
| Verify Pub Key                | 21,445     | 5,646       | 5,281   | 56        |
| Send P2P Transaction          | 20,475     | 5,657       | 5,726   | 62        |
| Get P2P Payment Destination   | 13,077     | 7,855       | 6,391   | 115       |
| Default Client Options        | 3,102      | 39,799      | 13,976  | 152       |
| New Client                    | 2,960      | 41,008      | 15,840  | 174       |
| Load BRFCs                    | 2,949      | 41,832      | 14,704  | 169       |
| Sender Request Verify         | 518        | 230,894     | 4,388   | 115       |
| Sender Request Sign           | 356        | 337,052     | 9,757   | 171       |
| Check DNSSEC                  | 1          | 266,799,625 | 55,840  | 539       |
| Check SSL                     | 1          | 388,776,125 | 856,032 | 8,190     |

> These benchmarks reflect fast, allocation-free lookups for most retrieval functions, ensuring optimal performance in production environments.
> Performance benchmarks for the core functions in this library, executed on an Apple M1 Max (ARM64).

<br/>

## 🛠️ Code Standards
Read more about this Go project's [code standards](.github/CODE_STANDARDS.md).

<br/>

## 🤖 AI Usage & Assistant Guidelines
Read the [AI Usage & Assistant Guidelines](.github/tech-conventions/ai-compliance.md) for details on how AI is used in this project and how to interact with AI assistants.

<br/>

## 👥 Maintainers
| [<img src="https://github.com/icellan.png" height="50" alt="Siggi" />](https://github.com/icellan) | [<img src="https://github.com/galt-tr.png" height="50" alt="Galt" />](https://github.com/galt-tr) | [<img src="https://github.com/mrz1836.png" height="50" alt="MrZ" />](https://github.com/mrz1836) |
|:--------------------------------------------------------------------------------------------------:|:-------------------------------------------------------------------------------------------------:|:------------------------------------------------------------------------------------------------:|
|                                [Siggi](https://github.com/icellan)                                 |                                [Dylan](https://github.com/galt-tr)                                 |                                [MrZ](https://github.com/mrz1836)                                 |

<br/>

## 🤝 Contributing
View the [contributing guidelines](.github/CONTRIBUTING.md) and please follow the [code of conduct](.github/CODE_OF_CONDUCT.md).

### How can I help?
All kinds of contributions are welcome :raised_hands:!
The most basic way to show your support is to star :star2: the project, or to raise issues :speech_balloon:.

[![Stars](https://img.shields.io/github/stars/bsv-blockchain/go-paymail?label=Please%20like%20us&style=social&v=1)](https://github.com/bsv-blockchain/go-paymail/stargazers)

<br/>

## 📝 License

[![License](https://img.shields.io/badge/license-OpenBSV-blue?style=flat&logo=springsecurity&logoColor=white)](LICENSE)

