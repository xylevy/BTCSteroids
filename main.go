
package main

import (
	hammer "BTCSteroids/steroids"
	"fmt"
	"sync"
)


func main() {
	// Init
	output := make(chan hammer.Result)

	h := hammer.New(hammer.WorkersAll())

	//Submit work
	var wg sync.WaitGroup
	for _, addr := range hammer.SampleAddresses6 {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			output <- h.GetBalance(addr)
		}(addr)
	}

	// close the output at the right time
	go func() {
		wg.Wait()
		close(output)

	}()

	for res := range output {
		// Only print addresses with a balance
		if res.BalanceTotal > 0 {
			fmt.Printf("%+v\n", res)
		}
	}

}

