# shitcoin

A simplified blockchain application for the sake of learning and testing critical components that a blockchain network needs.

## Requirements

[Go](http://golang.org) 1.16 or newer.

## Installation

```bash
go install
```

## Getting Started

### Docker

You can simply use the [docker-compose. yml](https://github.com/alikarimi999/shitcoin/blob/main/docker-compose.yml) file and set up a Blockchain network with four Nodes that are synced and connected to each other.

```bash
$ docker-compose up
```

#### NOTES

1. before running the docker-compose file you need to set a "MINER" environment variable with a valid address(you can generate valid addresses with [wallet](https://github.com/alikarimi999/wallet#new-account) app)
for each NODE's container in the file(this is the address that received the Mining reward if the node succeeds to mine the block)

2. Remember if you stopped containers and you want to set up a network again first remove these stopped containers

```bash
docker container prune
```

3. for sending transactions to nodes use [wallet](https://github.com/alikarimi999/walle) app

## Command Line

1. For starting the network first use the following command to run a node and create a genesis block

```bash
shitcoin newchain -port=5000 -miner <address> -dbpath <database_path>
```

### NOTE

in blockchain apps like bitcoin or Ethereum genesis block is hardcoded but in this app each time you want to set up a network you have to mine a new genesis block

2. after the first node mined genesis block now you can run new nodes and connect them to the first node by the following command.

```bash
$ shitcoin connect -address http://localhost:5000  -port=5001 -miner <address> 
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

shitcoin is available under the MIT license. See the [LICENSE](https://github.com/alikarimi999/shitcoin/blob/master/LICENSE) file for more info.
