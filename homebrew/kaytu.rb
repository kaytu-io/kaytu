class Kaytu < Formula
  desc "CLI application for Kaytu"
  homepage "https://github.com/kaytu-io/cli-program"
  version "VERSION_HOMEBREW"
  license "MIT"

  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/kaytu-io/kaytu/releases/download/vVERSION_HOMEBREW/kaytu_VERSION_HOMEBREW_darwin_amd64.tar.gz"
    sha256 "HASH_MAC_AMD64"
    def install
      bin.install "kaytu" => "kaytu"
    end
  end

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/kaytu-io/kaytu/releases/download/vVERSION_HOMEBREW/kaytu_VERSION_HOMEBREW_darwin_arm64.tar.gz"
    sha256 "HASH_MAC_ARM64"
    def install
      bin.install "kaytu" => "kaytu"
    end
  end

  if OS.linux? && Hardware::CPU.arm?
    url "https://github.com/kaytu-io/kaytu/releases/download/vVERSION_HOMEBREW/kaytu_VERSION_HOMEBREW_linux_arm64.tar.gz"
    sha256 "HASH_LINUX_ARM64"
    def install
      bin.install "kaytu" => "kaytu"
    end
  end

  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/kaytu-io/kaytu/releases/download/vVERSION_HOMEBREW/kaytu_VERSION_HOMEBREW_linux_amd64.tar.gz"
    sha256 "HASH_LINUX_AMD64"
    def install
      bin.install "kaytu" => "kaytu"
    end
  end


  test do
    system "#{bin}/kaytu", "--version"
  end
end