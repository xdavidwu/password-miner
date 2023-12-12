package stratum

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"log"

	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/scanhash"
	"github.com/dustin/go-humanize"
)

func NameToHash(name string) *hash.Hash {
	switch name {
	case "md5":
		h := md5.New()
		return &h
	case "sha1":
		h := sha1.New()
		return &h
	case "sha256":
		h := sha256.New()
		return &h
	case "sha512":
		h := sha512.New()
		return &h
	default:
		return nil
	}
}

func SwitchedScan(in chan StratumJobParams, out chan StratumSubmitParams) chan struct{} {
	stop := make(chan struct{})
	var s *scanhash.ScanHash = nil
	scannerStop := make(chan struct{})
	scannerOut := make(chan scanhash.MeteredResult)
	algo := ""
	jobId := ""
	id := ""
	go func() {
		for {
			select {
			case <-stop:
				close(scannerStop)
				return
			case i := <-in:
				if i.Algo != algo {
					h := NameToHash(i.Algo)
					if h == nil {
						log.Fatalf("unsupported algorithm: %v", i.Algo)
					}
					algo = i.Algo
					log.Printf("Pool switch algorithm to %v\n", algo)
					s = &scanhash.ScanHash{Hash: *h}
				}
				t, err := hex.DecodeString(i.Target)
				if err != nil {
					log.Fatalf("malformed target (%v): %v", i.Target, err)
				}
				close(scannerStop)
				scannerOut = make(chan scanhash.MeteredResult)
				scannerStop = s.MeteredScan(i.Blob, t, scannerOut)
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
