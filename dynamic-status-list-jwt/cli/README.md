# Dynamic Status List CLI <!-- omit in toc -->

A simple command-line interface for experiencing the Dynamic Status List (DSL)
for JWT.

> [!IMPORTANT]
> This is an early draft. Not all functions are implemented.

## Table of Contents <!-- omit in toc -->

- [Download and Build](#download-and-build)
  - [Prerequisites](#prerequisites)
  - [Installation Steps](#installation-steps)
- [Quick Start](#quick-start)
  - [Issue a Mock JWT](#issue-a-mock-jwt)
  - [Create a Status List Entry](#create-a-status-list-entry)
  - [Compute the Revocation Identifier](#compute-the-revocation-identifier)
  - [Recompute the Dynamic Status List](#recompute-the-dynamic-status-list)
  - [Revoke a JWT](#revoke-a-jwt)
  - [Verify JWT Revocation Status](#verify-jwt-revocation-status)
- [Advanced features](#advanced-features)
  - [Crate detached status list metadata](#crate-detached-status-list-metadata)
- [Roadmap](#roadmap)

## Download and Build

To get started, install the CLI tool and test its capabilities. For an in-depth understanding, refer to the [design document](dsl-jwt.md).

### Prerequisites

Ensure you have `Go` installed (version 1.18 or later). If not, install it [here](https://go.dev/doc/install).

### Installation Steps

Clone the repository and navigate to the CLI directory:

```bash
git clone github.com/alenhorvat/dsl-jwt-profile
cd cli
```

Install dependencies and build the CLI tool:

```bash
go mod tidy
go build
```

You're now ready to use the DSL CLI. To verify the installation, run:

```bash
./dsl --help
```

This should display the available commands and options.

## Quick Start

This guide demonstrates how the Dynamic Status List works. Non-essential JWT claims are omitted for clarity.

### Issue a Mock JWT

Before revoking a JWT, you must issue one. Run:

```bash
./dsl issue
```

This generates a `mock-jwt.json` file, containing a signed JWT with a `jti` claim and an `sdp` (status list distribution point) claim.

### Create a Status List Entry

To make the JWT revocable, create a new status list entry:

```bash
./dsl new -i mock-jwt.json
```

This updates `dsl.json` with the latest DSL entries (`dsl_jwt` claim), which maps JWT identifiers (`jti`) to their status within the dynamic status list. A full list of registered JWT `jti` claims is stored in `dsl-map.json`.

### Compute the Revocation Identifier

As a holder, you can compute the revocation identifier (found in `dsl.json#/dsl_jwt`) using:

```bash
./dsl wallet -i mock-jwt.json -t 1739179906
```

The computed status list identifier (`sid`) is printed and saved in `holder_status-list-identifier.json`.

If no timestamp is provided, the identifier is computed at the time of execution. Holders can precompute identifiers for any past or future time.

### Recompute the Dynamic Status List

Recompute the dynamic status list using:

```bash
./dsl recompute
```

This updates `dsl.json`. To compute the status list at a specific UNIX timestamp `t`, use:

```bash
./dsl recompute -t 1739139573
```

### Revoke a JWT

To revoke a JWT, specify the `jti` identifier, which can be found in the `mock-jwt.json` file under the `jti` claim:

```bash
./dsl revoke --jti 123
```

_Note: Future versions (v2) will support revocation time and reason._

### Verify JWT Revocation Status

To verify a holder’s proof, run:

```bash
./dsl verify -s dsl.json -p holder_status-list-identifier.json
```

This checks whether the provided identifier is valid or revoked based on the status list.

## Advanced features

### Crate detached status list metadata

In case you're unable to modify the JWT issuance, you can create detached status list metadata by calling

```bash
./dsl new -i mock-jwt.json --detached
```

The method will store the detached JWT status list metadata in the `mock-jwt.json` with the following two claims in the payload

```json
{
  "sdb": "http://localhost:PORT/sdb/1",
  "sub": "e28fceae96a7e84079c5efe922e03264"
}
```

where `sbp` is the status distribution point and `sub` MUST match the `jti` value of the JWT.

Note: we assume `jti` is defined in the JWT. If even that's missing, one could use digest of the JWT as identifier.

## Roadmap

- Support for revocation metadata/extensions: Encrypted revocation metadata
- Seed is a simple shared secret, strengthen it with ARKG
