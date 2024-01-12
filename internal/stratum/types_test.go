package stratum_test

import (
	"encoding/json"
	"testing"

	"github.com/xdavidwu/password-miner/internal/stratum"
)

type valueAssert struct {
	value any
	truth any
	name string
}

func assertValues(t *testing.T, vs []valueAssert) {
	for _, v := range vs {
		if v.value != v.truth {
			t.Fatalf("Unexpected %v: %v, expected: %v", v.name, v.value, v.truth)
		}
	}
}

// real world protocol dump tests

func TestUnmarshalStratumLoginRequest(t *testing.T) {
	data := []byte(`{"method": "login", "params": {"login": "44sWgFsoaVuSTHaH8LnRzvLQXvisq1bwc7zj8nkd3CZMZ227aHqDYN8Hw1pk1i38hYJtUKnZJHpqSiXnjht8BF4qLe32Xom", "pass": "x", "algo": ["cn/r", "cn/half", "cn-pico/trtl"], "algo-perf": {"cn/r": 17.897772, "cn/half": 101.873193, "cn-pico/trtl": 418.607085}}, "id": 1}`)
	s := stratum.StratumLoginRequest{}
	if e := json.Unmarshal(data, &s); e != nil {
		t.Fatalf("Fail to unmarshal: %v", e)
	}
	assertValues(t, []valueAssert{
		{s.Method, "login", "method"},
		{s.Id, 1, "id"},
		{s.Params.Login, "44sWgFsoaVuSTHaH8LnRzvLQXvisq1bwc7zj8nkd3CZMZ227aHqDYN8Hw1pk1i38hYJtUKnZJHpqSiXnjht8BF4qLe32Xom", "params.login"},
		{s.Params.Pass, "x", "params.pass"},
	})
}

func TestUnmarshalStratumLoginResponse(t *testing.T) {
	data := []byte(`{"jsonrpc":"2.0","id":1,"error":null,"result":{"id":"32768461","job":{"blob":"0808aab6d5ab064fc276ca88bbbff7c7340352a876dc1a8aeac1dd0aa5b2caf1f2d5c792787f1400000000e2d08acf309725be96963f165ac08d107befabf6ae09c76c2440e0bb96d5d72f010000000000000000000000000000000000000000000000000000000000000000","algo":"cn/half","height":2890063,"job_id":"32768462","target":"f0f8c301","id":"32768461"},"status":"OK"}}`)
	s := stratum.StratumLoginResponse{}
	if e := json.Unmarshal(data, &s); e != nil {
		t.Fatalf("Fail to unmarshal: %v", e)
	}
	var nilError *stratum.JsonRpcError = nil
	assertValues(t, []valueAssert{
		{s.Id, 1, "id"},
		{s.Error, nilError, "error"},
		{s.Result.Status, "OK", "result.status"},
		{s.Result.Job.Algo, "cn/half", "result.job.algo"},
		{s.Result.Job.Blob, "0808aab6d5ab064fc276ca88bbbff7c7340352a876dc1a8aeac1dd0aa5b2caf1f2d5c792787f1400000000e2d08acf309725be96963f165ac08d107befabf6ae09c76c2440e0bb96d5d72f010000000000000000000000000000000000000000000000000000000000000000", "result.job.blob"},
		{s.Result.Job.Target, "f0f8c301", "result.job.target"},
		{s.Result.Job.JobId, "32768462", "result.job.job_id"},
		{s.Result.Job.Id, "32768461", "result.job.id"},
	})
}

func TestUnmarshalStratumSubmitRequest(t *testing.T) {
	data := []byte(`{"method": "submit", "params": {"id": "32768461", "job_id": "32768462", "nonce": "3c0c0000", "result": "92b6470f0cb7fec9633640ff8fb5a515dd8dea2519e0b2d2029818b20428de00"}, "id":1}`)
	s := stratum.StratumSubmitRequest{}
	if e := json.Unmarshal(data, &s); e != nil {
		t.Fatalf("Fail to unmarshal: %v", e)
	}
	assertValues(t, []valueAssert{
		{s.Method, "submit", "method"},
		{s.Id, 1, "id"},
		{s.Params.Id, "32768461", "params.id"},
		{s.Params.JobId, "32768462", "params.job_id"},
		{s.Params.NOnce, "3c0c0000", "params.nonce"},
		{s.Params.Result, "92b6470f0cb7fec9633640ff8fb5a515dd8dea2519e0b2d2029818b20428de00", "params.result"},
	})
}

func TestUnmarshalStratumSubmitResponse(t *testing.T) {
	data := []byte(`{"jsonrpc":"2.0","id":1,"error":null,"result":{"status":"OK"}}`)
	s := stratum.StratumSubmitResponse{}
	if e := json.Unmarshal(data, &s); e != nil {
		t.Fatalf("Fail to unmarshal: %v", e)
	}
	var nilError *stratum.JsonRpcError = nil
	assertValues(t, []valueAssert{
		{s.Id, 1, "id"},
		{s.Error, nilError, "error"},
		{s.Result.Status, "OK", "result.status"},
	})
}

func TestUnmarshalStratumJobNotif(t *testing.T) {
	data := []byte(`{"method":"job","params":{"blob":"0808aab6d5ab064fc276ca88bbbff7c7340352a876dc1a8aeac1dd0aa5b2caf1f2d5c792787f1400000000d7018193991e9c90c84817365d593e2a3e6f9c54533969908438c728b0060088010000000000000000000000000000000000000000000000000000000000000000","algo":"cn/half","height":2890063,"job_id":"32771451","target":"bfbd1100","id":"32768461"},"jsonrpc":"2.0"}`)
	s := stratum.StratumJobNotif{}
	if e := json.Unmarshal(data, &s); e != nil {
		t.Fatalf("Fail to unmarshal: %v", e)
	}
	assertValues(t, []valueAssert{
		{s.Method, "job", "method"},
		{s.Id, 0, "id"},
		{s.Params.Algo, "cn/half", "params.algo"},
		{s.Params.Blob, "0808aab6d5ab064fc276ca88bbbff7c7340352a876dc1a8aeac1dd0aa5b2caf1f2d5c792787f1400000000d7018193991e9c90c84817365d593e2a3e6f9c54533969908438c728b0060088010000000000000000000000000000000000000000000000000000000000000000", "params.blob"},
		{s.Params.Target, "bfbd1100", "params.target"},
		{s.Params.JobId, "32771451", "params.job_id"},
		{s.Params.Id, "32768461", "params.id"},
	})
}
