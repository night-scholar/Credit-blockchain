package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
)

type Wallets struct {
	WalletMap map[string]*WalletKeyPair
}

//创建wallets
func NewWallets(myId string) * Wallets{
	//把所有钱包从本地加载出来
	var ws Wallets
	ws.WalletMap = make(map[string]*WalletKeyPair)
	if !ws.LoadFromFile(myId){
		fmt.Println("加载钱包数据失败")
	}
	//返回实例
	return &ws
}

const WalletName = "wallet.dat"

func (ws *Wallets) CreateWallet(myId string) string {
	wallet := NewWalletKeyPair()
	address := wallet.GetAddress()
	ws.WalletMap[address] = wallet

	res := ws.SaveToFile(myId)
	if !res{
		fmt.Println("创建钱包失败")
	}
	return address
}

func (ws *Wallets) SaveToFile(myId string) bool {
	var buffer bytes.Buffer
	//将接口类型进行注册
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(ws)
	if err!=nil{
		fmt.Println("钱包序列化失败")
		return false
	}
	content := buffer.Bytes()
	//保存到文件
	err = ioutil.WriteFile("./"+myId+"./"+WalletName , content , 0600)
	if err!=nil{
		fmt.Println("钱包创建失败")
		return false
	}
	return true
}

func (ws *Wallets) LoadFromFile(myId string) bool {
	if !IsFileExit("./"+myId+"./"+WalletName){
		fmt.Println("钱包文件不存在，准备创建")
		return true
	}
	//读取文件
	content , err := ioutil.ReadFile("./"+myId+"./"+WalletName)
	if err!=nil{
		fmt.Println("文件读取失败")
		return false
	}
	//gob解码
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))
	var wallets Wallets
	err = decoder.Decode(&wallets)
	if err != nil{
		return false
	}
	ws.WalletMap = wallets.WalletMap
	return true
}

func (ws *Wallets) ListAddress() []string {
	var addresses []string
	for address := range ws.WalletMap{
		addresses = append(addresses, address)
	}
	return addresses
}