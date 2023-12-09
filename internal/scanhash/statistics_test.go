package scanhash_test

import (
	"testing"
	"time"

	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/scanhash"
)

type fakeHash struct {}

func (fakeHash) Write([]byte) (int, error) {
	return 0, nil
}

func (fakeHash) Sum(b []byte) []byte {
	<-time.After(10 * time.Millisecond)
	return []byte{}
}

func (fakeHash) Reset() {}

func (fakeHash) Size() int {
	return 0
}

func (fakeHash) BlockSize() int {
	return 0
}

func TestMeteredScan(t *testing.T) {
	m := scanhash.MeteredScanHash{
		ScanHash: scanhash.ScanHash{Hash: fakeHash{}},
	}
	out := make(chan scanhash.MeteredResult)
	stop := m.MeteredScan("", []byte{}, out)
	res := <-out
	if res.HashRate > 100 || res.HashRate < 95 {
		t.Fatalf("Expected hashrate about 100, got %v", res.HashRate)
	}
	close(stop)
}
