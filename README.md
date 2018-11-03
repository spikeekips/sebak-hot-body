# sebak-hot-body

This is stress testing tool for SEBAK.

## Build

```
$ go build
```

## Run `hot-body`

```
$ sebak-hot-body go -h
Run hot-body

Usage:
  sebak-hot-body go <secret seed> [flags]

Flags:
      --concurrent int            number of transactions, they will be sent concurrently (default 10)
      --confirm-duration string   duration for checking transaction confirmed (default "60s")
  -h, --help                      help for go
      --log string                set log file (default "./hot-body-20181103143943.log")
      --log-format string         log format, {terminal, json} (default "terminal")
      --log-level string          log level, {crit, error, warn, info, debug} (default "info")
      --operations int            number of operations in one transaction (default 1)
      --request-timeout string    timeout for requests (default "30s")
      --result-output string      result output file (default "./hot-body-result-20181103143943.log")
      --sebak string              sebak endpoint (default "http://127.0.0.1:12345")
      --timeout string            timeout for running (default "1m")
```

You already know the secret seed of one SEBAK account, `SCQ67SHPVLG6AQ3CP2JRM5GJVO5FX3S7GYZSGQPN3DLTT7P4VR3ZF6HN`, `hot-body` will create the testing accounts and send payment to them from this account. 

```
$ ./sebak-hot-body go \
    --concurrent 300 \
    --timeout 10m \
    SCQ67SHPVLG6AQ3CP2JRM5GJVO5FX3S7GYZSGQPN3DLTT7P4VR3ZF6HN \
```

This will `300` requests continueously for `10` minutes. This will produce the `hot-body` log and `hot-body-result` log.


## Getting Result

```
$ ./sebak-hot-body result -h
Parse result

Usage:
  ./sebak-hot-body result <result log> [flags]

Flags:
  -h, --help                help for result
      --log string          set log file (default "./hot-body-20181022133423.log")
      --log-format string   log format, {terminal, json} (default "terminal")
      --log-level string    log level, {crit, error, warn, info, debug} (default "info")
```

```
$ ./sebak-hot-body result hot-body-result-20181022133321.log
+---------------+----------------------+---------------------------------+
| * config      |         testing time |                           3m0s  |
|               |  concurrent requests |                            100  |
|               |      initial account |  GAPYEQH7MC5SG...H5G3KDRJ66IQQ  |
|               |      request timeout |                            30s  |
|               |     confirm duration |                           1m0s  |
|               |           operations |                             50  |
+---------------+----------------------+---------------------------------+
| * network     |           network id |             test sebak-network  |
|               |      initial balance |            1000000000000000000  |
|               |           block time |                             5s  |
|               |         base reserve |                        1000000  |
|               |             base fee |                          10000  |
+---------------+----------------------+---------------------------------+
| * node        |             endpoint |    https://172.31.25.219:12001  |
|               |              address |  GAQXW2KFRUCC7...IZTILTT4AN3XW  |
|               |                state |                      CONSENSUS  |
|               |         block height |                            382  |
|               |           block hash |  8vryhacYGGRqc...tw2xrSNfxeXpG  |
|               |       block totaltxs |                          10777  |
|               |       block totalops |                         520616  |
+---------------+----------------------+---------------------------------+
| * time        |              started |  2018-11-04T16:36:35.275133000  |
|               |                ended |  2018-11-04T16:39:42.161060000  |
|               |        total elapsed |                    3m6.885927s  |
+---------------+----------------------+---------------------------------+
| * result      |           # requests |                           3239  |
|               |         # operations |                         161950  |
|               |          error rates |              0.00000％ (0/3239) |
|               |     max elapsed time |                  30.0496222300  |
|               |     min elapsed time |                   0.6617497240  |
|               |         distribution |                                 |
|               |                      |        0-5 : 58.47484％ /  1894 |
|               |                      |        5-10: 35.53566％ /  1151 |
|               |                      |       10-15:  4.87805％ /   158 |
|               |                      |       15-20:  0.80272％ /    26 |
|               |                      |       20-25:  0.24699％ /     8 |
|               |                      |       25-30:  0.03087％ /     1 |
|               |                      |       30-35:  0.03087％ /     1 |
|               |                      |       35-40:  0.00000％ /     0 |
|               |         expected OPS |                            866  |
|               |             real OPS |                            866  |
+---------------+----------------------+---------------------------------+
| * error       |             no error |                                 |
+---------------+----------------------+---------------------------------+
| * sebak-error |      sebak-error-133 |                56 |  70.00000％ |
|               |      sebak-error-176 |                24 |  30.00000％ |
+---------------+----------------------+---------------------------------+
```
