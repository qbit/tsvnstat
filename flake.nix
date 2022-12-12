{
  description = "tsvnstat: a tailscale aware vmcstat server";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

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
          tsvnstat = pkgs.buildGoModule {
            pname = "tsvnstat";
            version = "v0.0.9";
            src = ./.;

            vendorSha256 = "sha256-kneXJLcNfKSvXpPmUU3Y7oFcVlknQ9Dsjd4Rq4bJrGo=";
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
              echo "Go `${pkgs.go}/bin/go version`"
            '';
            nativeBuildInputs = with pkgs; [ git go gopls go-tools ];
          };
        });
    };
}

