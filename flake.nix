{
  inputs = {
    staging.url = "github:NixOS/nixpkgs/staging";
    nixpkgs.url = "github:NixOS/nixpkgs/master";
    devenv.url = "github:cachix/devenv/v0.6.3";
    llm-agents.url = "github:numtide/llm-agents.nix";
  };

  nixConfig = {
    extra-trusted-public-keys =
      "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw=";
    extra-substituters = "https://devenv.cachix.org";
  };

  outputs = { self, nixpkgs, staging, devenv, ... }@inputs:
    let
      pkgs = nixpkgs.legacyPackages."x86_64-linux";
      bleedingEdge = staging.legacyPackages."x86_64-linux";
      llm-agents = inputs.llm-agents.packages."x86_64-linux";

    in {
      devShell.x86_64-linux = devenv.lib.mkShell {
        inherit inputs pkgs;

        modules = [
          ({ pkgs, lib, ... }: {

            # This is your devenv configuration
            packages = [
              pkgs.go
              pkgs.gopls
              pkgs.gotests
              pkgs.gosec
              pkgs.golangci-lint
              pkgs.gofumpt
              pkgs.gotools

              llm-agents.gemini-cli
            ];
          })
        ];
      };
    };
}
