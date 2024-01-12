package stratum_test

import (
	"fmt"
	"testing"

	"github.com/xdavidwu/password-miner/internal/stratum"
)

func TestSwitchedScan(t *testing.T) {
	in := make(chan stratum.StratumJobParams)
	out := make(chan stratum.StratumSubmitParams)
	stop := stratum.SwitchedScan(in, out)

	for i, algo := range [][2]string{
		{"md5", "8743b52063cd84097a65d1633f5c74f5"},
		{"sha1", "b89eaac7e61417341b710b727768294d0e6a277b"},
	} {
		jobId := fmt.Sprintf("jobid%v", i)
		id := fmt.Sprintf("id%v", i)
		job := stratum.StratumJobParams{
			JobId: jobId,
			Id: id,
			Algo: algo[0],
			Blob: "hashca ",
			Target: algo[1][:len(algo[1]) - 2],
		}
		in <- job
		sol := <-out
		assertValues(t, []valueAssert{
			{sol.Id, id, "id"},
			{sol.JobId, jobId, "job_id"},
			{sol.NOnce, "hashcat", "nonce"},
			{sol.Result, algo[1], "result"},
		})
	}
	close(stop)
}
