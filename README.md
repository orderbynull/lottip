# Lottip

MySQL proxy server with browser based GUI.
**Under heavy development.**

# Installation

1) Build static file to embed into binary: `./esc -o static.go -prefix web -ignore="\.idea.*|\.DS.*" web`

2) Buld with `static.go`: `go build lottip.go static.go`

3) Run: `./lottip --listen=:5050 --mysql=192.168.0.195:3306`
