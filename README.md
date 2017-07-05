# Lottip

Lottip is proxy for **MySQL RDBMS** with web GUI. It will show you what's happening under the hood of your database layer.
As it sits between your application and MySQL server there's no need to use tools like Wireshark or enable general logs to see which queries are being executed.
It comes as single binary with zero dependencies and consists of 2 parts: proxy server and embedded GUI.

# Screenshots
|  Query list   |Freezed query|Query error   |In app results       |
|---------------|-------------|--------------|---------------------|
| ![Example 1](https://1.downloader.disk.yandex.ru/disk/82afbaa6ce9aad75d933f5b301786370158163a0adc0946bbb1c57586c00048e/5952ccb4/fKqInKw3d7bLFOeFnMGnhIpCh-g1Fsa4ASTpYYM1Tq82SVhYPSUYXiw5SnQjlbcFsWJBAtfosLEWTLDBVSUHpAtqvJgORLVMKQXgkxmqx4ar8npumZHI4midPdWhecNq?uid=0&filename=query-list.png&disposition=inline&hash=&limit=0&content_type=image%2Fpng&fsize=529592&hid=9cf9439e72300a253f78baa1c5ac1b9e&media_type=image&tknv=v2&etag=2b9e9b1c4ab35daec10ad11d196b00f6)           |![Example 2](https://4.downloader.disk.yandex.ru/disk/836e493fc28cad669e4b9e08f77dbad5ffefcbcb75f12cc82144fb889cb4e1f0/5952cce1/fKqInKw3d7bLFOeFnMGnhPyNSkKL-VufKsCx8zmN0C_LvxSEN1CQimedqoRS48qfyYUGTtS4vgYL-MRO3PJt1i9iIgiAuomYMwlRghktoIer8npumZHI4midPdWhecNq?uid=0&filename=sleep.png&disposition=inline&hash=&limit=0&content_type=image%2Fpng&fsize=490096&hid=9fa692657191070cc11adac12c41b091&media_type=image&tknv=v2&etag=72184792faf353d534ba0bad4bbb00e8)             |![Example 3](https://4.downloader.disk.yandex.ru/disk/193675e44a0db993aa65e58900ef1f7f6b4ecd01448480ed90600c2b1e483495/5952cd0e/fKqInKw3d7bLFOeFnMGnhJWYcZLrDbJ2m4Ak3bcAhlbBVUXp0oG_6XUrPQumLMcZp2j7Ne5xIPM6_Kl6F49JukxTSTqvk8hav2RNxfSeYhKr8npumZHI4midPdWhecNq?uid=0&filename=error.png&disposition=inline&hash=&limit=0&content_type=image%2Fpng&fsize=475041&hid=5a82c44688dcb6a3ceec8cfb04d03143&media_type=image&tknv=v2&etag=e52f07e3149ab48ebbbb3805466d88b8)              |![Example 4](https://2.downloader.disk.yandex.ru/disk/28539b74a03c0f8e08ad048eca4cd0a00d3145c5261c66291f31f9837459a55b/5952cd2a/fKqInKw3d7bLFOeFnMGnhEPh6u0x8S8JJr_NiRfyge4I2oaBbqGsQkgPUs-6m9Q6GBCqLoGbFe9sx_rNyVeCQAnRijJwHc7qQqqKW9RpAkmr8npumZHI4midPdWhecNq?uid=0&filename=dialog.png&disposition=inline&hash=&limit=0&content_type=image%2Fpng&fsize=586418&hid=6a2c28a6b60fc17dac6672823d00cdc8&media_type=image&tknv=v2&etag=5c1abf8803c6cbc22ab923645a7f6a75)             |

# Main features
**Lottip is on it's early stage of development.**

For now you can:
1. See all queries sent from your application to MySQL grouped by connection it belongs to.
2. Track query execution result: success(green indicator), pending(yellow indicator) and error(red indicator).
3. See query execution time(it includes time to transfer data over network).
4. Filter queries by string.
5. Execute any query and see results immediately.

# Installation
###### Binary
Get binaries from [releases](https://github.com/orderbynull/lottip/releases) page

###### From sources on Mac/Linux
    go get github.com/orderbynull/lottip
    go install github.com/mjibson/esc
    cd $GOPATH/src/github.com/orderbynull/lottip
    $GOPATH/bin/esc -o fs.go -prefix web -include=".*\.css|.*\.js|.*\.html|.*\.png" web
    go build
    ./lottip
    
# How to run
There're 4 simple steps to get everything up and running:
1. Run binary from terminal like this: `./lottip_linux_amd64`.
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
1. Upload lottip binary to remote dev server and run it like this: `./lottip_linux_amd64`
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
| `--mysql-dsn`          | `""`            |If you need to execute queries from the app you need to provide DSN for MySQL server. DSN format: `[username[:password]@][protocol[(address)]]/[dbname[?param1=value1&...&paramN=valueN]]` All values are optional. So the minimal DSN is `/dbname`. If you do not want to preselect a database, leave `dbname` empty: `/` *Example: `--mysql-dsn=root:root@/`*

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
