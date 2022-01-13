package steroids

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const blockonomicsURL = "https://www.blockonomics.co/api/balance"
const blockonomicsBatchLimit = 25
const blockonomicsAPIKey="8I9i201dCcSkDDqv2dOALPGYCEYedcMS52SahbkJgts"

type blockonomicsBalances struct {
	Addr        string  `json:"addr"`
	Confirmed   float64 `json:"confirmed"`
	Unconfirmed float64 `json:"unconfirmed"`
}

type blockonomicsResponse struct {
	Response []blockonomicsBalances `json:"response"`
}


// Blockonomics worker
type Blockonomics struct {
	W
}

// NewBlockonomics returns an initialized Blockonomics worker
func NewBlockonomics() *Blockonomics {
	return &Blockonomics{
		W{
			name:  "blockonomics",
			input: make(chan Request),
		},
	}
}

// Start the blockonomics worker
func (bl *Blockonomics) Start() {
	requests := make([]Request, 0, blockonomicsBatchLimit)
	for {
		// we wait upto 5 seconds to gather as many addresses (upto query limit)
		ticker := time.NewTicker(5 * time.Second)
		select {
		case request := <-bl.input:
			requests = append(requests, request)
			if len(requests) == blockonomicsBatchLimit {
				bl.process(requests)
				requests = []Request{}
			}
		case <-ticker.C:
			if len(requests) > 0 {
				bl.process(requests)
				requests = []Request{}
			}
		}
	}
}

func (bl *Blockonomics) do(addresses []string) (blockonomicsResponse, error) {
	addrs := strings.Join(addresses, " ")

	req, err := json.Marshal(map[string]string{
		"addr": addrs,
	})

	//fmt.Println(req)
	if err != nil {
		return blockonomicsResponse{}, err
	}

	client := &http.Client{}

	r, _ := http.NewRequest("POST",blockonomicsURL,
		bytes.NewBuffer(req))

	r.Header.Add("Authorization","Bearer "+blockonomicsAPIKey)
	r.Header.Add("Content-Type", "application/json")

	resp,err :=client.Do(r)

	if err != nil {
		return blockonomicsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return blockonomicsResponse{},
			fmt.Errorf("error response from blockonomics, got status code: %v", resp.StatusCode)
	}

	var result blockonomicsResponse

	//bodyBytes, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//bodyString := string(bodyBytes)
	//fmt.Println(bodyString)

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (bl *Blockonomics) process(requests []Request) {
	addresses := make([]string, 0, len(requests))
	addrToChan := make(map[string]chan Result)
	for _, req := range requests {
		addresses = append(addresses, req.Address)
		addrToChan[req.Address] = req.Output
	}
	resp, err := bl.do(addresses)
	if err != nil {
		fmt.Println(bl.Name()+":", err)
		go submitRequests(requests, bl.input) // return input channel for processing
		return
	}
	for _, p := range resp.Response {
		h := Result{
			Source:             bl.Name(),
			Address:            p.Addr,
			BalanceConfirmed:   p.Confirmed,
			BalanceUnconfirmed: p.Unconfirmed,
			BalanceTotal:       p.Confirmed + p.Unconfirmed,
		}
		go func(p blockonomicsBalances) {
			addrToChan[p.Addr] <- h
		}(p)
	}
}
