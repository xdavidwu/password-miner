# password-miner

Distributed brute-force preimage attacks based on pool-mining

## Requirements

Go 1.21 or newer

## Usage

`pmbench` is a benchmark utility that takes no arguments

`pmd` is the pool daemon, meant for running on a single dedicated control node

`pmc` is the client, runs on computation nodes

```
Usage of pmd:
  -address string
    	Address to listen on (default "0.0.0.0:1234")
  -hashes string
    	File containing list of hases, formatted in lines of
    	<type> <hash>
    	Where hash is in hex, and may be a prefix for quick demonstration
    	Supported types: md5, sha1, sha224, sha256, sha384, sha512
    	 (default "hashes")
```

```
Usage of pmc:
  -address string
    	Address of the pool (default "127.0.0.1:1234")
  -password string
    	Password (default "password")
  -username string
    	Username (default "username")
```

Authentication is not implemented yet, any username or password will work

## Glossary

### Pool

The controlling node that delivers jobs to computation nodes, and find the actual preimage from solutions to jobs

### Client

The computation nodes, like miners in pool-mining

### Work

A hash to solve

### Target

A prefix of Work, with length determined by Difficulty

### Job

A Target and a search space, for Client to search on

### Difficulty

Expected hashes to find a solution to Job. Pool adjusts Difficulty (via length of Target) of each Client that the Client find solutions to Jobs at a desired rate, to low enough that Pool can stably monitor inferred performance and effort of Client from statistics, but also high enough that there are few solutions to find actual preimage from.
