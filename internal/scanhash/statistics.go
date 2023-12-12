package scanhash

import (
	"time"
)

type MeteredResult struct {
	ScanResult
	HashRate float64
}

func (s ScanHash) MeteredScan(template string, prefix []byte, out chan MeteredResult) chan struct{} {
	hit := make(chan ScanResult)

	start := time.Now().UnixNano()
	stop := s.Scan(template, prefix, hit)

	go func() {
		for s := range hit {
			end := time.Now().UnixNano()
			out <- MeteredResult{
				ScanResult: s,
				HashRate: float64(s.Iterations) / float64(end - start) * 1e9,
			}
		}
		close(out)
	}()
	return stop
}
