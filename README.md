# Lottip

Lottip is proxy for **MySQL RDBMS** with web GUI. It will show you what's happening under the hood of your ORM.
As it sits between your application and MySQL server there's no need to use tools like Wireshark to see which queries are being executed.
Lottip comes as single binary with zero dependencies and consists of 2 parts: proxy server and embedded GUI for better experience.

# Screenshots
Here's how query list looks like:
![Example 1](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/1.png)
Each query can be expanded for detailed view:
![Example 2](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/2.png)
Bottom query freezed for 10 seconds:
![Example 3](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/3.png)
Ooops, query returned an error:
![Example 4](https://raw.githubusercontent.com/orderbynull/lottip/master/shots/4.png)

# Main features
**Please note Lottip is on it's early stage of development.**

For now you can:
1. See all queries sent from your application to MySQL grouped by connection it belongs to.
2. Track query execution result: success(green indicator), pending(yellow indicator) and error(red indicator).
3. Expand/collapse each query to see more/less details.
4. See query execution time.

# ToDo
- [ ] Write Unit tests
- [ ] Implement every aspect of MySQL protocol
- [ ] Add query filtering by string or by it's status
- [ ] Add sql code highlighting
- [ ] Add sql code formatting
- [ ] Add possibility to execute/explain query right from GUI and see results
- [ ] Add support of PostgreSQL protocol 
- [ ] ... and more

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
