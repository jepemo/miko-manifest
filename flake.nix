{
  description = "Configuration management for Kubernetes manifests with templating capabilities";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = if (self ? rev) then "git-${builtins.substring 0 7 self.rev}" else "dev";
      in {
        packages.default = pkgs.buildGoModule {
          pname = "miko-manifest";
          version = version;
          src = ./.;
          
          # This will need to be updated after the first build
          # Run: nix build .# 2>&1 | grep -o 'got:.*' | cut -d' ' -f2
          vendorHash = null;
          
          ldflags = [
            "-s" "-w"
            "-X main.version=${version}"
            "-X main.commit=${self.rev or "dirty"}"
            "-X main.buildTime=unknown"
          ];

          postInstall = ''
            mkdir -p $out/share/miko-manifest
            cp -r templates $out/share/miko-manifest/
            cp -r config $out/share/miko-manifest/
          '';

          meta = with pkgs.lib; {
            description = "Configuration management for Kubernetes manifests with templating capabilities";
            homepage = "https://github.com/jepemo/miko-manifest";
            license = licenses.mit;
            maintainers = [ ];
            platforms = platforms.unix;
          };
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            goreleaser
          ];
        };
      });
}
