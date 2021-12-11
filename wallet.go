package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/mr-tron/base58/base58"
	"golang.org/x/crypto/ripemd160"
	"log"
)

type WalletKeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey []byte //由私钥生成
}

func NewWalletKeyPair() *WalletKeyPair{
	privateKey , err := ecdsa.GenerateKey(elliptic.P256() , rand.Reader)
	if err!=nil{
		log.Panic(err)
	}
	publicKeyRow := privateKey.PublicKey
	publicKey := append(publicKeyRow.X.Bytes() , publicKeyRow.Y.Bytes()...)
	return &WalletKeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

func (w *WalletKeyPair) GetAddress() string {
	publicHash := HashPubKey(w.PublicKey)
	version := 0x00
	//21字节的数据
	payload := append([]byte{byte(version)},publicHash...)
	checkSum := CheckSum(payload)
	payload = append(payload , checkSum...)

	address := base58.Encode(payload)
	return address
}
func HashPubKey(pubKey []byte) []byte {
	hash := sha256.Sum256(pubKey)
	rip160Hasher := ripemd160.New()
	_,err := rip160Hasher.Write(hash[:])
	if err!=nil{
		log.Panic(err)
	}
	publicHash := rip160Hasher.Sum(nil)
	return publicHash
}
func CheckSum(payload []byte) []byte{
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	//4字节校验码
	checkSum := second[0:4]
	return checkSum
}

func IsValidAddress(address string) bool{
	decodeInfo , _:= base58.Decode(address)
	if len(decodeInfo) != 25{
		return false
	}
	payload := decodeInfo[0:len(decodeInfo)-4]
	//自己求出来的校验码
	checksum1 := CheckSum(payload)
	//解出来的校验码
	checksum2 := decodeInfo[len(decodeInfo)-4:]
	return bytes.Equal(checksum1,checksum2)
}

func IsMyAddress(address string,myId string) bool {
	ws := NewWallets(myId)
	addresses := ws.ListAddress()
	for _,myAddress := range addresses{
		if address == myAddress{
			return true
		}
	}
	return false
}