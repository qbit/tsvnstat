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
            version = "v0.0.11";
            src = ./.;

            vendorSha256 = "sha256-QcuL7qjNTloLBYnT2pZtYDZlj2/qm8sNS7EUVBbCp7U=";
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

