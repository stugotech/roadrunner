# roadrunner 

Roadrunner is well known for foiling Wyle E. Coyote's violent plans that he constructs with challenges he bought from the ACME corporation.  This tools solves ACME challenges.

![Roadrunner](https://upload.wikimedia.org/wikipedia/en/e/ee/Roadrunner_looney_tunes.png)

## Use

    Usage:
      roadrunner [command]

    Available Commands:
      help        Help about any command
      serve       Runs a server which solves ACME challenges

    Flags:
          --config string            config file
      -l, --listen string            interface to listen for challenges on (default "0.0.0.0:8080")
          --pathPrefix string        first component of URI path to challenges (default ".well-known/acme-challenge")
          --store string             KV store to use [etcd|consul|boltdb|zookeeper] (default "etcd")
          --storeNodes stringSlice   comma-seperated list of KV (URI authority only) (default [127.0.0.1:2379])
          --storePrefix string       prefix to use when looking up values in KV store (will look in "challenges" sub path) (default "roadrunner")

## Example

Start the server:

    roadrunner --store=etcd --store-nodes=127.0.0.1:2379 --store-prefix=roadrunner \
        --listen=":8080" --path-prefix=".well-known/acme-challenge" serve &

Set a value in etcd: 

    etcdctl set /roadrunner/challenges/fred flintstone

Test the retrieval:

    curl http://127.0.0.1:8080/.well-known/acme-challenge/fred

The server should respond with:

    flintstone