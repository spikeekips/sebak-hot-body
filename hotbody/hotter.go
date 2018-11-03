package hotbody

import (
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"sync"
	"syscall"
	"time"

	logging "github.com/inconshreveable/log15"
	"github.com/stellar/go/keypair"

	"boscoin.io/sebak/lib/common"
	"boscoin.io/sebak/lib/errors"
	"boscoin.io/sebak/lib/network"
	"boscoin.io/sebak/lib/node"
	"boscoin.io/sebak/lib/node/runner/api"
	"boscoin.io/sebak/lib/transaction"
	"boscoin.io/sebak/lib/transaction/operation"
)

var (
	nullLogger logging.Logger
)

func init() {
	nullLogger = logging.New()
	nullLogger.SetHandler(
		logging.FuncHandler(func(*logging.Record) error {
			return nil
		}),
	)
}

type HotterConfig struct {
	Node            node.NodeInfo `json:"node"`
	T               int           `json:"t"`
	KP              *keypair.Full `json:"-"`
	InitAccount     string        `json:"init-account"`
	Timeout         time.Duration `json:"timeout"`
	RequestTimeout  time.Duration `json:"request-timeout"`
	ConfirmDuration time.Duration `json:"confirm-duration"`
	ResultOutput    string        `json:"result-output"`
	Operations      int           `json:"operations"`
}

func (r HotterConfig) GetTime() time.Time {
	return time.Time{}
}

func (r HotterConfig) GetType() string {
	return "config"
}

func (r HotterConfig) GetElapsed() int64 {
	return 0
}

func (r HotterConfig) GetRawError() map[string]interface{} {
	return map[string]interface{}{}
}

func (r HotterConfig) GetError() error {
	return nil
}

func (r HotterConfig) GetErrorType() RecordErrorType {
	return RecordErrorUnknown
}

func (r HotterConfig) Serialize() ([]byte, error) {
	return common.JSONMarshalIndent(r)
}

type Hotter struct {
	sync.RWMutex
	HotterConfig

	result          *Result
	client          *HTTP2Client
	keys            map[string]*keypair.Full
	createdAccounts []string
	runningAccounts *RunningAccounts
	cachedAddresses map[string][]string
	run             chan string
}

func NewHotter(
	config HotterConfig,
	client *HTTP2Client,
) (hotter *Hotter, err error) {
	hotter = &Hotter{
		HotterConfig: config,
		client:       client,
		keys: map[string]*keypair.Full{
			config.KP.Address(): config.KP,
		},
		run: make(chan string),
	}

	hotter.result, err = NewResult(config)

	return
}

