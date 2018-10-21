# sebak-hot-body

This is stress testing tool for SEBAK.

## Build

```
$ go build github.com/spikeekips/sebak-hot-body/
```

## Usage

```
$ ./sebak-hot-body -h
Usage: ./sebak-hot-body <secret seed>

  -confirm-duration string
    	duration for checking transaction confirmed (default "60s")
  -log string
    	set log file
  -log-format string
    	log format, {terminal, json} (default "terminal")
  -log-level string
    	log level, {crit, error, warn, info, debug} (default "info")
  -request-timeout string
    	timeout for requests (default "30s")
  -result-output string
    	result output file (default "./result-20181021-173534.481562000+0900.log")
  -sebak string
    	sebak endpoint (default "http://127.0.0.1:12345")
  -t int
    	number of transactions, they will be sent concurrently (default 10)
  -timeout string
    	timeout for running (default "1m")
```

You already know the secret seed of one SEBAK account, `SCQ67SHPVLG6AQ3CP2JRM5GJVO5FX3S7GYZSGQPN3DLTT7P4VR3ZF6HN`, `hot-body` will create the testing accounts and send payment to them from this account. 

```
$ ./sebak-hot-body \
    -confirm-duration 5m \
    -t 300 \
    SCQ67SHPVLG6AQ3CP2JRM5GJVO5FX3S7GYZSGQPN3DLTT7P4VR3ZF6HN \
```

This will `300` requests continueously for `5` minutes.
