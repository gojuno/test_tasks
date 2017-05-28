# Juno Inc., Test Task: In-memory cache

Simple implementation of Redis-like in-memory cache

Desired features:
- Key-value storage with string, lists, dict support
- Per-key TTL
- Operations:
  - Get
  - Set
  - Update
  - Remove
  - Keys
- Custom operations(Get i element on list, get value by key from dict, etc)
- Golang API client
- Telnet-like/HTTP-like API protocol

Provide some tests, API spec, deployment docs without full coverage, just a few cases and some examples of telnet/http calls to the server. 

Optional features:
- persistence to disk/db **NOT IMPLEMENTED**
- scaling(on server-side or on client-side, up to you) **IMPLEMENTED OVER CONSISTENT HASHING IN CLIENT**
- auth **NOT IMPLEMENTED**
- perfomance tests **go bench IMPLEMENTED ONLY FOR SET,GET,LPUSH,LRANGE**

### Current only following redis commands are implemented
- GET
- SET
- DEL
- EXPIRE
- LPUSH
- LRANGE
- LSET
- LREM
- HGET
- HSET
- HDEL

### Installation and run docker container, run internal benchmark and push new images when successfull inside Vagrant
    vagrant plugin install hostmanager
    vagrant up --provision
    vagrant ssh juno_test
    sudo bash
    cd /vagrant
    ./run_docker.sh

### Native docker installation and run inside docker-compose
    docker-compose build
    docker-compose up -d kv_server 
    docker-compose run --rm kv_client

### Run redis-benchamrk inside Vagrant
    vagrant plugin install hostmanager
    vagrant up --provision
    vagrant ssh juno_test
    sudo bash
    cd /vagrant
    GOPATH=/vagrant go run cmd/kv_server.go &
    redis-benchmark -p 8379 -n 50000 -c 100 -t set,get,lpush,lrange,hget,hset -q 

### Run inline benchamrk inside Vagrant
    vagrant up --provision
    vagrant ssh juno_test
    sudo bash
    cd /vagrant
    GOPATH=/vagrant go test junoKvClient -test.v -test.bench . -test.run ^$ -test.benchtime 10s
    GOPATH=/vagrant go test junoKvServer -test.v -test.bench . -test.run ^$ -test.benchtime 10s
    
### Contributing rules
    # install http://hub.github.com
    # create new issue on https://github.com/Slach/test_tasks/issues
    git clone git@github.com:Slach/test_tasks.git ./juno-test
    cd ./juno-test

    git config core.autocrlf false
    git config core.eol lf
    git config user.name <YourName>
    git config user.email <Your@Email>

    git flow init
    git flow feature start feature_name
    git add *
    git commit -s -m "commit description and link to issue i.e. #1"
    git flow feature publish feature_name
    hub pull-request --browse -m "Implemented feature X see issue #1" -i 1 -b Slach:master -h Slach:feature_name 
    
### Simple client API usage
```go
    package main
    import (
        junoKvClient "github.com/Slach/juno-test/src/junoKvClient"
        "bytes"
        "log"
    )
    
    func main() {
        // @todo maybe consul service discovery need here? ;)
        servers_with_weight := map[string]int{
            "localhost:8379":1,
            //"host2:8379":2, // use multiple host weight for add server
        }
        c := junoKvClient.NewClient(servers_with_weight)
        v := 1
        c.Set("test",v, 0)        
        if !bytes.Equal([]byte(c.Get("test")),[]byte(v)) {
            log.Fatal("test FAILED")
        }
    }
```