func (h *Hotter) Start() (err error) {
	log.Debug("hotter started")

	var initAccount BlockAccount
	if initAccount, err = h.GetAccount(h.KP.Address(), false); err != nil {
		return
	}

	log.Debug("init account found", "account", initAccount)
	if initAccount.Balance < 1 {
		err = fmt.Errorf("init account does not have enough balance: %v", initAccount.Balance)
		return
	}

	numberOfAccounts := int(math.Max(float64(h.T), float64(h.Operations))) + 1

	n := numberOfAccounts / h.Node.Policy.OperationsLimit
	if numberOfAccounts%h.Node.Policy.OperationsLimit > 0 {
		n += 1
	}
	for i := 0; i < n; i++ {
		l := h.Node.Policy.OperationsLimit
		if (i+1)*h.Node.Policy.OperationsLimit > numberOfAccounts {
			l = numberOfAccounts % h.Node.Policy.OperationsLimit
		}

		var targets []string
		for j := 0; j < l; j++ {
			k := h.NewKeypair()
			targets = append(targets, k.Address())
		}
		if err = h.createAccounts(h.KP, h.Node.Policy.BaseReserve*100, targets...); err != nil {
			return
		}
		h.createdAccounts = append(h.createdAccounts, targets...)
		log.Debug("created accounts", "count", len(targets))
	}

	h.runningAccounts = &RunningAccounts{}

	log.Debug("cached accounts")
	h.cachedAddresses = map[string][]string{}
	for _, address := range h.createdAccounts {
		for _, otherAddress := range h.createdAccounts {
			if address == otherAddress {
				continue
			}
			h.cachedAddresses[address] = append(h.cachedAddresses[address], otherAddress)
		}
	}

	log.Debug("created all accounts", "count", len(h.keys)-1)

	log.Debug("start to make SEBAK to be hotter and hotter")

	stopChan := make(chan bool)

	go func() {
		var startStop bool
		for {
			select {
			case <-stopChan:
				startStop = true
			case address := <-h.run:
				if startStop {
					return
				}
				go func(a string) {
					if h.runningAccounts.IsActive(address) {
						return
					}

					h.runningAccounts.SetActive(a)
					defer func(b string) {
						h.runningAccounts.SetDeactive(b)
					}(a)

					log_ := log.New(logging.Ctx{"m": "request", "address": A(a)})
					log_.Debug("start request", "running", h.runningAccounts.Len())

					if err := h.request(a); err != nil {
						if _, ok := err.(*ErrorStopRunning); ok {
							log_.Debug("stop request", "address", a, "reason", err)
							return
						}
						log_.Error("request failed", "address", a, "error", err)
					}
					log_.Debug("end", "running", h.runningAccounts.Len())

					go func() {
						h.run <- a
					}()
				}(address)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
				log.Debug("actives", "running", h.runningAccounts.Len())
				time.Sleep(1 * time.Second)
			}
		}
	}()

	h.result.Write("started")
	for _, address := range h.createdAccounts[:h.T] {
		h.run <- address
	}

	select {
	case <-time.After(h.Timeout):
		log.Debug("will be stopped; waiting for the existing requests closing", "timeout", h.Timeout)

		stopChan <- true
	}

	for {
		if h.runningAccounts.Len() != 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		log.Debug("will be stopped", "running", h.runningAccounts.Len())
		break
	}

	h.result.Write("ended")

	//close(h.run)
	close(stopChan)
	h.result.Close()

	return
}

func (h *Hotter) NewKeypair() *keypair.Full {
	k, _ := keypair.Random()

	h.Lock()
	defer h.Unlock()

	h.keys[k.Address()] = k

	return k
}

func (h *Hotter) GetAccount(address string, ignoreLog bool) (ac BlockAccount, err error) {
	var log_ logging.Logger
	if ignoreLog {
		log_ = nullLogger
	} else {
		log_ = log.New(logging.Ctx{"m": "GetAccount", "uid": common.GenerateUUID(), "address": A(address)})
	}

	url := fmt.Sprintf("%s/%s/accounts/%s", network.UrlPathPrefixAPI, api.APIVersionV1, address)
	log_.Debug("starting", "url", url)

	var b []byte
	for i := 0; i < 3; i++ {
		if b, err = h.client.Get(url, nil); err != nil {
			if !ignoreLog {
				log_.Error("failed", "error", err)
			}
			continue
		}
		err = nil
		break
	}
	if err != nil {
		log_.Error("failed to get account", "error", err)
		return
	}

	if ac, err = NewAccountFromJSON(b); err != nil {
		log_.Error("failed NewAccountFromJSON()", "error", err)
		return
	}

	log_.Debug("success", "account", ac)
	return
}

func (h *Hotter) GetTransaction(hash string, ignoreLog bool) (ctx Transaction, err error) {
	var log_ logging.Logger
	if ignoreLog {
		log_ = nullLogger
	} else {
		log_ = log.New(logging.Ctx{"m": "GetTransaction", "uid": common.GenerateUUID(), "hash": hash})
	}

	url := fmt.Sprintf("%s/%s/transactions/%s", network.UrlPathPrefixAPI, api.APIVersionV1, hash)
	log_.Debug("starting", "url", url)

	var b []byte
	if b, err = h.client.Get(url, nil); err != nil {
		if !ignoreLog {
			log_.Error("failed", "error", err)
		}
		return
	}

	if ctx, err = NewTransactionFromJSON(b); err != nil {
		log_.Error("failed to NewTransactionFromJSON()", "error", err)
		return
	}

	log_.Debug("success", "transaction", ctx)
	return
}

