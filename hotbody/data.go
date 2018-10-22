package hotbody

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	"boscoin.io/sebak/lib/common"
)

func A(address string) string {
	return fmt.Sprintf("%s.%s", address[:4], address[len(address)-4:])
}

func AA(addresses []string) []string {
	var a []string
	for _, address := range addresses {
		a = append(a, A(address))
	}

	return a
}

type BlockAccount struct {
	Address    string        `json:"address"`
	Balance    common.Amount `json:""`
	SequenceID uint64        `json:"sequence_id"`
	/*
		Linked          string        `json:"linked"`
		OperationsURL   string        `json:"operations-url"`
		TransactionsURL string        `json:"transactions-url"`
	*/
}

func NewAccountFromJSON(b []byte) (ac BlockAccount, err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	if err = json.Unmarshal(b, &ac); err != nil {
		return
	}
	/*
		ac.OperationsURL = m["_links"].(map[string]interface{})["operations"].(map[string]interface{})["href"].(string)
		ac.TransactionsURL = m["_links"].(map[string]interface{})["transactions"].(map[string]interface{})["href"].(string)
	*/

	return
}

func (ac BlockAccount) Serialize() ([]byte, error) {
	return json.Marshal(ac)
}

func (ac BlockAccount) String() string {
	s, _ := common.JSONMarshalIndent(ac)
	return string(s)
}

type Transaction struct {
	Created        string        `json:"created"`
	Fee            common.Amount `json:"fee"`
	Hash           string        `json:"hash"`
	OperationCount uint64        `json:"operation_count"`
	SequenceID     uint64        `json:"sequenceid"`
	Source         string        `json:"source"`
	/*
		OperationsURL  string        `json:"operations-url"`
		AccountURL     string        `json:"account-url"`
	*/
}

func NewTransactionFromJSON(b []byte) (ctx Transaction, err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	if err = json.Unmarshal(b, &ctx); err != nil {
		return
	}
	/*
		ctx.OperationsURL = m["_links"].(map[string]interface{})["operations"].(map[string]interface{})["href"].(string)
		ctx.AccountURL = m["_links"].(map[string]interface{})["account"].(map[string]interface{})["href"].(string)
	*/

	return
}

func (ctx Transaction) Serialize() ([]byte, error) {
	return json.Marshal(ctx)
}

func (ctx Transaction) String() string {
	s, _ := common.JSONMarshalIndent(ctx)
	return string(s)
}

type SortByBlockAccountBalance []BlockAccount

func (s SortByBlockAccountBalance) Len() int {
	return len(s)
}
func (s SortByBlockAccountBalance) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortByBlockAccountBalance) Less(i, j int) bool {
	return s[i].Balance < s[j].Balance
}

type RunningAccounts struct {
	sync.Map
}

func (r *RunningAccounts) Len() int {
	var count int
	r.Range(func(_ interface{}, v interface{}) bool {
		if v.(bool) == true {
			count++
		}
		return true
	})

	return count
}

func (r *RunningAccounts) IsActive(address string) bool {
	v, ok := r.Load(address)
	return ok && v.(bool) == true
}

func (r *RunningAccounts) SetActive(address string) {
	r.Store(address, true)
}

func (r *RunningAccounts) SetDeactive(address string) {
	r.Store(address, false)
}

func PickKeysRandom(addresses []string, n int, excludes ...string) []string {
	find := func(a []int, n int) bool {
		for _, i := range a {
			if i == n {
				return true
			}
		}

		return false
	}

	l := len(addresses)
	var indices []int
	for i := 0; i < n; i++ {
		var j int
		for {
			j = rand.Intn(l)
			if len(indices) < 1 || !find(indices, j) {
				break
			}
		}

		indices = append(indices, j)
	}

	var found int
	var founds []string
	for i, address := range addresses {
		if len(indices[found:]) < 1 {
			break
		}

		if !find(indices[found:], i) {
			continue
		}

		if _, ok := common.InStringArray(excludes, address); ok {
			continue
		}
		founds = append(founds, address)
		found++
	}

	return founds
}

func PickKeysRandom2(addresses []string, n int) []string {
	if len(addresses) > n {
		shuffled := make([]string, len(addresses))
		for i, v := range rand.Perm(len(addresses)) {
			if i >= n {
				break
			}

			shuffled[i] = addresses[v]
		}
		return shuffled[:n]
	}

	return addresses
}

type Record interface {
	GetType() string
	GetElapsed() int64
	GetError() error
	GetErrorType() RecordErrorType
	GetRawError() map[string]interface{}
}

type RecordErrorType string

const (
	RecordErrorUnknown    RecordErrorType = "unknown"
	RecordErrorECONNRESET RecordErrorType = "reset-by-peer"
)

/*
{
    "addresses": [
        "GBXUE3BYSLTDG74BJSPQKRIBK2KBFJLJVFOZP5LT6VO262UXCXPQMY5V"
    ],
    "count": 300,
    "elapsed": "1.9459782410",
    "error": null,
    "type": "create-accounts"
}
*/
type RecordCreateAccounts struct {
	Type      string                 `json:"type"`
	Addresses []string               `json:"addresses"`
	Count     uint64                 `json:"count"`
	Elapsed   string                 `json:"elapsed"`
	Error     map[string]interface{} `json:"error"`
}

func (r RecordCreateAccounts) GetType() string {
	return r.Type
}

func (r RecordCreateAccounts) GetElapsed() int64 {
	p, _ := ParseRecordElapsedTime(r.Elapsed)
	return p
}

func (r RecordCreateAccounts) GetRawError() map[string]interface{} {
	return r.Error
}

func (r RecordCreateAccounts) GetError() error {
	if len(r.Error) < 1 {
		return nil
	}

	return fmt.Errorf("%v", r.Error)
}

func (r RecordCreateAccounts) GetErrorType() RecordErrorType {
	return ParseRecordError(r.Error)
}

/*
{
    "addresses": [
        "GCOO5YBOIFXELMXBW5QAXQXURTDBBLXZAWQE424DIAWOYEUXG3QTFMLK"
    ],
    "amount": "1",
    "count": 1,
    "elapsed": "2.1623947930",
    "error": null,
    "source": "GDNSUHR7G5LS6WTVHQJULOTEXXCYBPNK7NXB323VEBCEY7LEJWFEEXSN",
    "type": "payment"
}
*/
type RecordPayment struct {
	Type      string                 `json:"type"`
	Addresses []string               `json:"addresses"`
	Count     uint64                 `json:"count"`
	Elapsed   string                 `json:"elapsed"`
	Error     map[string]interface{} `json:"error"`
	Amount    common.Amount          `json:"amount"`
	Source    string                 `json:"source"`
}

func (r RecordPayment) GetType() string {
	return r.Type
}

func (r RecordPayment) GetElapsed() int64 {
	p, _ := ParseRecordElapsedTime(r.Elapsed)
	return p
}

func (r RecordPayment) GetRawError() map[string]interface{} {
	return r.Error
}

func (r RecordPayment) GetError() error {
	if len(r.Error) < 1 {
		return nil
	}

	return fmt.Errorf("%v", r.Error)
}

func (r RecordPayment) GetErrorType() RecordErrorType {
	return ParseRecordError(r.Error)
}
