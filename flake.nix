{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    nix-appimage.url = "github:ralismark/nix-appimage";
  };

  outputs =
    {
      self,
      nixpkgs,
      nix-appimage,
      ...
    }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };

      leptWithAlias = pkgs.symlinkJoin {
        name = "leptonica-with-alias";
        paths = [ pkgs.leptonica ];

        postBuild = ''
          ln -s $out/lib/libleptonica.so $out/lib/liblept.so
        '';
      };
    in
    {
      packages.${system} = {
        appimage = nix-appimage.bundlers.${system}.default self.packages.${system}.default;

        default = pkgs.buildGoModule {
          pname = "countmein";
          version = "0.0.1";
          vendorHash = "sha256-dyhDOEv7IUPhdz+PL1C0AC55oxna2XRdmwFRPleHS8s=";
          src = self;
          nativeBuildInputs = with pkgs; [ pkg-config ];
          meta.mainProgram = "countmein";

          buildInputs = [
            leptWithAlias
            pkgs.tesseract
          ];
        };
      };

      devShells.${system}.default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [ pkg-config ];

        packages = with pkgs; [
          atlas
          buf
          coreutils
          findutils
          gnumake
          go
          httpie
          lsof
          mariadb
          openssh
          protoc-gen-connect-go
          protoc-gen-doc
          protoc-gen-go
          rsync
          sqlc
          sshpass
          tbls
        ];

        buildInputs = with pkgs; [
          leptWithAlias
          tesseract
        ];
      };
    };
}
