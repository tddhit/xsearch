# XSearch 
A distributed high-performance light-weight search engine


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
$ go install github.com/tddhit/xsearch/cmd/xsearch
```

This will retrieve the project and install the `xsearch` command line utility into your `$GOBIN` path.

#### Starting metad service

```
$ xsearch metad 
```

#### Starting diskqueue cluster

```
$ diskqueue service --addr 127.0.0.1:9000 --cluster-addr 127.0.0.1:9001 --id node1 --mode cluster 
$ diskqueue service --addr 127.0.0.1:9100 --cluster-addr 127.0.0.1:9101 --id node2 --mode cluster --leader 127.0.0.1:9001 
$ diskqueue service --addr 127.0.0.1:9200 --cluster-addr 127.0.0.1:9201 --id node3 --mode cluster --leader 127.0.0.1:9001
```

#### Starting searchd service

```
$ xsearchd searchd --addr 127.0.0.1:10200 --admin 127.0.0.1:10201 --datadir ./searchd1_data --metad grpc://127.0.0.1:10100 --diskqueue grpc://127.0.0.1:9000 
$ xsearchd searchd --addr 127.0.0.1:10210 --admin 127.0.0.1:10211 --datadir ./searchd2_data --metad grpc://127.0.0.1:10100 --diskqueue grpc://127.0.0.1:9000 
```

#### Starting proxy service

```
$ xsearch proxy --addr 127.0.0.1:10300 --admin 127.0.0.1:10301 --namespaces news --metad grpc://127.0.0.1:10100 --diskqueue grpc://127.0.0.1:9000
```

#### Creating a namespace
```
$ xsearch client metad create  --namespace news --shardnum 3 --factor 2
```

#### Adding nodes to namespace
```
$ xsearch client metad add --namespace news  --node 127.0.0.1:10200
$ xsearch client metad add --namespace news  --node 127.0.0.1:10210
```

#### Automatic sharding
```
$ xsearch client metad balance --namespace news
```

#### Committing Operation
```
$ xsearch client metad commit --namespace news
```

#### Indexing Document
```
$ xsearch client proxy index --namespace news --content 'github被微软收购' --addr grpc://127.0.0.1:10300
```

#### Searching Query
```
$ xsearch client proxy search --namespace news --query 'github' --start 0 --count 10 --addr grpc://127.0.0.1:10300
```
