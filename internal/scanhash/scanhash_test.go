package scanhash_test

import (
	"crypto/md5"
	"testing"

	"github.com/xdavidwu/password-miner/internal/scanhash"
)

func TestIterateString(t *testing.T) {
	out := make(chan string)
	stop := scanhash.IterateString("00}", out)
	for _, i := range []string{"00}", "00~", "01 "} {
		v := <-out
		if v != i {
			t.Fatalf("Expected %v, got %v", i, v)
		}
	}
	close(stop)
	out = make(chan string)
	stop = scanhash.IterateString("~~~", out)
	for _, i := range []string{"~~~", "    "} {
		v := <-out
		if v != i {
			t.Fatalf("Expected %v, got %v", i, v)
		}
	}
	close(stop)
}

func TestScanHash(t *testing.T) {
	s := scanhash.ScanHash{Hash: md5.New()}
	// https://hashcat.net/wiki/doku.php?id=example_hashes
	// hashcat: 8743b52063cd84097a65d1633f5c74f5
	out := make(chan scanhash.ScanResult)
	stop := s.Scan("hashca ", "8743b52063cd84097a65d1633f5c74", out)
	res := <-out
	if res.Solution != "hashcat" {
		t.Fatalf("Scanner skipped expected solution to: %v", res.Solution)
	}
	if res.Iterations != 't' - ' ' + 1 {
		t.Fatalf("Scanner report unexpected iterations: %v", res.Iterations)
	}
	close(stop)
}
