package pool

import (
	"encoding/hex"
	"errors"
	"io"
	"slices"
	"strings"
	"time"

	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal"
	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/stratum"
)

type Client struct {
	r io.Reader
	w io.Writer
	Work
	Job stratum.StratumJobParams
	ValidShares int
	InvalidShares int
	Start *time.Time
	Shares []string
}

var resultOK = stratum.StratumResult{
	Status: "OK",
}

func rejectSubmission(id int, message string) (stratum.StratumSubmitResponse, error) {
	e := stratum.JsonRpcError{
		Code: -1, // XXX: no one seems to care what this is
		Message: message,
	}
	return stratum.StratumSubmitResponse{
		JsonRpcResponse: stratum.JsonRpcResponse{
			JsonRpcMessage: stratum.JsonRpcMessage{
				Version: "2.0",
				Id: id,
			},
			Error: &e,
		},
	}, errors.New(message)
}

func (c Client) ValidateSubmission(s stratum.StratumSubmitRequest) (stratum.StratumSubmitResponse, error) {
	if s.Params.JobId != c.Job.JobId {
		// TODO be more generous about stales
		return rejectSubmission(s.Id, "Block expired")
	}

	i, found := slices.BinarySearch(c.Shares, s.Params.NOnce)

	if found {
		return rejectSubmission(s.Id, "Duplicate share")
	}

	if !strings.HasPrefix(s.Params.Result, c.Job.Target) {
		return rejectSubmission(s.Id, "Invalid share")
	}

	h := internal.NameToHash(c.Job.Algo)

	if h == nil {
		panic("hash unsupported but used")
	}

	(*h).Write([]byte(s.Params.NOnce))
	hash := (*h).Sum(nil)
	if s.Params.Result != hex.EncodeToString(hash) {
		return rejectSubmission(s.Id, "Invalid share")
	}

	c.Shares = slices.Insert(c.Shares, i, s.Params.NOnce)

	if s.Params.Result == c.Work.Hash {
		// TODO solution found
	}

	return stratum.StratumSubmitResponse{
		JsonRpcResponse: stratum.JsonRpcResponse{
			JsonRpcMessage: stratum.JsonRpcMessage{
				Version: "2.0",
				Id: s.Id,
			},
		},
		Result: &resultOK,
	}, nil
}
