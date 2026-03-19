{
  description = "Nix flake to package (locally) opcua-browser, an OPC UA browser for the terminal";

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
        packages.default = pkgs.buildGoModule.override { go = pkgs.go_1_26; } {
          pname = "opcua-browser";
          version = "0.0.3";
          src = ./.;

          vendorHash = "sha256-QUOIY64FuDSb1AU55RGhj5ewlgexsKMRXRKwfcb+HlQ=";

          installPhase = ''
            	mkdir -p $out/share/applications
            	cp opcua-browser.desktop $out/share/applications
            	'';

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
