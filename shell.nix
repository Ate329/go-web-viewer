{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gotools
    gopls
    go-outline
    gopkgs
    godef
    golint
  ];

  shellHook = ''
    export GOPATH=$PWD/.go
    export PATH=$GOPATH/bin:$PATH
    if [ ! -f go.mod ]; then
      go mod init gotermbrowser
      go get github.com/gdamore/tcell/v2
      go get github.com/rivo/tview
      go get golang.org/x/net/html
    fi

    if [ ! -f go.sum ]; then
      go mod tidy
    fi
  '';
}
