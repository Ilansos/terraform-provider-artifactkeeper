# Publishing

This provider is packaged for the Terraform Registry through GitHub Releases.

## CI

Pull requests to `main` run `.github/workflows/ci.yml`.

The CI workflow:

- checks Go formatting,
- downloads modules,
- runs `go vet ./...`,
- runs `go test ./...`,
- builds the provider binary,
- runs `govulncheck ./...`.

It does not publish anything.

## Release Flow

Releases are created manually by pushing a semantic version tag:

```bash
git tag -s v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

You can also tag the merge commit explicitly by SHA:

```bash
git tag -s v0.1.0 <merge_commit_sha> -m "v0.1.0"
git push origin v0.1.0
```

The release workflow:

- runs `go test ./...`,
- builds cross-platform ZIP artifacts with GoReleaser,
- signs the checksum file with GPG,
- creates a GitHub Release.

The Terraform Registry discovers provider versions from those GitHub Releases.

## Required GitHub Secrets

Configure these repository secrets before the first release:

- `GPG_PRIVATE_KEY`: ASCII-armored private GPG key used to sign checksums.
- `GPG_PASSPHRASE`: Passphrase for the private key.
- `GPG_FINGERPRINT`: Fingerprint or key ID passed to `gpg --local-user`.

`GITHUB_TOKEN` is provided automatically by GitHub Actions.

## Terraform Registry Expectations

The GoReleaser configuration publishes assets named like:

```text
terraform-provider-artifactkeeper_0.1.0_linux_amd64.zip
terraform-provider-artifactkeeper_0.1.0_manifest.json
terraform-provider-artifactkeeper_0.1.0_SHA256SUMS
terraform-provider-artifactkeeper_0.1.0_SHA256SUMS.sig
```

The provider binary inside each ZIP is named like:

```text
terraform-provider-artifactkeeper_v0.1.0
```

The manifest asset is generated from `terraform-registry-manifest.json` and declares Terraform Plugin Framework protocol version `6.0`.

The GitHub repository should include a public key matching the signing key, usually as an ASCII-armored file in the repository root or in the release notes according to the Terraform Registry publishing instructions.

## Dry Run

Before creating a real tag, run:

```bash
goreleaser release --snapshot --clean
```

The dry run writes artifacts into `dist/` and does not publish a GitHub Release.
