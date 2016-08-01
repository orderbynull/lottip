# Lottip

MySQL proxy server with browser based GUI.
**Under development.**

# Installation

1) `go get -t github.com/orderbynull/lottip`

2) `go get github.com/mjibson/esc`

3) `go install github.com/mjibson/esc`

4) `cd $GOPATH/src/github.com/orderbynull/lottip`

5) `$GOPATH/bin/esc -o="embed/embed.go" -prefix="web" -ignore="\.idea.*|\.DS.*" -pkg="embed" web`

6) `go build -o server main.go` 

7) `./server --listen=:5050 --mysql=192.168.0.195:3306`