package electrum

type basicResp struct {
	Result string `json:"result"`
}

// GetFeeResp represents the response to GetFee().
type GetFeeResp struct {
	Result float32 `json:"result"`
}

// GetFee returns the estimated transaction fee per kilobytes for a transaction
// to be confirmed within a target number of blocks.
// https://electrumx.readthedocs.io/en/latest/protocol-methods.html#blockchain-estimatefee
func (s *Server) GetFee(target uint32) (float32, error) {
	var resp GetFeeResp

	err := s.Request("blockchain.estimatefee", []interface{}{target}, &resp)
	if err != nil {
		return -1, err
	}

	return resp.Result, err
}

// GetRelayFee returns the minimum fee a transaction must pay to be accepted into the
// remote server memory pool.
// https://electrumx.readthedocs.io/en/latest/protocol-methods.html#blockchain-relayfee
func (s *Server) GetRelayFee() (float32, error) {
	var resp GetFeeResp

	err := s.Request("blockchain.relayfee", []interface{}{}, &resp)
	if err != nil {
		return -1, err
	}

	return resp.Result, err
}

// GetFeeHistogramResp represents the response to GetFee().
type getFeeHistogramResp struct {
	Result [][2]uint64 `json:"result"`
}

// GetFeeHistogram returns a histogram of the fee rates paid by transactions in the
// memory pool, weighted by transacation size.
// https://electrumx.readthedocs.io/en/latest/protocol-methods.html#mempool-get-fee-histogram
func (s *Server) GetFeeHistogram() (map[uint32]uint64, error) {
	var resp getFeeHistogramResp

	err := s.Request("mempool.get_fee_histogram", []interface{}{}, &resp)
	if err != nil {
		return nil, err
	}

	feeMap := make(map[uint32]uint64)
	for i := 0; i < len(resp.Result); i++ {
		feeMap[uint32(resp.Result[i][0])] = resp.Result[i][1]
	}

	return feeMap, err
}