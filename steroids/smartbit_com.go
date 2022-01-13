package steroids


// TODO Halfway there => Logic to handle unused addresses, nothing is returned for unused address when queried with multiple others

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

const smartbitURL = "https://api.smartbit.com.au/v1/blockchain/address/%s"
const smartbitBatchLimit = 30 // upper limit number of address queries at a time is 200 TODO 200


type smartBitResponse struct {
	Success   bool        `json:"success"`
	Addresses []Addresses `json:"addresses"`
}
type Total struct {
	Received         string `json:"received"`
	ReceivedInt      int    `json:"received_int"`
	Spent            string `json:"spent"`
	SpentInt         int    `json:"spent_int"`
	Balance          string `json:"balance"`
	BalanceInt       int    `json:"balance_int"`
	InputCount       int    `json:"input_count"`
	OutputCount      int    `json:"output_count"`
	TransactionCount int    `json:"transaction_count"`
}
type Confirmed struct {
	Received         string `json:"received"`
	ReceivedInt      int    `json:"received_int"`
	Spent            string `json:"spent"`
	SpentInt         int    `json:"spent_int"`
	Balance          string `json:"balance"`
	BalanceInt       int    `json:"balance_int"`
	InputCount       int    `json:"input_count"`
	OutputCount      int    `json:"output_count"`
	TransactionCount int    `json:"transaction_count"`
}
type Unconfirmed struct {
	Received         string `json:"received"`
	ReceivedInt      int    `json:"received_int"`
	Spent            string `json:"spent"`
	SpentInt         int    `json:"spent_int"`
	Balance          string `json:"balance"`
	BalanceInt       int    `json:"balance_int"`
	InputCount       int    `json:"input_count"`
	OutputCount      int    `json:"output_count"`
	TransactionCount int    `json:"transaction_count"`
}
type MConfirmed struct {
	Balance    string `json:"balance"`
	BalanceInt int    `json:"balance_int"`
}
type MUnconfirmed struct {
	Balance    string `json:"balance"`
	BalanceInt int    `json:"balance_int"`
}
type Multisig struct {
	Confirmed   MConfirmed   `json:"confirmed"`
	Unconfirmed MUnconfirmed `json:"unconfirmed"`
}
type Addresses struct {
	Address     string      `json:"address"`
	Total       Total       `json:"total"`
	Confirmed   Confirmed   `json:"confirmed"`
	Unconfirmed Unconfirmed `json:"unconfirmed"`
	Multisig    Multisig    `json:"multisig"`
}

type smartbitNoBalErr struct {
	ErrorCode int
}

func (e *smartbitNoBalErr) Error() string {
	return fmt.Sprintf("smartbit: parse addresses failure, batch of queried wallets empty? XD status code: %v", e.ErrorCode)
}


//smartbit worker
type Smartbit struct {
	W
}

// NewSmartbit returns an initialized Smartbit worker
func NewSmartbit() *Smartbit {
	return &Smartbit{
		W{
			name:  "smartbit",
			input: make(chan Request),
		},
	}
}

func (sb *Smartbit) Start() {
	requests := make([]Request, 0, smartbitBatchLimit)
	for {
		// we wait upto 5 seconds to gather as many addresses (upto query limit)
		ticker := time.NewTicker(5 * time.Second)
		select {
		case request := <-sb.input:
			requests = append(requests, request)
			if len(requests) == smartbitBatchLimit {
				sb.process(requests)
				requests = []Request{}
			}
		case <-ticker.C:
			if len(requests) > 0 {
				sb.process(requests)
				requests = []Request{}
			}
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

func (sb *Smartbit) process(requests []Request) {
	addresses := make([]string, 0, len(requests))
	addrToChan := make(map[string]chan Result)
	for _, req := range requests {
		addresses = append(addresses, req.Address)
		addrToChan[req.Address] = req.Output
	}
	resp, err := sb.do(addresses)
	if err != nil {
		re, err := err.(*smartbitNoBalErr)
		if err {
			fmt.Println(re.Error())

		} else {
			fmt.Println(sb.Name()+":", err)
			go submitRequests(requests, sb.input) // return input channel for processing
			return
		}

	}

	//fmt.Printf("%v %v",resp.Addresses,addresses)
	//fmt.Println(len(resp.Addresses))

	returnedRes:=len(resp.Addresses)
	queriedAddr := len(addresses)


	if returnedRes<queriedAddr{
		if returnedRes == 0 {
			for _, add := range addresses {
				h:= Result{
					Source:             sb.Name(),
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
			var returnedAdd []string
			for _, add := range resp.Addresses {
				returnedAdd = append(returnedAdd,add.Address)
			}

			//fmt.Println(op.Unique(append(returnedAdd, addresses...)))

			for _, zeroAdd := range op.Unique(append(returnedAdd, addresses...)){
				h := Result{
					Source:             sb.Name(),
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


	for _, p := range resp.Addresses {
		h := Result{
			Source:             sb.Name(),
			Address:            p.Address,
			BalanceConfirmed:   float64(p.Confirmed.BalanceInt),
			BalanceUnconfirmed: float64(p.Unconfirmed.BalanceInt),
			BalanceTotal:       float64(p.Total.BalanceInt),
		}
		go func(p Addresses) {
			addrToChan[p.Address] <- h
		}(p)
	}
}

func (sb *Smartbit) do(addresses []string) (smartBitResponse, error) {
	addrs := strings.Join(addresses, ",")

	url:=fmt.Sprintf(smartbitURL , addrs)

	resp, err := http.Get(url)

	if err != nil {
		return smartBitResponse{}, err
	}
	defer resp.Body.Close()


	if resp.StatusCode != 200 {

		if resp.StatusCode == 400{
			return smartBitResponse{},&smartbitNoBalErr{ErrorCode:resp.StatusCode}
			//return smartBitResponse{},fmt.Errorf("error response from smart bit, queried wallets empty? XD")
		}

		return smartBitResponse{},
			fmt.Errorf("error response from smartbit, got status code: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil{
		return smartBitResponse{},
		fmt.Errorf("error reading response: %v",err)
	}

	reader := ioutil.NopCloser(bytes.NewReader(data)) // ReadCloser

	//reader := bytes.NewReader(data) // Alternative

	//log.Println(s.PrintResponseText(data),url) // Log response

	var result smartBitResponse
	err = json.NewDecoder(reader).Decode(&result)
	
	return result, err
}
