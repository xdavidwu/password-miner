package scanhash

import (
	"bytes"
	"encoding/hex"
	"hash"
)

type ScanResult struct {
	Solution string
	Hash []byte
	Iterations int
}

type ScanHash struct {
	hash.Hash
}

// iterate in ascii printable range: 0x20~0x7e, starting from template
func IterateString(template string, out chan string) chan struct{} {
	stop := make(chan struct{})
	t := []rune(template)
	go func() {
		for {
			select {
			case out <- string(t):
			addrune := true
			for i := len(t) - 1; i >= 0; i -= 1 {
				if t[i] + 1 <= 0x7e {
					t[i] += 1
					addrune = false
					break
				}
				t[i] = 0x20
			}
			if addrune {
				t = append([]rune{0x20}, t...)
			}
			case <-stop:
				close(out)
				return
			}
		}
	}()
	return stop
}

func HasPrefix(b []byte, prefix string) bool {
	if len(prefix) % 2 == 0 {
		bp, err := hex.DecodeString(prefix)
		if err != nil {
			panic(err)
		}
		return bytes.HasPrefix(b, bp)
	}
	bp, err := hex.DecodeString(prefix[:len(prefix) - 1])
	if err != nil {
		panic(err)
	}
	if !bytes.HasPrefix(b, bp) {
		return false
	}
	bp, err = hex.DecodeString(prefix[len(prefix) - 1:] + "0")
	if err != nil {
		panic(err)
	}
	return b[len(prefix) / 2] & 0xf0 == bp[0]
}

func (s ScanHash) Scan(template string, prefix string, out chan ScanResult) chan struct{} {
	candidates := make(chan string)
	stop := IterateString(template, candidates)
	iterations := 0
	go func() {
		for c := range candidates {
			iterations += 1
			s.Hash.Reset()
			s.Hash.Write([]byte(c))
			hash := s.Hash.Sum(nil)
			if HasPrefix(hash, prefix) {
				out <- ScanResult{
					Solution: c,
					Hash: hash,
					Iterations: iterations,
				}
			}
		}
		close(out)
	}()
	return stop
}
