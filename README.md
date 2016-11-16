# Lottip

Lottip is a proxy for **MySQL RDBMS** with simple and clean GUI. 
It's goal is to help developers to debug persistence layer of their apps. 
Lottip can **show SQL queries** within each connection, **filter** statements, create **SQL gists** and more.
It consists of 2 parts: proxy server and embedded GUI for better experience.

#Screenshots
![Example 1](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/1.png)
![Example 2](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/2.png)

# Main features


# Installation
###### From sources
    go get -t github.com/orderbynull/lottip
    go get github.com/mjibson/esc
    go install github.com/mjibson/esc
    cd $GOPATH/src/github.com/orderbynull/lottip
    $GOPATH/bin/esc -o="embed/embed.go" -prefix="web" -ignore="\.idea.*|\.DS.*" -pkg="embed" web
    go build -o server main.go
    ./server --listen=:4040 --mysql=192.168.0.195 --port=3306
    
###### Binary
TODO

# Usage
Here's an example of how to run Lottip: `./server --listen=:4040 --mysql=127.0.0.1`

This tells server to listen on any interface on port **4040** for incoming MySQL connections and to proxy them to local MySQL server listening on default port **3306**.
 
Now go to [http://127.0.0.1:8080](http://127.0.0.1:8080) and you'll see nice GUI.

# Options
| option available       | description                                                                                                          
| ---------------------- | ----------------------------------------------------------------------------------------------------  
| `--listen`             | `<ip>:<port>` of proxy server. Your code should make connections to that address to make proxy work. Example: `--listen=:4040`        
| `--mysql`              | MySQL server address. Example: `--mysql=192.168.0.195`
| `--port`               | MySQL server port. Example: `--port=3306`          
| `--addr`               | `<ip>:<port>` of embedded GUI. Example: `--addr=127.0.0.1:8081`
| `--verbose`            | Print debug information to console. Example: `--verbose=true`           

# Default options
TODO

# Change log
TODO

# Licence
TODO