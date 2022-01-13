package steroids

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

/* TODO - Brainflayer-esque worker

*/

const httpmqLevelDBUrl=  "http://127.0.0.1:1218/?name=%s&opt=view&pos="
const batchLimit = 3800


type BalanceDump struct {
	Addr			string
	Bal     		string
}

type LocalChecker struct {
	W
	currentCount int          		// upto 16 queries per second or 3840 - 4000 requests every 4 minutes
}

func NewLocalChecker() *LocalChecker {

	ldbc :=&LocalChecker{
		W:W{
			name:  "local checker",
			input: make(chan Request),
		},
		currentCount: 0,
	}
	return ldbc
}

func (ldbc *LocalChecker) Start() {
	requests := make([]Request, 0, batchLimit)
	for {

		if ldbc.currentCount == batchLimit {
			time.Sleep(4*time.Minute)
		}

		ticker := time.NewTicker(5 * time.Second)
		select {
		case request := <-ldbc.input:
			requests = append(requests, request)
			if len(requests) == batchLimit {
				ldbc.currentCount += len(requests)
				ldbc.process(requests)
				requests = []Request{}
			}
		case <-ticker.C:
			if len(requests) > 0 {
				ldbc.process(requests)
				requests = []Request{}
				ldbc.currentCount += len(requests)
			}
		}
	}

}
func (ldbc *LocalChecker) process(requests []Request) {
	addresses := make([]string, 0, len(requests))
	addrToChan := make(map[string]chan Result)
	for _, req := range requests {
		addresses = append(addresses, req.Address)
		addrToChan[req.Address] = req.Output
	}

	var wg sync.WaitGroup

	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConnsPerHost:   batchLimit,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	cl:= &http.Client{Transport: t}

	for _, address := range addresses {
		wg.Add(1)
		go func(address string) {
			defer wg.Done()
			resp, err := ldbc.do(address,cl)

			if err != nil {
				fmt.Println(ldbc.Name()+":", err)
				go submitRequests(requests, ldbc.input)
				return
			}

			s, err := strconv.ParseFloat(resp.Bal, 64)

			if err != nil {
				fmt.Println(ldbc.Name()+":", err)
				go submitRequests(requests, ldbc.input)
				return
			}


			h := Result{
				Source:             ldbc.Name(),
				Address:            address,
				BalanceConfirmed:   0,
				BalanceUnconfirmed: 0,
				BalanceTotal:       s,
			}

			addrToChan[address] <- h


		}(address)


	}
	wg.Wait()
}


func (ldbc *LocalChecker) do(addrs string, cl *http.Client) (BalanceDump, error) {

	url:=fmt.Sprintf(httpmqLevelDBUrl, addrs)

	resp, err := cl.Get(url)

	if err != nil {
		return BalanceDump{}, err
	}
	defer resp.Body.Close()


	if resp.StatusCode != 200 {


		return BalanceDump{},
			fmt.Errorf("error response from localchecker, got status code: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil{
		return BalanceDump{},
			fmt.Errorf("error reading response: %v",err)
	}


	reader := ioutil.NopCloser(bytes.NewReader(data))

	//fmt.Println(data)

	var result BalanceDump
	err = json.NewDecoder(reader).Decode(&result)


	return result, err
}
