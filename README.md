# sebak-hot-body

This is stress testing tool for SEBAK.

## Build

```
$ go build
```

## Run `hot-body`

```
$ ./sebak-hot-body go -h
Run hot-body

Usage:
  ./sebak-hot-body go <secret seed> [flags]

Flags:
      --concurrent int            number of transactions, they will be sent concurrently (default 10)
      --confirm-duration string   duration for checking transaction confirmed (default "60s")
  -h, --help                      help for go
      --log string                set log file
      --log-format string         log format, {terminal, json} (default "terminal")
      --log-level string          log level, {crit, error, warn, info, debug} (default "info")
      --request-timeout string    timeout for requests (default "30s")
      --result-output string      result output file (default "/Users/spikeekips/sebak-hot-body/result-20181021182023.log")
      --sebak string              sebak endpoint (default "http://127.0.0.1:12345")
      --timeout string            timeout for running (default "1m")
```

You already know the secret seed of one SEBAK account, `SCQ67SHPVLG6AQ3CP2JRM5GJVO5FX3S7GYZSGQPN3DLTT7P4VR3ZF6HN`, `hot-body` will create the testing accounts and send payment to them from this account. 

```
$ ./sebak-hot-body go \
    -confirm-duration 5m \
    -concurrent 300 \
    SCQ67SHPVLG6AQ3CP2JRM5GJVO5FX3S7GYZSGQPN3DLTT7P4VR3ZF6HN \
```

This will `300` requests continueously for `5` minutes.


## Getting Result

```
$ ./sebak-hot-body result -h
Parse result

Usage:
  ./sebak-hot-body result <result output> [flags]

Flags:
  -h, --help                help for result
      --log string          set log file (default "/Users/spikeekips/workspace/blockchainos/sebak/src/github.com/spikeekips/sebak-hot-body/hot-body-20181022133423.log")
      --log-format string   log format, {terminal, json} (default "terminal")
      --log-level string    log level, {crit, error, warn, info, debug} (default "info")
```

```
$ ./sebak-hot-body result hot-body-result-20181022133321.log
+--------------+----------------------+---------------------------------+
| * config     |         testing time |                           1m0s  |
|              |  concurrent requests |                            200  |
|              |      initial account |  GDMBBEFF63J3K...P3R7FNPOBPCOM  |
|              |      request timeout |                           1m0s  |
|              |     confirm duration |                           1m0s  |
|              |           operations |                            100  |
+--------------+----------------------+---------------------------------+
| * network    |           network id |             sebak-test-network  |
|              |      initial balance |           10000000000000000000  |
|              |           block time |                            10s  |
|              |         base reserve |                        1000000  |
|              |             base fee |                          10000  |
+--------------+----------------------+---------------------------------+
| * node       |             endpoint |         http://localhost:12345  |
|              |              address |  GCPQRIR6PGZEW...XC64U7DURAJDB  |
|              |                state |                        BOOTING  |
|              |         block height |                              2  |
|              |           block hash |  GV6djNAvsBK8A...6VQvwuBFgdoth  |
|              |       block totaltxs |                              1  |
+--------------+----------------------+---------------------------------+
| * result     |           # requests |                            201  |
|              |          error rates |               3.98010％ (8/201) |
|              |     max elapsed time |                  64.1662369370  |
|              |     min elapsed time |                   8.2988145110  |
|              |                  OPS |                 321.6666666667  |
+--------------+----------------------+---------------------------------+
| * error      |              unknown |                 8 | 100.00000％ |
+--------------+----------------------+---------------------------------+
```
