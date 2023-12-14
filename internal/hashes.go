package internal

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
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
