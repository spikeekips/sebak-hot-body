# sebak-hot-body

This is stress testing tool for SEBAK.

## Build

```
$ go build
```

## Usage

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
