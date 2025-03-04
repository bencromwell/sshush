{
  stdenv,
  fetchFromGitHub,
}:

stdenv.mkDerivation {
  pname = "sshush";
  version = "2.0.0";

  src = fetchFromGitHub {
    owner = "bencromwell";
    repo = "sshush";
    rev = "v2.0.0";
    sha256 = "SWQ6Whcib6QN30rwUbbyBW1+ovwc8K3Ocwp9372YcbQ=";
  };

  installPhase = ''
    mkdir -p $out/bin
    cp sshush $out/bin
  '';
}
