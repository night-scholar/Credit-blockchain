package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"time"
)

//cli.go的配套文件，实现具体的命令

//func (cli *CLI) AddBlock(txs []*Transaction) {
//
//	cli.bc.AddBlock(txs)
//	fmt.Printf("添加区块成功\n")
//}

func (cli *CLI) GetInformation(addr, myId string) {
	if !IsMyAddress(addr, myId) || !IsValidAddress(addr) {
		fmt.Printf("%s不是我的地址或地址无效", addr)
		return
	}
	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()
	bc.GetInformation(addr)
}

func (cli *CLI) PrintChain() {
	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()
	it := bc.NewIterator()
	for {
		block := it.Next()

		fmt.Println("-----------------------------------------------------------------")
		fmt.Printf("Version : %x\n", block.Version)
		fmt.Printf("PrevBlockHash : %x\n", block.PrevBlockHash)
		fmt.Printf("MerkleRoot : %x\n", block.MerkleRoot)
		timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Printf("TimeStamp  : %s\n", timeFormat)
		fmt.Printf("Hash : %x\n", block.Hash)
		//fmt.Printf("Data : %s\n", block.Data)
		if bytes.Equal(block.PrevBlockHash, []byte{}) {
			fmt.Println("区块链遍历结束")
			break
		}
	}
}

func (cli *CLI) Send(from, to string, userid int, myId string) {
	if !IsMyAddress(from, myId) || !IsValidAddress(from) {
		fmt.Printf("%s不是我的地址或地址无效\n", from)
		return
	}
	if !IsValidAddress(to) {
		fmt.Printf("%s交易地址无效\n", to)
		return
	}
	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()
	//创建普通交易
	tx := NewTransaction(from, to, userid, bc, myId)
	if tx == nil {
		fmt.Println("交易无效，征信数据不存在")
		return
	}
	//添加到区块
	bc.AddBlock([]*Transaction{tx})
	//生成json
	b, err := json.Marshal(tx.TXOutputs)
	if err != nil {
		fmt.Printf("json.Marshal failed, err:%v\n", err)
		return
	}
	err = ioutil.WriteFile("transaction.json", b, os.ModeAppend)
	if err != nil {
		return
	}
}

func (cli *CLI) CreateBlockChain(addr string) {
	if !IsValidAddress(addr) {
		fmt.Println("地址无效")
		return
	}
	bc := CreateBlockchain(addr)

	if bc == nil {
		return
	}
	defer bc.db.Close()
	fmt.Println("创建区块链成功")
}

func (cli *CLI) CreateWallet(myId string) {
	ws := NewWallets(myId)
	address := ws.CreateWallet(myId)
	fmt.Printf("新的钱包地址为 : %s ", address)
}

func (cli *CLI) listAddresses(myId string) {
	ws := NewWallets(myId)
	addresses := ws.ListAddress()
	for _, address := range addresses {
		fmt.Printf("address : %s\n", address)
	}
}

func (cli *CLI) printTX() {
	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()
	it := bc.NewIterator()

	for {
		block := it.Next()
		for _, tx := range block.Transactions {
			fmt.Printf("tx : %v\n", tx)
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

//上传数据
func (cli *CLI) Upload(myAddress , userid string) {
	bc := NewBlockChain()
	if bc == nil {
		return
	}
	defer bc.db.Close()
	//上传征信数据
	coinbase := UploadInformation(myAddress, userid)
	txs := []*Transaction{coinbase}
	bc.AddBlock(txs)
}
