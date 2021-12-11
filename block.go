package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"log"
	"time"
)

const genesisInfo = "hello world"

//定义结构
type Block struct {
	Version       uint64 //区块版本号
	PrevBlockHash []byte //前区块哈希
	MerkleRoot    []byte //Merkle根
	TimeStamp     uint64 //从1970年1月1日至今的秒数
	Hash          []byte //当前区块哈希
	//Data          []byte //数据，目前为字节流 后续使用交易代替
	Transactions []*Transaction
}

type MerkleNode struct {
	Left *MerkleNode
	Right *MerkleNode
	Data []byte
}
type MerkleTree struct {
	node *MerkleNode
}
func NewMerkleNode(left,right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	}else {
		prevHashes := append(left.Data,right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}
	mNode.Left = left
	mNode.Right = right
	return &mNode
}
func NewMerkleTree(data [][]byte) *MerkleTree  {
	var nodes []MerkleNode
	if len(data) % 2 != 0 {
		data = append(data, data[len(data) - 1])
	}
	for _, dataitem := range data {
		node := NewMerkleNode(nil, nil, dataitem)
		nodes = append(nodes, *node)
	}
	for i := 0; i<len(data)/2; i++ {
		var newNodes []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newNodes = append(newNodes , *node)
		}
		nodes = newNodes
	}
	mTree := MerkleTree{&nodes[0]}
	return &mTree
}
func (block *Block) HashTransactions()  {
	var hashes [][]byte
	//交易的ID就是交易的哈希值
	for _,tx := range block.Transactions{
		for _,output := range tx.TXOutputs{
			outputbyte , _ := json.Marshal(output.Value)
			hashes = append(hashes , outputbyte)
		}
	}
	mTree := NewMerkleTree(hashes)
	block.MerkleRoot = mTree.node.Data
}

//创建区块 对Block的每一个字段填充数据
func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := Block{
		Version:       00,
		PrevBlockHash: prevBlockHash,
		MerkleRoot:    []byte{}, //先填充为空
		TimeStamp:     uint64(time.Now().Unix()),
		Hash:          []byte{},     //先填充为空
		Transactions: txs,
	}
	block.SetHash()
	block.HashTransactions()//计算Merkle根
	return &block
}

//为了生成区块哈希，我们生成一个简单的函数来计算哈希值
func (block *Block) SetHash() {
	var data []byte
	tmp := [][]byte{
		uintToByte(block.Version),
		block.PrevBlockHash,
		block.MerkleRoot,
		uintToByte(block.TimeStamp),
		//block.Data,//TODO
	}
	data = bytes.Join(tmp, []byte{})

	hash := sha256.Sum256(data) //hash是一个32位的数组
	block.Hash = hash[:]
}

//1. gob是go语言内置的编码包
//2. 他可以对任意数据类型进行编码和解码
//3. 编码时，先要创建编码器，编码器进行编码
//4. 解码时，先要创建解码器，解码器进行解码

//序列化 将区块转换成字节流
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer

	//定义编码器
	encoder := gob.NewEncoder(&buffer)
	//编码器对结构进行编码
	err := encoder.Encode(&block)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}

//反序列化
func DeSerialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	//解码器对结构进行解码
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}