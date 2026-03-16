{
  description = "Nix flake to package (locally) opcua-browser: OPC UA browser for the terminal";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "opcua-browser";
          version = "0.0.2";
          src = ./.;

          go = pkgs.go_1_26;
          vendorHash = "sha256-/URqW9ZgeWo22ucAyFwWzXUnBpBxYcgbwADsdIcJ2ZU=";
        };

        # This makes the package accessible in a development shell
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_26

            pkg-config # C dependencies
            sqlite
          ];
          CGO_ENABLED = 1;

          # nativeBuildInputs = with pkgs; [ git ];
        };
      }
    );
}
