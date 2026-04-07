{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in
    {
      packages.${system} = {
        default = pkgs.buildGoModule {
          env.CGO_ENABLED = 0;
          pname = "countmein";
          version = "0.0.1";
          vendorHash = "sha256-TVVLwQyBesy9HFa9OLtcNy9UBjQcGfalQfL9J8WPAiI=";
          src = self;
        };
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          atlas
          buf
          gnumake
          go
          httpie
          lsof
          mariadb
          openssh
          protoc-gen-connect-go
          protoc-gen-doc
          protoc-gen-go
          sqlc
          sshpass
          tbls
        ];
      };
    };
}
