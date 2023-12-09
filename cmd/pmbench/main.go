package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/dustin/go-humanize"
	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/scanhash"
)

func main() {
	for _, algo := range []struct{
		name string
		impl scanhash.MeteredScanHash
		template string
		target string
	}{
		{
			name: "MD5",
			impl: scanhash.MeteredScanHash{ScanHash: scanhash.ScanHash{Hash: md5.New()}},
			template: "hase   ",
			target: "8743b52063cd84097a65d1633f5c74f5",
		},
		{
			name: "SHA1",
			impl: scanhash.MeteredScanHash{ScanHash: scanhash.ScanHash{Hash: sha1.New()}},
			template: "hase   ",
			target: "b89eaac7e61417341b710b727768294d0e6a277b",
		},
		{
			name: "SHA256",
			impl: scanhash.MeteredScanHash{ScanHash: scanhash.ScanHash{Hash: sha256.New()}},
			template: "hase   ",
			target: "127e6fbfe24a750e72930c220a8e138275656b8e5d8f48a98c3c92df2caba935",
		},
		{
			name: "SHA512",
			impl: scanhash.MeteredScanHash{ScanHash: scanhash.ScanHash{Hash: sha512.New()}},
			template: "hase   ",
			target: "82a9dda829eb7f8ffe9fbe49e45d47d2dad9664fbb7adf72492e3c81ebd3e29134d9bc12212bf83c6840f10e8246b9db54a4859b7ccd0123d86e5872c1e5082f",
		},
	}{
		t, err := hex.DecodeString(algo.target)
		if err != nil {
			panic(err)
		}
		out := make(chan scanhash.MeteredResult)
		stop := algo.impl.MeteredScan(algo.template, t, out)
		res := <-out
		fmt.Printf("%v:\t%v\n", algo.name, humanize.SIWithDigits(res.HashRate, 3, "H/s"))
		close(stop)
	}
}
