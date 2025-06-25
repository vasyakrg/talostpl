class Talostpl < Formula
  desc "Interactive and non-interactive Talos K8s config generator"
  homepage "https://github.com/vasyakrg/talostpl"
  url "https://github.com/vasyakrg/talostpl/releases/download/v1.0.0/talostpl-darwin-arm64"
  version "1.0.0"
  sha256 "9843d546bd541b9bf58e2e1c3c85aa8ff0b2f3705630b84b9123be55a99d5202"

  def install
    bin.install "talostpl-darwin-arm64" => "talostpl"
  end

  test do
    system "#{bin}/talostpl", "--version"
  end
end
