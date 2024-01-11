package internal

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

type H struct {
	Name string
	F func () hash.Hash
}

var SupportList = []H{
	{"md5", md5.New},
	{"sha1", sha1.New},
	{"sha224", sha256.New224},
	{"sha256", sha256.New},
	{"sha384", sha512.New384},
	{"sha512", sha512.New},
}

func NameToHash(name string) hash.Hash {
	for _, h := range SupportList {
		if name == h.Name {
			return h.F()
		}
	}
	return nil
}
