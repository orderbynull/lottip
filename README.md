# Lottip

Lottip is proxy for **MySQL RDBMS** with web GUI. It will show you what's happening under the hood of your database layer.
As it sits between your application and MySQL server there's no need to use tools like Wireshark to see which queries are being executed.
It comes as single binary with zero dependencies and consists of 2 parts: proxy server and embedded GUI.

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
**Lottip is on it's early stage of development.**

For now you can:
1. See all queries sent from your application to MySQL grouped by connection it belongs to.
2. Track query execution result: success(green indicator), pending(yellow indicator) and error(red indicator).
3. Expand/collapse each query to see more/less details.
4. See query execution time.
5. Filter queries by string

# Installation
###### Binary
Get binaries from [releases](https://github.com/orderbynull/lottip/releases) page

###### From sources on Mac/Linux
    go get github.com/orderbynull/lottip
    go install github.com/mjibson/esc
    cd $GOPATH/src/github.com/orderbynull/lottip
    $GOPATH/bin/esc -o static/static.go -pkg="static" -prefix web -include=".*\.css|.*\.js|.*\.html|.*\.png" web
    go build
    ./lottip
    
# How to run
There're 4 simple steps to get everything up and running:
1. Run binary from terminal like this: `./lottip_darwin_386`.
You'll see something like this:

    `Forwarding queries from '127.0.0.1:4041' to '127.0.0.1:3306'`
    
    `Web gui available at 'http://127.0.0.1:9999'`
     
2. Tell your app to connect to MySQL via port *4041* instead of *3306*.
3. Go to [http://127.0.0.1:9999](http://127.0.0.1:9999) and you'll see nice GUI.
4. Play with your app and see all SQL-queries flowing between your app and MySQL. 
**No need for page refresh because websockets are used to communicate between frontend and backend.**

# Use cases
Here're few use cases i use on my everyday basis so it may be helpful to someone.

###### Use locally
Just run lottip on your local machine and point your app to it.
You can also run few lottip instances each on it's own port. 
This is an easy way to keep multiple app separated and view queries independently.

###### Use remotely
Let's say you're writing your PHP code locally but run it on dev server and do not want to expose lottip to outside world.
In this case here's what you may do:
1. Upload lottip binary to remote dev server and run it like this: `./lottip_darwin_386`
2. Create ssh tunnel from your local machine to remote dev server like this: `ssh -nNT -L 9999:127.0.0.1:9999 user@your-devserver.com`.
   This command will map your local `:9999` to remote `:9999`
3. Tell your remote app to use MySQL on port `:4041`
4. Open [http://127.0.0.1:9999](http://127.0.0.1:9999) locally.

# Options

You can change default values to whatever you need.

| option available       |  default value  | description                                                                                                          
| ---------------------- |-----------------|-------------------------------------------------------------------------------------------------  
| `--proxy`              | `127.0.0.1:4041`|`<ip>:<port>` of proxy server. Your code should make connections to that address to make proxy work. *Example: `--proxy=127.0.0.1:4045`*        
| `--mysql`              | `127.0.0.1:3306`|`<ip>:<port>` of MySQL server. *Example: `--mysql=192.168.0.195:3308`*
| `--gui`                | `127.0.0.1:9999`|`<ip>:<port>` of embedded GUI. *Example: `--gui=127.0.0.1:8080`*

# ToDo
- [ ] Write Unit tests
- [ ] Implement more features of MySQL protocol
- [x] Add query filtering by string
- [ ] Add sql code highlighting
- [ ] Add sql code formatting
- [x] Add possibility to execute query right from GUI and see results
- [ ] Add ssl support
- [ ] Add support of PostgreSQL protocol 
- [ ] ... and more

# Known problems
Currently lottip does not support secure connections via ssl. The workaround is to disable SSL on MySQL server or connect with option like [--ssl-mode=DISABLED](https://dev.mysql.com/doc/refman/5.7/en/secure-connection-options.html#option_general_ssl-mode)

# Contribute
You're very welcome to report bugs, make pull requests, share your thoughts and ideas!

# Licence
MIT
