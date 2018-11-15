# XSearch 
A distributed high-performance search engine


### Features
 * Pure go implement
 * Metadata service centralization control
 * Two-level mapping sharding
 * Use binlog for replication
 * Near realtime
 * Provides a plugin mechanism to implement custom query analysis and rerank algorithms
 * Supports automatic or manual migration of sharding

### Architecture

```
graph TB
Metad-->Searchd1
Metad-->Searchd2
Searchd1-->Proxy1
Searchd1-->Proxy2
Searchd2-->Proxy1
Searchd2-->Proxy2
```

### Getting Started

#### Installing

To start using XSearch, install Go and run:
```
$ go get github.com/tddhit/xsearch
$ cd $GOPATH/src/github.com/tddhit/xsearch/cmd/xsearch
$ go build
$ cd $GOPATH/src/github.com/tddhit/diskqueue/cmd/diskqueue
$ go build
```

This will retrieve the xsearch code and the dependent libraries, and build the `xsearch`/`diskqueue` command line utility.

#### Starting diskqueue cluster

```
$ cd $GOPATH/src/github.com/tddhit/diskqueue/cmd/diskqueue
$ ./diskqueue service --addr 127.0.0.1:9000 --cluster-addr 127.0.0.1:9010 --id node1 --mode cluster --datadir ./data1 --pidpath ./diskqueue1.pid
$ ./diskqueue service --addr 127.0.0.1:9100 --cluster-addr 127.0.0.1:9110 --id node2 --mode cluster --datadir ./data2 --leader grpc://127.0.0.1:9000 --pidpath ./diskqueue2.pid
$ ./diskqueue service --addr 127.0.0.1:9200 --cluster-addr 127.0.0.1:9210 --id node3 --mode cluster --datadir ./data3 --leader grpc://127.0.0.1:9000 --pidpath ./diskqueue3.pid
```

#### Starting metad service

```
$ cd $GOPATH/src/github.com/tddhit/xsearch/cmd/xsearch
$ ./xsearch metad 
```

#### Starting searchd service

```
$ ./xsearch searchd --addr 127.0.0.1:10200 --admin 127.0.0.1:10201 --datadir ./searchd1_data --pidpath ./searchd1.pid --metad grpc://127.0.0.1:10100 --diskqueue diskqueue://127.0.0.1:9000,127.0.0.1:9100,127.0.0.1:9200/leader
$ ./xsearch searchd --addr 127.0.0.1:10210 --admin 127.0.0.1:10211 --datadir ./searchd2_data --pidpath ./searchd2.pid --metad grpc://127.0.0.1:10100 --diskqueue diskqueue://127.0.0.1:9000,127.0.0.1:9100,127.0.0.1:9200/leader 
```

#### Starting proxy service

```
$ ./xsearch proxy --addr 127.0.0.1:10300 --admin 127.0.0.1:10301 --namespaces news --pidpath ./proxy.pid --metad grpc://127.0.0.1:10100 --diskqueue diskqueue://127.0.0.1:9000,127.0.0.1:9100,127.0.0.1:9200/leader
```

#### Creating a namespace
```
$ ./xsearch client metad create  --namespace news --shardnum 3 --factor 2
$ ./xsearch client metad info
```

#### Adding nodes to namespace
```
$ ./xsearch client metad add --namespace news  --node 127.0.0.1:10200
$ ./xsearch client metad add --namespace news  --node 127.0.0.1:10210
$ ./xsearch client metad info
```

#### Automatic sharding
```
$ ./xsearch client metad balance --namespace news
$ ./xsearch client metad info
```

#### Committing Operation
```
$ ./xsearch client metad commit --namespace news
$ ./xsearch client metad info
$ ./xsearch client searchd info --addr grpc://127.0.0.1:10201
$ ./xsearch client searchd info --addr grpc://127.0.0.1:10211
$ ./xsearch client proxy info --addr grpc://127.0.0.1:10301
```

#### Indexing Document
```
$ ./xsearch client proxy index --namespace news --content 'Microsoft completes GitHub acquisition' --addr grpc://127.0.0.1:10300
```

#### Searching Query
```
$ ./xsearch client proxy search --namespace news --query 'github' --start 0 --count 10 --addr grpc://127.0.0.1:10300
```
