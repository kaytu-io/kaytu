class Kaytu < Formula
  desc "CLI application for Kaytu"
  homepage "https://github.com/kaytu-io/cli-program"
  version "VERSION_HOMEBREW"
  license "MIT"

  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/kaytu-io/kaytu/releases/download/vVERSION_HOMEBREW/kaytu-darwin-amd64"
    sha256 "HASH_MAC_AMD64"
    def install
      bin.install "ktucli-macos-amd64" => "kaytu"
    end
  end

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/kaytu-io/kaytu/releases/download/vVERSION_HOMEBREW/kaytu-darwin-arm64"
    sha256 "HASH_MAC_ARM64"
    def install
      bin.install "ktucli-macos-arm64" => "kaytu"
    end
  end

  if OS.linux? && Hardware::CPU.arm?
    url "https://github.com/kaytu-io/kaytu/releases/download/vVERSION_HOMEBREW/kaytu-linux-arm64"
    sha256 "HASH_LINUX_ARM64"
    def install
      bin.install "ktucli-linux-arm64" => "kaytu"
    end
  end

  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/kaytu-io/kaytu/releases/download/vVERSION_HOMEBREW/kaytu-linux-amd64"
    sha256 "HASH_LINUX_AMD64"
    def install
      bin.install "ktucli-linux-amd64" => "kaytu"
    end
  end


  test do
    system "#{bin}/kaytu", "--version"
  end
end