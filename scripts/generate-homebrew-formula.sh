#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 3 || $# -gt 4 ]]; then
  echo "usage: $0 <version> <github_owner> <repo> [output_formula_path]" >&2
  echo "example: $0 v0.1.0 acme simple-writer ./Formula/simple-writer.rb" >&2
  exit 1
fi

version="$1"
owner="$2"
repo="$3"
out_path="${4:-./Formula/simple-writer.rb}"

case "$version" in
  v*) ;;
  *)
    echo "version must start with 'v' (example: v0.1.0)" >&2
    exit 1
    ;;
esac

src_url="https://github.com/${owner}/${repo}/archive/refs/tags/${version}.tar.gz"

echo "fetching source tarball: ${src_url}"
sha256="$(curl -fsSL "${src_url}" | shasum -a 256 | awk '{print $1}')"

mkdir -p "$(dirname "${out_path}")"

cat >"${out_path}" <<EOF
class SimpleWriter < Formula
  desc "Distraction-free terminal writer with Vim bindings and markdown rendering"
  homepage "https://github.com/${owner}/${repo}"
  url "${src_url}"
  sha256 "${sha256}"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", "-ldflags", "-s -w", "-o", bin/"simple", "./cmd/simple"
  end

  test do
    assert_match "raw mode", shell_output("#{bin}/simple 2>&1", 1)
  end
end
EOF

echo "wrote formula to ${out_path}"
