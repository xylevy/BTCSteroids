package steroids

import (
	op "BTCSteroids/operations"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const bitcoinChainAPIUrl  = "https://api-r.bitcoinchain.com/v1/address/%s"
const bitcoinAPIBatchLimit = 25

type BalanceResponse []struct {
	Address                      string  `json:"address"`
	Balance                      float64 `json:"balance"`
	Hash160                      string  `json:"hash_160"`
	TotalRec                     float64 `json:"total_rec"`
	Transactions                 int     `json:"transactions"`
	UnconfirmedTransactionsCount int     `json:"unconfirmed_transactions_count"`
}

type BitcoinChain struct {
	W
}


// NewBlockchainInfo returns an initialized BlockchainInfo worker
func NewBitcoinChain() *BitcoinChain {
	return &BitcoinChain{
		W{
			name:  "bitcoinchain",
			input: make(chan Request),
		},
	}
}

func (bcc *BitcoinChain) Start() {
	requests := make([]Request, 0, bitcoinAPIBatchLimit)
	for {
		// we wait upto 5 seconds to gather as many addresses (upto query limit)
		ticker := time.NewTicker(5 * time.Second)
		select {
		case request := <-bcc.input:
			requests = append(requests, request)
			if len(requests) == bitcoinAPIBatchLimit {
				bcc.process(requests)
				requests = []Request{}
			}
		case <-ticker.C:
			if len(requests) > 0 {
				bcc.process(requests)
				requests = []Request{}
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func (bcc *BitcoinChain) process(requests []Request) {
	addresses := make([]string, 0, len(requests))
	addrToChan := make(map[string]chan Result)
	for _, req := range requests {
		addresses = append(addresses, req.Address)
		addrToChan[req.Address] = req.Output
	}
	resp, err := bcc.do(addresses)
	if err != nil {
		fmt.Println(bcc.Name()+":", err)
		go submitRequests(requests, bcc.input) // return input channel for processing
		return
	}

	var returnedAdd []string
	for _,c:= range resp{
		if c.Address != ""{
			returnedAdd = append(returnedAdd,c.Address)
		}
	}
	returnedRes:=len(returnedAdd)
	queriedAddr := len(addresses)

	if returnedRes<queriedAddr{
		if returnedRes == 0 {
			for _, add := range addresses {
				h:= Result{
					Source:             bcc.Name(),
					Address:            add,
					BalanceConfirmed:   float64(0),
					BalanceUnconfirmed: float64(0),
					BalanceTotal:       float64(0),
				}
				go func(p string) {
					addrToChan[p] <- h
				}(add)

			}
		} else {
			//fmt.Println(op.Unique(append(returnedAdd, addresses...)))
			for _, zeroAdd := range op.Unique(append(returnedAdd, addresses...)){
				h := Result{
					Source:             bcc.Name(),
					Address:            zeroAdd,
					BalanceConfirmed:   float64(0),
					BalanceUnconfirmed: float64(0),
					BalanceTotal:       float64(0),
				}
				go func(p string) {
					addrToChan[p] <- h
				}(zeroAdd)

			}

		}
	}


	for _, p := range resp {
		h := Result{
			Source:             bcc.Name(),
			Address:            p.Address,
			BalanceConfirmed:   p.TotalRec,
			BalanceUnconfirmed: float64(p.UnconfirmedTransactionsCount),
			BalanceTotal:       p.Balance,
		}
		go func(p string) {
			addrToChan[p] <- h
		}(p.Address)
	}
}

func (bcc *BitcoinChain) do(addresses []string) (BalanceResponse, error) {
	addrs := strings.Join(addresses, ",")

	url:=fmt.Sprintf(bitcoinChainAPIUrl, addrs)

	resp, err := http.Get(url)

	//fmt.Println(url)

	if err != nil {
		return BalanceResponse{}, err
	}
	defer resp.Body.Close()


	if resp.StatusCode != 200 {


		return BalanceResponse{},
			fmt.Errorf("error response from bitcoinchain, got status code: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil{
		return BalanceResponse{},
			fmt.Errorf("error reading response: %v",err)
	}

	reader := ioutil.NopCloser(bytes.NewReader(data)) // ReadCloser

	//reader := bytes.NewReader(data) // Alternative

	//log.Println(s.PrintResponseText(data),url) // Log response

	var result BalanceResponse
	err = json.NewDecoder(reader).Decode(&result)

	//result.append(result)

	return result, err
}
