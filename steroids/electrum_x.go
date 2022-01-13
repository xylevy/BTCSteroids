package steroids

import (
	"BTCSteroids/electrum"
	s "BTCSteroids/services"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)


var ElectrumXNodes = []string{
	"electrum.emzy.de:50001",
	"electrum.blockstream.info:50001",
	"korea.electrum-server.com:50001",
	"de.poiuty.com:50001",
	"stavver.dyshek.org:50001",

}

const electrumxBatchLimit = 5 // You can increase this number to collect more addresses for this worker

var hasher =sha256.New()

type GetBalanceResp struct {
	Result GetBalanceResult `json:"result"`
}

// GetBalanceResult represents the content of the result field in the response to GetBalance().
type GetBalanceResult struct {
	Confirmed   float64 `json:"confirmed"`
	Unconfirmed float64 `json:"unconfirmed"`
}

type ElectrumX struct {
	W
	//*electrum.Server

}

// NewElectrumX returns an initialized ElectrumX worker
func NewElectrumX() *ElectrumX {
	return &ElectrumX{
		W{
			name:  "electrumx",
			input: make(chan Request),
		},
	}

}


func (ex *ElectrumX) Start() {
	requests := make([]Request, 0, electrumxBatchLimit)
	for {
		// we wait upto 5 seconds to gather as many addresses (upto query limit)
		ticker := time.NewTicker(5 * time.Second)
		select {
		case request := <-ex.input:
			requests = append(requests, request)
			if len(requests) == electrumxBatchLimit {
				ex.process(requests)
				requests = []Request{}
			}
		case <-ticker.C:
			if len(requests) > 0 {
				ex.process(requests)
				requests = []Request{}
			}
		}
	}
}

func (ex *ElectrumX) process(requests []Request) {
	addresses := make([]string, 0, len(requests))
	addrToChan := make(map[string]chan Result)
	for _, req := range requests {
		addresses = append(addresses, req.Address)
		addrToChan[req.Address] = req.Output
	}

	var wg sync.WaitGroup

	wg.Add(len(addresses))

	seed := rand.NewSource(time.Now().Unix())
	r := rand.New(seed) // initialize local pseudorandom generator

	nodeAddress:=ElectrumXNodes[r.Intn(len(ElectrumXNodes))]

	source := strings.Split(nodeAddress,":")[0]


	go func(nodeAddress string) {
		defer wg.Done()
		node := electrum.NewServer()
		if err := node.ConnectTCP(nodeAddress); err != nil {
			//log.Fatal(err,nodeAddress)
		}
		go func() {
			for {
				if err := node.Ping(); err != nil {
					//log.Fatal(err)
				}
				time.Sleep(5 * time.Second)
			}
		}()

		//fmt.Println(addresses,nodeAddress)

		for _, address := range addresses {

			addrScriptHash := s.ScripthashAddr(address,hasher)

			resp, err := ex.do(addrScriptHash,node)


			if err != nil {
				fmt.Println(ex.Name()+":"+nodeAddress," ->",err)
				go submitRequests(requests, ex.input) // return input channel for reprocessing
				return
			}

			h := Result{
				Source:             ex.Name()+ ":"+ source,
				Address:            address,
				BalanceConfirmed:   resp.Result.Confirmed,
				BalanceUnconfirmed: resp.Result.Unconfirmed,
				BalanceTotal:       resp.Result.Confirmed + resp.Result.Unconfirmed,
			}
			go func(address string) {
				addrToChan[address] <- h
			}(address)


		}

	}(nodeAddress)




}

func (ex *ElectrumX) do(scripthash string, x *electrum.Server) (GetBalanceResp, error) {
	var resp GetBalanceResp

	err := x.Request("blockchain.scripthash.get_balance", []interface{}{scripthash}, &resp)


	if err != nil {
		return GetBalanceResp{}, err
	}

	//log.Println(resp.Result)

	return resp, err
}

