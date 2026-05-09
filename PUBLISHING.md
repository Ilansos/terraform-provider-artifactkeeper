# Publishing

This provider is packaged for the Terraform Registry through GitHub Releases.

## CI

Pull requests to `main` run `.github/workflows/ci.yml`.

The CI workflow:

- checks Go formatting,
- downloads modules,
- runs `go test ./...`,
- builds the provider binary.

It does not publish anything.

## Release Flow

Releases are created automatically when changes are merged to `main`.

The release workflow:

- runs `go test ./...`,
- finds the latest `vMAJOR.MINOR.PATCH` tag,
- computes the next version,
- creates and pushes the new tag on the merge commit,
- builds cross-platform ZIP artifacts with GoReleaser,
- signs the checksum file with GPG,
- creates a GitHub Release.

The Terraform Registry discovers provider versions from those GitHub Releases.

## Version Bumping

The release workflow supports automatic semantic version bumps from commit messages since the last release tag:

- `major`: any commit message containing `BREAKING CHANGE` or a Conventional Commit breaking marker such as `feat!:`.
- `minor`: any commit message starting with `feat:` or `feat(scope):`.
- `patch`: all other changes.

If there are no existing release tags, the workflow starts from `v0.0.0`, so the first patch release is `v0.0.1`.

You can also run the release workflow manually from GitHub Actions and choose `patch`, `minor`, or `major` with the `workflow_dispatch` input.

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
