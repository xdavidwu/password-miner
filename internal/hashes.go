package internal

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"hash/fnv"
)

func NameToHash(name string) hash.Hash {
	switch name {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	case "sha512":
		return sha512.New()
	case "fnv32a":
		return fnv.New32()
	default:
		return nil
	}
}
