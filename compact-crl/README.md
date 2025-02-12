# CCRL CLI Tool Documentation

## Overview

`ccrl` is a command-line tool for managing and testing Compact Certificate Revocation Lists (CRLs) with different revocation list profiles. It includes benchmarking tools and CRL creation utilities, particularly focusing on the Bit Status List extension.

Supported profiles

- CRL Bit String Status extension
- Next: Dynamic CRL (following the design of Dynamic Status List)

## Installation

Ensure you have Go installed, then build and install the tool:

```sh
git clone <repository_url>
cd ccrl
go mod tidy
go build
```

## Create signing certificates

Before you begin, you need to create signing certificates. You can do so by running the script

```bash
bash create-certs.sh
```

## Usage

Run the tool with:

```sh
./ccrl <command> [flags]
```

### Commands

#### `bench`

Runs a benchmark test for CRLs with the Bit Status List extension.

```sh
./ccrl bench [flags]
```

Flags

- `-b, --bit-string-list` → Run the Bit String CRL benchmark.
- `-l, --lower-limit <int>` → Smallest bit string list size (10^l)
- `-u, --upper-limit <int>` → Largest bit string list size (10^u)
- `-c, --crt <path>` → Path to the CRL signing certificate (default: `certs/rootCA.crt`).
- `-k, --key <path>` → Path to the CRL signing key (default: `certs/rootCA.key`).

Example:

```sh
./ccrl bench -b
```

#### `bsl`

Generates a CRL with a Bit String CRL extension from a hex-encoded byte array. Bit String list MUST be represented as a byte array.

```sh
./ccrl bsl [flags]
```

Flags:

- `-b, --bit-string-bytes <hex>` → Hex-encoded bit string list.
- `-p, --bit-string-path <path>` → Path to a file containing the hex-encoded bit string list.
- `-z, --compress` → Compress the byte array using zlib.
- `-c, --crt <path>` → Path to the CRL signing certificate (default: `certs/rootCA.crt`).
- `-k, --key <path>` → Path to the CRL signing key (default: `certs/rootCA.key`).

```sh
./ccrl bsl -b "a1b2c3d4"
```

## License

This project is open-source. See the LICENSE file of the repository for details.

## Known limitations

- There is an official recommendation: for backward compatibility with some
deployed implementations, the serial number encoded value should fit in 20 bytes