func (h *Hotter) createAccounts(sourceKP *keypair.Full, amount common.Amount, targets ...string) (err error) {
	log_ := log.New(logging.Ctx{
		"m":   "create-accounts",
		"uid": common.GenerateUUID(),
	})

	defer func(l logging.Logger) {
		log_.Debug(
			"done",
			"error", err,
		)
	}(log_)

	log_.Debug(
		"starting",
		"source", A(sourceKP.Address()),
		"target", AA(targets),
		"amount", amount,
	)

	if amount < h.Node.Policy.BaseReserve {
		err = fmt.Errorf("insufficient amount for create-account: %v", amount)
		log_.Error(err.Error())
		return
	}

	var ac BlockAccount
	if ac, err = h.GetAccount(sourceKP.Address(), true); err != nil {
		log_.Error(err.Error())
		return
	}
	sequenceID := ac.SequenceID

	var ops []operation.Operation
	for _, target := range targets {
		op, _ := operation.NewOperation(operation.CreateAccount{
			Target: target,
			Amount: amount,
		})
		ops = append(ops, op)
	}

	var tx transaction.Transaction
	if tx, err = transaction.NewTransaction(sourceKP.Address(), sequenceID, ops...); err != nil {
		log_.Error(err.Error())
		return
	}

	tx.Sign(sourceKP, []byte(h.Node.Policy.NetworkID))
	log_.Debug("transaction created", "transaction", tx.GetHash())

	defer func(t time.Time, l logging.Logger) {
		h.result.Write(
			"create-accounts",
			"elapsed", ElapsedTime(t),
			"count", len(targets),
			"addresses", targets,
			"error", err,
		)
	}(time.Now(), log_)

	if err = h.sendTransaction(tx); err != nil {
		log_.Error("failed to send transaction", "error", err)
		return
	}

	// check transaction is stored in block
	var ctx Transaction
	for {
		if ctx, err = h.GetTransaction(tx.GetHash(), true); err == nil {
			break
		}
		err = nil
		time.Sleep(time.Duration(300) * time.Millisecond)
	}

	log_.Debug(
		"transaction confirmed",
		"confirmed transaction", ctx,
	)

	return
}

func (h *Hotter) payment(sourceKP *keypair.Full, amount common.Amount, targets ...string) (err error) {
	log_ := log.New(logging.Ctx{"m": "payment", "uid": common.GenerateUUID()})

	defer func(l logging.Logger) {
		log_.Debug(
			"done",
			"error", err,
		)
	}(log_)

	log_.Debug(
		"starting",
		"source", A(sourceKP.Address()),
		"target", AA(targets),
		"amount", amount,
	)

	if amount < 0 {
		err = errors.OperationAmountUnderflow
		return
	}

	var ac BlockAccount
	if ac, err = h.GetAccount(sourceKP.Address(), true); err != nil {
		log_.Error(err.Error())
		return
	}
	sequenceID := ac.SequenceID

	var ops []operation.Operation
	for _, target := range targets {
		op, _ := operation.NewOperation(operation.Payment{
			Target: target,
			Amount: amount,
		})
		ops = append(ops, op)
	}

	var tx transaction.Transaction
	if tx, err = transaction.NewTransaction(sourceKP.Address(), sequenceID, ops...); err != nil {
		log_.Error(err.Error())
		return
	}

	tx.Sign(sourceKP, []byte(h.Node.Policy.NetworkID))
	log_.Debug("transaction created", "transaction", tx.GetHash())

	defer func(t time.Time, l logging.Logger) {
		h.result.Write(
			"payment",
			"elapsed", ElapsedTime(t),
			"count", len(targets),
			"addresses", targets,
			"amount", amount,
			"source", sourceKP.Address(),
			"error", err,
		)
	}(time.Now(), log_)

	if err = h.sendTransaction(tx); err != nil {
		log_.Error("failed to send transaction", "error", err)
		return
	}

	// check transaction is stored in block
	done := make(chan Transaction)

	go func() {
		var ctx Transaction
		for {
			if ctx, err = h.GetTransaction(tx.GetHash(), true); err == nil {
				err = nil
				done <- ctx
				return
			}
			time.Sleep(time.Duration(300) * time.Millisecond)
		}
	}()

	select {
	case ctx := <-done:
		log_.Debug(
			"payment transaction confirmed",
			"confirmed transaction", ctx,
		)
	case <-time.After(h.ConfirmDuration):
		err = fmt.Errorf("timeout: %v", h.ConfirmDuration)
		log_.Error(
			"payment transaction failed to confirm",
			"error", "timeout",
			"timeout", h.ConfirmDuration,
		)
	}

	return
}

