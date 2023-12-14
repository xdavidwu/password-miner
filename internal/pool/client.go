package pool

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"math"
	"slices"
	"strings"
	"time"

	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal"
	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/stratum"
)

type Client struct {
	r io.Reader
	w io.Writer
	Id string
	Work
	Job stratum.StratumJobParams
	ValidShares int
	InvalidShares int
	EstHashes float64
	Difficulty float64
	LastSub *time.Time
	Start *time.Time
	Shares []string
}

const DesiredSubmissionIntSec = 10
const JobDifficultyAdjustTimeout = 60
// TODO should be per-algo
// TODO support initial difficulty selection (login benchmark info)
const InitiallyAssumedHashRate = 1e6

var resultOK = stratum.StratumResult{
	Status: "OK",
}

func (c Client) HashRate() float64 {
	return c.EstHashes / float64(c.LastSub.Unix() - c.Start.Unix())
}

// XXX we should make our target more flexible, maybe into bits
func (c Client) DifficultyTargetForHashRate(hr float64) (float64, string) {
	l := int(math.Log2(hr * DesiredSubmissionIntSec) / 8)
	return math.Pow(256, float64(l)), c.Work.Hash[:l * 2]
}

func (c Client) DifficultyTarget() (float64, string) {
	return c.DifficultyTargetForHashRate(c.HashRate())
}

func (c Client) Blob() string {
	return "" //TODO
}

func (c Client) Control() {
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
	if s.Params.JobId != c.Job.JobId || s.Params.Id != c.Id {
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

	// TODO rate limiting (low diff shares)

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
	c.ValidShares += 1
	c.EstHashes += c.Difficulty
	t := time.Now()
	c.LastSub = &t

	if s.Params.Result == c.Work.Hash {
		log.Printf("Solution found for %v by %v: \"%v\"\n", c.Work.Hash, c.Id, s.Params.NOnce)
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
