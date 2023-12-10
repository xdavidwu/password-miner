package stratum

type JsonRpcMessage struct {
	Version string `json:"jsonrpc"`
	Id	int `json:"id,omitempty"`
}

type JsonRpcRequest struct {
	JsonRpcMessage `json:",inline"`
	Method string `json:"method"`
}

type JsonRpcError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type JsonRpcResponse struct {
	JsonRpcMessage `json:",inline"`
	Error *JsonRpcError `json:"error,omitempty"`
}

type StratumJobParams struct {
	Blob string `json:"blob"`
	Algo string `json:"algo"`
	JobId string `json:"job_id"`
	// XXX: likely instance id, does it support reconnection?
	Id string `json:"id"`
	Target string `json:"target"`
}

// no id, no response
type StratumJobNotif struct {
	JsonRpcRequest `json:",inline"`
	Params StratumJobParams `json:"params"`
}

type StratumSubmitParams struct {
	JobId string `json:"job_id"`
	Id string `json:"id"`
	NOnce string `json:"nonce"`
	Result string `json:"result"`
}

type StratumSubmitRequest struct {
	JsonRpcRequest `json:",inline"`
	Params StratumSubmitParams `json:"params"`
}

type StratumResult struct {
	// XXX: why?
	Status string `json:"status,omitempty"`
}

type StratumSubmitResponse struct {
	JsonRpcResponse `json:",inline"`
	Result *StratumResult `json:"result,omitempty"`
}

type StratumLoginParams struct {
	Login string `json:"login"`
	Pass string `json:"pass"`
}

type StratumLoginRequest struct {
	JsonRpcRequest `json:",inline"`
	Params StratumLoginParams `json:"params"`
}

type StratumLoginResult struct {
	StratumResult `json:",inline"`
	Job *StratumJobParams `json:"job,omitempty"`
	Id string `json:"id"`
}

type StratumLoginResponse struct {
	JsonRpcResponse `json:",inline"`
	Result *StratumLoginResult `json:"result,omitempty"`
}
