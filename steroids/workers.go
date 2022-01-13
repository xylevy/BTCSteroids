package steroids

// TODO add new workers
// Result holds the balance of each address
// All balances in Satoshi
type Result struct {
	Source, Address    string
	BalanceTotal       float64
	BalanceConfirmed   float64
	BalanceUnconfirmed float64
}

// Request holds a balance request
type Request struct {
	Address string
	Output  chan Result
}

// Worker defines the methods a balance querying worker must implement
type Worker interface {
	// Name of the worker
	Name() string

	// Start worker processing
	Start()

	// Get balance of address using this worker
	GetBalance(string) Result

	// Set the input channel of worker
	setInput(chan Request)
}

// W contains the fields that nearly all workers need
// Also implements some of the Worker interface
type W struct {
	name  string
	input chan Request
}

// Name of the worker
func (w W) Name() string {
	return w.name
}

// GetBalance queries and returns balance of given address
func (w W) GetBalance(addr string) Result {
	output := make(chan Result)
	defer close(output)
	req := Request{
		Address: addr,
		Output:  output,
	}
	w.input <- req
	return <-output
}

// SetInput sets the input channel of the worker
func (w *W) setInput(input chan Request) {
	w.input = input
}

// Helpers

// WorkersAll returns slice containing all workers
func WorkersAll() []Worker {
	return []Worker{
		NewBlockonomics(),
		NewBlockcypher(),
		NewSmartbit(),
		NewElectrumX(),
		NewBlockchainInfo(),
		NewBitcoinChain(),
		//NewLocalChecker(),
	}
}

// SubmitAddresses submits addresses to a channel
// Usually called in a goroutine
func submitRequests(requests []Request, ch chan<- Request) {
	for _, rq := range requests {
		ch <- rq
	}
}
