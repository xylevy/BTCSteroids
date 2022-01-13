package steroids

// T is the hammer object
type T struct {
	input chan Request
}

// New returns a new Hammer object
func New(workers []Worker) T {
	input := make(chan Request)
	for _, w := range workers {
		w.setInput(input)
		go w.Start()
	}
	return T{input: input}
}

// GetBalance Returns a balance Result object for the given address
func (h T) GetBalance(addr string) Result {
	output := make(chan Result)
	defer close(output)
	req := Request{
		Address: addr,
		Output:  output,
	}
	h.input <- req
	return <-output
}
