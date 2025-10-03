# Building moxli

## Standard Build Process

**Always use the Makefile for building.** This ensures consistent output locations and version information.

```bash
make build
```

This will:
- Create the `dist/` directory if it doesn't exist
- Build the binary to `dist/moxli` with version information embedded
- Include git commit hash and build date

## Running the Binary

After building:

```bash
./dist/moxli
```

## Other Make Targets

- `make test` - Run all tests
- `make lint` - Run golangci-lint
- `make fmt` - Format Go code
- `make clean` - Remove build artifacts
- `make install` - Install to $GOPATH/bin (for local development)

## ⚠️ Important: Do NOT use `go build` directly

Running `go build` directly creates binaries in inconsistent locations:
- `go build ./cmd/moxli` → creates `moxli-bin` in project root
- `go build -o bin/moxli ./cmd/moxli` → creates `bin/moxli`

These locations are gitignored but shouldn't be created in the first place.

**Always use `make build` instead**, which:
- Outputs to the consistent `dist/moxli` location
- Includes version information via ldflags
- Matches team conventions

## Version Information

The Makefile automatically injects:
- `VERSION` - defaults to 0.1.0, override with `make build VERSION=1.0.0`
- Git commit hash (short)
- Build timestamp (UTC)
