# wonaming
Consul-based naming impl for grpc balancer.

## example

### etcd

#### client
go run main.go --reg http://127.0.0.1:2379

#### server
go run main.go --reg http://127.0.0.1:2379


### consul

#### client
go run main.go --reg 127.0.0.1:8500

#### server
go run main.go --reg 127.0.0.1:8500
