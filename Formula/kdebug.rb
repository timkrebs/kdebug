class Kdebug < Formula
  desc "CLI tool that automatically diagnoses common Kubernetes issues and provides actionable suggestions"
  homepage "https://github.com/timkrebs/kdebug"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/timkrebs/kdebug/releases/download/v1.0.0/kdebug-darwin-amd64"
      sha256 "1beb1c20dc6b8a4e9a8ae37a89110c7763c5bcf9b7ecb06dfb8a6d740ec122dd"

      def install
        bin.install "kdebug-darwin-amd64" => "kdebug"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/timkrebs/kdebug/releases/download/v1.0.0/kdebug-darwin-arm64"
      sha256 "dfea298609945139bf38ec98764ea7a1aa4ea6449ca4d64b8fa2457049ec90bc"

      def install
        bin.install "kdebug-darwin-arm64" => "kdebug"
      end
    end
  end

  def caveats
    <<~EOS
      kdebug is a Kubernetes debugging tool. Make sure you have:
      - kubectl installed and configured
      - Access to a Kubernetes cluster
      
      Get started with:
        kdebug pod --help
        kdebug cluster --help
    EOS
  end

  test do
    system "#{bin}/kdebug", "--version"
    system "#{bin}/kdebug", "--help"
  end
end