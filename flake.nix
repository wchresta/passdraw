{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        passdraw = pkgs.buildGoModule {
          pname = "passdraw";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-PYoO3JMlIbtF8sHm+pO2RQN6nJKIc001toGY7/b+t0I=";
        };
      in {
        packages = {
          inherit passdraw;
          default = passdraw;
        };

        devShell = pkgs.mkShell {
          inputsFrom = [ passdraw ];
          buildInputs = with pkgs; [
            pkgs.go_1_24
            pkgs.gotools
            pkgs.golangci-lint
            pkgs.gopls
            pkgs.go-outline
            pkgs.gopkgs

            pkgs.cobra-cli
          ];
        };
      });
}
