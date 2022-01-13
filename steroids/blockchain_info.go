package steroids

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//https://blockchain.info/balance?active=19e6eBCzeKwGtQ1D3SyM7pmTsUeYca4sSG

const blockChainAPIUrl  = "https://blockchain.info/balance?active=%s"
const blockChainAPIBatchLimit = 45


// Balances structure of the response to the balance request.
type Balances map[string]Balance

// Balance describes the available data at the same address
// when you request a balance.
type Balance struct {
	FinalBalance  uint64 `json:"final_balance"`
	NTx           uint64 `json:"n_tx"`
	TotalReceived uint64 `json:"total_received"`
}

type BlockchainInfo struct {
	W
}


// NewBlockchainInfo returns an initialized BlockchainInfo worker
func NewBlockchainInfo() *BlockchainInfo {
	return &BlockchainInfo{
		W{
			name:  "blockchaininfo",
			input: make(chan Request),
		},
	}
}

func (bci *BlockchainInfo) Start() {
	requests := make([]Request, 0, blockChainAPIBatchLimit)
	for {
		// we wait upto 5 seconds to gather as many addresses (upto query limit)
		ticker := time.NewTicker(5 * time.Second)
		select {
		case request := <-bci.input:
			requests = append(requests, request)
			if len(requests) == blockChainAPIBatchLimit {
				bci.process(requests)
				requests = []Request{}
			}
		case <-ticker.C:
			if len(requests) > 0 {
				bci.process(requests)
				requests = []Request{}
			}
		}

		time.Sleep(3 * time.Second)
	}
}

func (bci *BlockchainInfo) process(requests []Request) {
	addresses := make([]string, 0, len(requests))
	addrToChan := make(map[string]chan Result)
	for _, req := range requests {
		addresses = append(addresses, req.Address)
		addrToChan[req.Address] = req.Output
	}
	resp, err := bci.do(addresses)
	if err != nil {
		fmt.Println(bci.Name()+":", err)
		go submitRequests(requests, bci.input) // return input channel for processing
		return
	}

	//fmt.Println(resp)
	for k, v := range resp {
		h := Result{
			Source:             bci.Name(),
			Address:            k,
			BalanceConfirmed:   float64(v.TotalReceived),
			BalanceUnconfirmed: float64(0),
			BalanceTotal:       float64(v.FinalBalance),
		}
		go func(p string) {
			addrToChan[p] <- h
		}(k)
	}
}

func (bci *BlockchainInfo) do(addresses []string) (Balances, error) {
	addrs := strings.Join(addresses, "|")

	url:=fmt.Sprintf(blockChainAPIUrl, addrs)

	resp, err := http.Get(url)

	//fmt.Println(url)

	if err != nil {
		return Balances{}, err
	}
	//defer resp.Body.Close()

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()


	if resp.StatusCode != 200 {


		return Balances{},
			fmt.Errorf("error response from blockchaininfo, got status code: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil{
		return Balances{},
			fmt.Errorf("error reading response: %v",err)
	}

	reader := ioutil.NopCloser(bytes.NewReader(data)) // ReadCloser

	//reader := bytes.NewReader(data) // Alternative

	//log.Println(s.PrintResponseText(data),url) // Log response

	var result Balances
	err = json.NewDecoder(reader).Decode(&result)

	//result.append(result)

	return result, err
}
