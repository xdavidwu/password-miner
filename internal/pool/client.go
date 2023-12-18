package pool

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"slices"
	"strings"
	"time"

	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal"
	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/stratum"
	"github.com/dustin/go-humanize"
)

type Client struct {
	R *bufio.Reader
	W io.Writer
	Id string
	JobIdInt int
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
const DifficultyAdjustIntSec = 30
// TODO should be per-algo
// TODO support initial difficulty selection (login benchmark info)
const InitiallyAssumedHashRate = 1e6

var resultOK = stratum.StratumResult{
	Status: "OK",
}

func (c Client) HashRate() float64 {
	return c.EstHashes / (float64(c.LastSub.UnixNano() - c.Start.UnixNano()) / 1e9)
}

// XXX we should make our target more flexible, maybe into bits
func (c Client) DifficultyTargetForHashRate(hr float64) (float64, string) {
	l := int(math.Log2(hr * DesiredSubmissionIntSec) / 4)
	return math.Pow(16, float64(l)), c.Work.Hash[:l]
}

func (c Client) DifficultyTarget() (float64, string) {
	return c.DifficultyTargetForHashRate(c.HashRate())
}

func (c Client) Blob() string {
	h := internal.NameToHash(c.Work.Algo)
	sz := int(math.Ceil(float64(h.Size() * 8) / math.Log2(96)))
	return fmt.Sprintf("%s-%d-", c.Id, c.JobIdInt) + strings.Repeat(" ", sz)
}

func (c *Client) ResetAlgo() {
	c.Shares = []string{}
	t := time.Now()
	c.Start = &t
	c.LastSub = nil
	c.EstHashes = 0
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

func (c *Client) ValidateSubmission(s stratum.StratumSubmitRequest) (stratum.StratumSubmitResponse, error) {
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

	h.Write([]byte(s.Params.NOnce))
	hash := h.Sum(nil)
	if s.Params.Result != hex.EncodeToString(hash) {
		return rejectSubmission(s.Id, "Invalid share")
	}

	c.Shares = slices.Insert(c.Shares, i, s.Params.NOnce)
	c.ValidShares += 1
	c.EstHashes += c.Difficulty
	t := time.Now()
	c.LastSub = &t

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

func (c *Client) newJob(t string) {
	c.JobIdInt += 1
	c.Job = stratum.StratumJobParams{
		JobId: fmt.Sprintf("%v", c.JobIdInt),
		Id: c.Id,
		Blob: c.Blob(),
		Algo: c.Work.Algo,
		Target: t,
	}
}

func (c *Client) handleStratum(
		submissions chan stratum.StratumSubmitRequest,
		results chan stratum.StratumSubmitResponse,
		jobs chan stratum.StratumJobParams,
		works chan Work, sol chan string) chan struct{} {
	stop := make(chan struct{})
	hr := InitiallyAssumedHashRate
	diffAdj := make(<-chan time.Time)
	go func() {
		for {
			select {
			case s := <-submissions:
				// TODO banning/abort connection on client misbehavior
				res, err := c.ValidateSubmission(s)
				if err == nil {
					hr = c.HashRate()
					log.Printf("Client(%v): %v", c.Id, humanize.SIWithDigits(hr, 3, "H/s"))
					if strings.HasPrefix(s.Params.Result, c.Work.Hash) {
						sol <- s.Params.NOnce
						log.Printf("Solution found for %v by %v: \"%v\"\n", c.Work.Hash, c.Id, s.Params.NOnce)
					}
				}
				results <- res
			case w := <-works:
				if w.Algo != c.Work.Algo {
					c.ResetAlgo()
					// XXX: should do invidual hr estimation
				}
				c.Work = w
				d, t := c.DifficultyTargetForHashRate(hr)
				c.Difficulty = d
				c.newJob(t)
				diffAdj = time.After(DifficultyAdjustIntSec * time.Second)
				log.Printf("Client(%v): new job via new work at difficulty %g", c.Id, d)
				jobs <- c.Job
			case <-diffAdj:
				if c.EstHashes == 0 { // never submitted, no estimation
					hr /= 256
					d, t := c.DifficultyTargetForHashRate(hr)
					c.Difficulty = d
					c.newJob(t)
					diffAdj = time.After(DifficultyAdjustIntSec * time.Second)
					log.Printf("Client(%v): new job via blind scaling at difficulty %g", c.Id, d)
					jobs <- c.Job
				} else {
					d, t := c.DifficultyTargetForHashRate(hr)
					if d != c.Difficulty {
						c.Difficulty = d
						c.newJob(t)
						diffAdj = time.After(DifficultyAdjustIntSec * time.Second)
						log.Printf("Client(%v): new job via scaling at difficulty %g", c.Id, d)
						jobs <- c.Job
					}
				}
			case <-stop:
				close(results)
				close(jobs)
				return
			}
		}
	}()
	return stop
}

func (c *Client) Control(works chan Work, solutions chan string) {
	submissions := make(chan stratum.StratumSubmitRequest)
	results := make(chan stratum.StratumSubmitResponse)
	jobs := make(chan stratum.StratumJobParams)
	stop := c.handleStratum(submissions, results, jobs, works, solutions)

	go func() {
		for {
			b, err := c.R.ReadBytes(byte('\n'))
			if err != nil {
				close(stop)
				return
			}
			s := stratum.StratumSubmitRequest{}
			err = json.Unmarshal(b, &s)
			if err != nil {
				log.Printf("Client(%v): unrecognized message\n", c.Id)
			}
			submissions <- s
		}
	}()

	for {
		select {
		case r, ok := <-results:
			if !ok {
				return
			}
			o, err := json.Marshal(r)
			if err != nil {
				panic(err)
			}
			_, err = c.W.Write(append(o, byte('\n')))
			if err != nil {
				close(stop)
				return
			}
		case j, ok := <-jobs:
			if !ok {
				return
			}
			job := stratum.StratumJobNotif{
				JsonRpcRequest: stratum.JsonRpcRequest{
					JsonRpcMessage: stratum.JsonRpcMessage{
						Version: "2.0",
					},
					Method: "job",
				},
				Params: j,
			}
			o, err := json.Marshal(job)
			if err != nil {
				panic(err)
			}
			_, err = c.W.Write(append(o, byte('\n')))
			if err != nil {
				close(stop)
				return
			}
		}
	}
}
