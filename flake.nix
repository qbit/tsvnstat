{
  description = "tsvnstat: a tailscale aware vmcstat server";

  inputs.nixpkgs.url = "github:nixos/nixpkgs";

  outputs = { self, nixpkgs }:
    let
      supportedSystems =
        [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in {
      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          tsvnstat = pkgs.buildGo121Module {
            pname = "tsvnstat";
            version = "v0.0.14";
            src = ./.;

            vendorHash = "sha256-96EtvRK7DRPF7y2iG8dtBCOjk3oRnPrQhmMf8c0WpqI=";
            proxyVendor = true;
          };
        });

      defaultPackage = forAllSystems (system: self.packages.${system}.tsvnstat);
      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            shellHook = ''
              PS1='\u@\h:\@; '
              nix run github:qbit/xin#flake-warn
              echo "Go `${pkgs.go}/bin/go version`"
            '';
            nativeBuildInputs = with pkgs; [ git go_1_21 gopls go-tools ];
          };
        });
    };
}

