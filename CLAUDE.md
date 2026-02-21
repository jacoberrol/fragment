# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Run
go run . <subcommand>

# Test
go test ./...

# Test a single package
go test ./internal/shamir/...

# Lint (requires golangci-lint)
golangci-lint run
```

## Architecture

`fragment` is a CLI tool (built with Cobra) for secret sharing: it encrypts files/directories and splits the encryption key into shares using Shamir's Secret Sharing. Recovering the file requires combining a threshold number of shares.

### Encrypt flow (`fragment encrypt --shares N --threshold K <path>`)
1. Generate an ephemeral X25519 key pair via `filippo.io/age`
2. Archive the input (file or directory) as a gzipped tar using `internal/archive`
3. Encrypt the archive using age with the X25519 recipient
4. Split the age identity string (the private key) into N shares using `internal/shamir` (Shamir's Secret Sharing over GF(256))
5. Each share is packaged as a tar.gz containing `SHARE.txt` (Shamir share bytes), `MANIFEST.age` (the full encrypted blob), and `metadata.json`
6. That tar.gz is encrypted with a random NaCl secretstream key prepended to the ciphertext, then written to `share-N.fragment`

### Decrypt flow (`fragment decrypt [--key <hex>] share-1.fragment ... share-K.fragment`)
- Each `.fragment` file is decrypted: either using the embedded key (first 32 bytes of the file) via `share.Decode`, or using an externally supplied hex key via `share.DecodeExternKey`
- Shamir shares are combined with `shamir.Combine` to recover the age private key
- The age identity decrypts the `MANIFEST.age` blob, producing the original tar.gz, which is written to `out.tar.gz`

### Key packages
- `internal/shamir` — GF(256) arithmetic (`gf256.go`), polynomial generation/evaluation (`polynomials.go`), and `Split`/`Combine` entry points (`shamir.go`)
- `internal/share` — `Share` struct serialization: packs Shamir share + encrypted blob + metadata into a NaCl-encrypted tar.gz; supports both self-keyed (`Decode`) and externally keyed (`DecodeExternKey`) decoding
- `internal/archive` — tar.gz helpers: create from filesystem paths (`CreateDirArchive`/`CreateArchive`) or from in-memory `FileEntry` slices (`CreateArchiveEnt`/`ReadArchiveEnt`)
- `cmd/` — Cobra subcommands `encrypt` and `decrypt`

### Dependencies
- `filippo.io/age` — authenticated file encryption (X25519 + ChaCha20-Poly1305)
- `github.com/nathants/go-libsodium` — NaCl secretstream for per-share envelope encryption
- `github.com/spf13/cobra` — CLI framework
