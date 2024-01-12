package main

import (
	"encoding/hex"
	"fmt"

	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal"
	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/scanhash"
	"github.com/dustin/go-humanize"
)

func main() {
	for _, algo := range internal.SupportList {
		out := make(chan scanhash.MeteredResult)
		h := algo.F()
		h.Write([]byte("hashcat"))
		target := hex.EncodeToString(h.Sum([]byte{}))
		scan := scanhash.ScanHash{Hash: h}
		stop := scan.MeteredScan("hasf   ", target, out)
		res := <-out
		fmt.Printf("%v:\t%v\n", algo.Name, humanize.SIWithDigits(res.HashRate, 3, "H/s"))
		close(stop)
	}
}
