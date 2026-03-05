# Releasing

## Tag and push

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Pushing a `v*` tag triggers `.github/workflows/release.yml`.

## What the workflow produces

- `simple-writer_<version>_darwin_amd64.tar.gz`
- `simple-writer_<version>_darwin_arm64.tar.gz`
- `simple-writer_<version>_linux_amd64.tar.gz`
- `simple-writer_<version>_linux_arm64.tar.gz`
- `SHA256SUMS.txt`

All assets are attached to the GitHub Release for the tag.
