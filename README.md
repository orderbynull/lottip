# Lottip

MySQL proxy server with browser based GUI.
**Under development.**

#Screenshots
![Example 1](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/1.png)
![Example 2](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/2.png)

# Installation from sources

1) `go get -t github.com/orderbynull/lottip`

2) `go get github.com/mjibson/esc`

3) `go install github.com/mjibson/esc`

4) `cd $GOPATH/src/github.com/orderbynull/lottip`

5) `$GOPATH/bin/esc -o="embed/embed.go" -prefix="web" -ignore="\.idea.*|\.DS.*" -pkg="embed" web`

6) `go build -o server main.go` 

7) `./server --listen=:4040 --mysql=192.168.0.195 --port=3306`