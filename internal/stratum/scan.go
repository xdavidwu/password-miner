package stratum

import (
	"encoding/hex"
	"math"
	"log"

	"github.com/xdavidwu/password-miner/internal"
	"github.com/xdavidwu/password-miner/internal/scanhash"
	"github.com/dustin/go-humanize"
)

func SwitchedScan(in chan StratumJobParams, out chan StratumSubmitParams) chan struct{} {
	stop := make(chan struct{})
	var s *scanhash.ScanHash = nil
	scannerStop := make(chan struct{})
	scannerOut := make(chan scanhash.MeteredResult)
	algo := ""
	jobId := ""
	id := ""
	difficulty := float64(0)
	go func() {
		for {
			select {
			case <-stop:
				close(scannerStop)
				return
			case i := <-in:
				if i.Algo != algo {
					h := internal.NameToHash(i.Algo)
					if h == nil {
						log.Fatalf("unsupported algorithm: %v", i.Algo)
					}
					algo = i.Algo
					log.Printf("Pool switch algorithm to %v\n", algo)
					s = &scanhash.ScanHash{Hash: h}
				}
				d := math.Pow(16, float64(len(i.Target)))
				if d != difficulty {
					log.Printf("Pool set difficulty to %.4g\n", d)
					difficulty = d
				}
				close(scannerStop)
				scannerOut = make(chan scanhash.MeteredResult)
				scannerStop = s.MeteredScan(i.Blob, i.Target, scannerOut)
				jobId = i.JobId
				id = i.Id
			case sol := <- scannerOut:
				log.Printf("Solution found, %v, %v\n", algo, humanize.SIWithDigits(sol.HashRate, 3, "H/s"))
				out <- StratumSubmitParams{
					JobId: jobId,
					Id: id,
					NOnce: sol.Solution,
					Result: hex.EncodeToString(sol.Hash),
				}
			}
		}
	}()
	return stop
}