func (h *Hotter) sendTransaction(tx transaction.Transaction) (err error) {
	log_ := log.New(logging.Ctx{"m": "sendTransaction", "uid": common.GenerateUUID()})

	var b []byte

	var body []byte
	if body, err = tx.Serialize(); err != nil {
		return
	}

	retries := 3
	for i := 0; i < 3; i++ { // retry
		b, err = h.client.Post(network.UrlPathPrefixNode+"/message", body, nil)

		if err != nil {
			if i == retries-1 {
				break
			}

			log_e := log_.New(logging.Ctx{"error": err, "error-type": fmt.Sprintf("%T", err)})
			switch terr := err.(type) {
			case *url.Error:
				if nerr, ok := terr.Err.(net.Error); ok && nerr.Timeout() {
					log_e.Debug("timeout will be ignored", "try", i)
					err = nil
					continue
				}

				if nerr, ok := terr.Err.(*net.OpError); ok {
					log_e.Debug("found net.OpError", "details", nerr)
					if scerr, ok := nerr.Err.(*os.SyscallError); ok {
						log_e.Debug("found SyscallError", "type", scerr.Err)
						if scerr.Err == syscall.ECONNRESET {
							log_e.Debug("syscall.ECONNRESET will be ignored", "try", i)
							err = nil
							continue
						}
					}
				}
				log_e.Error("found *url.Error")
			case net.Error:
				if terr.Timeout() {
					log_e.Debug("timeout will be ignored", "try", i)
					err = nil
					continue
				}
				log_e.Error("found net.Error")
			}
		}
		break
	}

	if err != nil {
		log_.Error("response", "body", string(b), "error", err, "error-type", fmt.Sprintf("%T", err))
	} else {
		log_.Debug("response", "body", string(b))
	}

	return
}

func (h *Hotter) request(address string) (err error) {
	account, _ := h.GetAccount(address, true)

	var addresses []string
	for {
		//addresses = PickKeysRandom(h.createdAccounts, h.Operations, address)
		//addresses = PickKeysRandom(h.createdAccounts, 1, address)
		addresses = PickKeysRandom2(h.cachedAddresses[address], h.Operations)
		if len(addresses) > 0 {
			break
		}
	}

	var targets []string
	for _, address := range addresses {
		targets = append(targets, address)
	}

	requiredBalance := (common.Amount(h.Node.Policy.BaseFee) * common.Amount(len(targets))) + (h.Node.Policy.BaseFee * common.Amount(len(targets)))
	if account.Balance < requiredBalance {
		err = NewErrorStopRunning(
			"insufficient balance: balance=%v required=%v",
			account.Balance,
			requiredBalance,
		)
		return
	}

	err = h.payment(h.keys[address], common.Amount(1), targets...)

	return
}
