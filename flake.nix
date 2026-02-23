{
  description = "A Go development shell with CGO dependencies";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }@inputs:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        # sourceFiles = ./.;
      in
      {
        devShells.default = pkgs.mkShell {
          # Include Go and any C libraries needed by CGO
          buildInputs = with pkgs; [
            go
            # Example C dependencies for a project using cgo with sqlite
            pkg-config
            sqlite
          ];

          # Set environment variables if necessary, for example CGO_ENABLED=1 is default but good to be explicit
          CGO_ENABLED = 1;

          # Optional: additional tools for development
          # nativeBuildInputs = with pkgs; [ git ];
        };
      }
    );
}
