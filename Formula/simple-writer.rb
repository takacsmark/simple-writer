class SimpleWriter < Formula
  desc "Distraction-free terminal writer with Vim bindings and markdown rendering"
  homepage "https://github.com/OWNER/REPO"
  url "https://github.com/OWNER/REPO/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_TARBALL_SHA256"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", "-ldflags", "-s -w", "-o", bin/"simple", "./cmd/simple"
  end

  test do
    assert_match "raw mode", shell_output("#{bin}/simple 2>&1", 1)
  end
end
