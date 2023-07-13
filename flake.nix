{
  description = "ChessOpeningAnalyzer development flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
  }: let
    system = "x86_64-linux";
    pkgs =
      import nixpkgs {inherit system;};
    version = "0.2.0";

    scripts = let
      colors = {
        reset = "\\033[0m";
        red = "\\033[0;31m";
        green = "\\033[0;32m";
      };
      echo_success = ''echo -e "${colors.green}Success!${colors.reset}"'';
      echo_failed = ''echo -e "${colors.red}Failed${colors.reset}"'';
    in
      builtins.attrValues (builtins.mapAttrs pkgs.writeShellScriptBin {
        build-executables = let
          path_to_cli = "./cmd/opening_analyzer_cli";
          filename_endings = {
            "linux" = "";
            "windows" = ".exe";
            "darwin" = ".app";
            "freebsd" = "-freebsd";
            "netbsd" = "-netbsd";
            "openbsd" = "-openbsd";
          };
          build-for = platform: ending: ''
            echo "building for ${platform}..."
            CGO_ENABLED=0 GOOS=${platform} \
              go build -a -installsuffix cgo \
                -o dist/${version}/openinganalyzer${ending} ${path_to_cli} \
              && ${echo_success} || ${echo_failed}
          '';
          build-all = with builtins;
            concatStringsSep "\n" (attrValues
              (mapAttrs build-for filename_endings));
        in ''
          mkdir -p dist/${version}
          ${build-all}
        '';

        run-tests = ''
          go test ./... -coverprofile=cover.out
          go tool cover -html=cover.out -o cover.html
          rm cover.out
        '';
      });
    back = with pkgs; [go_1_20 gopls delve];
  in {
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = back ++ scripts;
      CGO_ENABLED = 0; # delve from nixpkgs refuses to work otherwise
    };
  };
}
