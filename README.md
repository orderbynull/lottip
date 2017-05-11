# Lottip

Lottip is a proxy for **MySQL RDBMS** with simple and clean GUI. 
It's goal is to help developers to debug persistence layer of their apps. 
Lottip can **show SQL queries** within each connection, **filter** statements and more.
It сomes as single  binary and consists of 2 parts: proxy server and embedded GUI for better experience.

#Screenshots
![Example 1](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/1.png)
![Example 2](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/2.png)

# Main features
TODO

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

# How to use

There're 4 simple steps to get everything up and running:
1. Run Lottip binary from console like this: `./lottip --listen=127.0.0.1:4040 --mysql=127.0.0.1:3306`.
2. Tell your app to connect to MySQL via port *4040* instead of *3306*.
3. Go to [http://127.0.0.1:8080](http://127.0.0.1:8080) and you'll see nice GUI.
4. Play with your app and see all SQL-queries flowing between your app and MySQL. No need for page refresh.

 

# Options
| option available       |  default value  | description                                                                                                          
| ---------------------- |-----------------|-------------------------------------------------------------------------------------------------  
| `--listen`             | `127.0.0.1:4040`|`<ip>:<port>` of proxy server. Your code should make connections to that address to make proxy work. *Example: `--listen=:4040`*        
| `--mysql`              | `127.0.0.1:3306`|`<ip>:<port>` of MySQL server. *Example: `--mysql=192.168.0.195:3306`*
| `--gui`                | `127.0.0.1:8080`|`<ip>:<port>` of embedded GUI. *Example: `--gui=127.0.0.1:8080`*
| `--verbose`            |      `false`    |Print debug information to console. *Example: `--verbose=true`*           

# Change log
TODO

# Licence
TODO