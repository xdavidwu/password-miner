package stratum_test

import (
	"bufio"
	"encoding/json"
	"io"
	"testing"

	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/stratum"
)

func TestStratum(t *testing.T) {
	username := "username"
	password := "password"
	in, input := io.Pipe()
	o, out := io.Pipe()
	output := bufio.NewReader(o)
	go stratum.Stratum(in, out, stratum.StratumLoginParams{
		Login: username,
		Pass: password,
	})

	loginReq := stratum.StratumLoginRequest{}
	raw, err := output.ReadBytes('\n')
	if err != nil {
		t.Fatalf("cannot read login request: %v", err)
	}
	err = json.Unmarshal(raw, &loginReq)
	if err != nil {
		t.Fatalf("cannot unmarshal login request: %v", err)
	}
	assertValues(t, []valueAssert{
		{loginReq.Params.Login, username, "params.login"},
		{loginReq.Params.Pass, password, "params.pass"},
	})


	job := stratum.StratumJobParams{
		JobId: "initjob",
		Id: "init",
		Algo: "md5",
		Blob: "hashca ",
		Target: "8743b52063cd84097a65d1633f5c74",
	}

	res := stratum.StratumLoginResult{
		StratumResult: stratum.StratumResult{
			Status: "OK",
		},
		Job: &job,
	}

	loginRes, err := json.Marshal(stratum.StratumLoginResponse{
		JsonRpcResponse: stratum.JsonRpcResponse{
			JsonRpcMessage: stratum.JsonRpcMessage{
				Version: "2.0",
				Id: loginReq.Id,
			},
		},
		Result: &res,
	})
	_, err = input.Write(append(loginRes, byte('\n')))
	if err != nil {
		t.Fatalf("failed to send login response: %v", err)
	}

	submission := stratum.StratumSubmitRequest{}
	raw, err = output.ReadBytes('\n')
	if err != nil {
		t.Fatalf("cannot read job submission: %v", err)
	}
	err = json.Unmarshal(raw, &submission)
	if err != nil {
		t.Fatalf("cannot unmarshal job submission: %v", err)
	}

	assertValues(t, []valueAssert{
		{submission.Params.Id, "init", "params.id"},
		{submission.Params.JobId, "initjob", "params.job_id"},
		{submission.Params.NOnce, "hashcat", "params.nonce"},
		{submission.Params.Result, "8743b52063cd84097a65d1633f5c74f5", "params.result"},
	})

	input.Close()
}
