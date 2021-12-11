package main

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/mr-tron/base58/base58"
	"log"
	"os"
	"time"
)

//使用bolt进行改写，需要两个字段
//1. bolt数据库的句柄
//2. 最后一个区块的哈希值
type BlockChain struct {
	db *bolt.DB //句柄

	tail []byte //代表最后一个区块的哈希值
}

const blockChainName = "blockChain.db"
const blockBucketName = "blockBucket"
const lastHashKey = "lastHashKey"

//实现创建区块链的方法
func CreateBlockchain(miner string) *BlockChain {
	if IsFileExit(blockChainName) {
		fmt.Println("区块链已经存在，不需要创建")
		return nil
	}
	//功能分析：
	//1. 获得区块链数据库的句柄，打开数据库，读写数据
	db, err := bolt.Open(blockChainName, 0600, nil) //"test.db"是数据库名称 0600是读写
	//向数据库中写入数据
	//从数据库中读取数据

	if err != nil {
		log.Panic(err)
	}

	var tail []byte

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blockBucketName))
		if err != nil {
			log.Panic(err)
		}
		//开始添加创世块
		coinbase := NewCoinbaseTx(miner, 0000000000, "创世块")
		genesisBlock := NewBlock([]*Transaction{coinbase}, []byte{})

		b.Put(genesisBlock.Hash, genesisBlock.Serialize() /*将区块序列化，转成字节流*/)
		b.Put([]byte(lastHashKey), genesisBlock.Hash)

		//测试，读取写入的数据
		//blockInfo := b.Get(genesisBlock.Hash)
		//block := deSerialize(blockInfo)
		//fmt.Printf("解码后的block数据：%s\n", block)

		//更新tail位最后一个区块的哈希值
		tail = genesisBlock.Hash
		return nil
	})
	//返回实例
	return &BlockChain{
		db:   db,
		tail: tail,
	}
}

//返回区块链实例
func NewBlockChain() *BlockChain {
	if !IsFileExit(blockChainName) {
		fmt.Println("区块链不存在，请先创建")
		return nil
	}
	db, err := bolt.Open(blockChainName, 0600, nil) //"test.db"是数据库名称 0600是读写
	if err != nil {
		log.Panic(err)
	}
	var tail []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))
		if b == nil {
			fmt.Println("bucket为空,请检查")
			os.Exit(1)
		}

		// 获取最后一个区块的哈希值，填充给tail
		tail = b.Get([]byte(lastHashKey))
		return nil
	})
	//返回实例
	return &BlockChain{
		db:   db,
		tail: tail,
	}
}

func (bc *BlockChain) AddBlock(txs []*Transaction) {
	//矿工得到交易时第一时间对交易进行验证
	validTXs := []*Transaction{}
	for _, tx := range txs {
		if bc.VerifyTransaction(tx) {
			fmt.Println("交易有效")
			validTXs = append(validTXs, tx)
		} else {
			fmt.Println("发现无效交易")
		}
	}
	bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))

		if b == nil {
			//如果b1为空，说明这个名字不存在，需要检查
			fmt.Println("bucket不存在，请检查")
			os.Exit(1) //退出
		}

		//创建一个区块
		block := NewBlock(txs, bc.tail)
		b.Put(block.Hash, block.Serialize() /*将区块序列化，转成字节流*/)
		b.Put([]byte(lastHashKey), block.Hash)

		//测试，读取写入的数据
		//blockInfo := b.Get(block.Hash)
		//nowblock := deSerialize(blockInfo)
		//fmt.Printf("解码后的block数据：%s\n", nowblock)

		//更新tail位最后一个区块的哈希值
		bc.tail = block.Hash

		return nil
	})
}

//定义一个区块链的迭代器，包含db，current
type BlockChainIterator struct {
	db      *bolt.DB //账本
	current []byte   //当前所指向区块的哈希值
}

//创建迭代器，使用bc进行初始化
func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{
		db:      bc.db,
		current: bc.tail,
	}
}

//读取数据
func (it *BlockChainIterator) Next() *Block {
	var block Block
	//读取数据库
	it.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucketName))
		if b == nil {
			fmt.Println("bucket不存在,请检查")
			os.Exit(1)
		}

		//读取数据
		blockInfo /*block的字节流*/ := b.Get(it.current)
		block = *DeSerialize(blockInfo)

		it.current = block.PrevBlockHash
		return nil
	})
	return &block
}

type UTXOInfo struct {
	TXID   []byte   //交易id
	Index  int64    //output的索引
	Output TXOutput //output本身
}

//找到属于我的交易
func (bc *BlockChain) FindMyUtxos(pubKeyHash []byte) []UTXOInfo {
	//创建迭代器
	it := bc.NewIterator()
	//var utxos []TXOutput //返回的结构
	var UTXOInfos []UTXOInfo               //返回的结构
	spentUTXOs := make(map[string][]int64) //标识已经消耗过的UTXO的结构，key是交易id，value是这个id里面的output的索引的数组
	//遍历账本
	for {
		block := it.Next()
		//遍历交易
		for _, tx := range block.Transactions {
			//标识出一个使用过的output需要两个数据：TXID和index
			if tx.IsCoinbase() == false {
				for _, input := range tx.TXInputs {
					if bytes.Equal(HashPubKey(input.PubKey), pubKeyHash) {
						fmt.Println("已经交易过的数据")
						key := string(input.TXID)
						spentUTXOs[key] = append(spentUTXOs[key], input.Index)
					}
				}
			}
			//遍历output
		OUTPUT:
			for i, output := range tx.TXOutputs {
				key := string(tx.TXid)
				indexes := spentUTXOs[key]
				if len(indexes) != 0 {
					fmt.Println("交易中有已经被交易的数据")
					for _, j := range indexes {
						if int64(i) == j {
							fmt.Println("当前数据已经被交易了")
							continue OUTPUT
						}
					}
				}
				//找到属于我的output
				if bytes.Equal(pubKeyHash, output.PubKeyHash) {
					fmt.Println("还未交易的数据")
					//utxos = append(utxos, output)
					utxoinfo := UTXOInfo{
						TXID:   tx.TXid,
						Index:  int64(i),
						Output: output,
					}
					UTXOInfos = append(UTXOInfos, utxoinfo)
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			fmt.Println("遍历区块链结束")
			break
		}
	}
	return UTXOInfos
}

func (bc *BlockChain) GetInformation(address string) {
	decodeInfo, _ := base58.Decode(address)
	pubKeyHash := decodeInfo[1 : len(decodeInfo)-4]
	utxoinfos := bc.FindMyUtxos(pubKeyHash) //得到所有的UTXO
	for _, utxoinfo := range utxoinfos {
		fmt.Println("姓名 : ", utxoinfo.Output.Value.Name)
		fmt.Println("身份证ID : ", utxoinfo.Output.Value.UserId)
		for _, information := range utxoinfo.Output.Value.Information {
			fmt.Println("借款日期 : ", time.Unix(int64(information.BorDate), 0).Format("2006-01-02 15:04:05"))
			fmt.Println("还款日期 : ", time.Unix(int64(information.RepDate), 0).Format("2006-01-02 15:04:05"))
			fmt.Println("是否逾期 : ", information.IsOverdue)
			fmt.Println("借款金额 : ", information.Value)
		}
	}
}

//找到适合的utxo
func (bc *BlockChain) FindNeedUtxos(pubKeyHash []byte, userid int) (map[string][]int64, UserData) {
	needUtxos := make(map[string][]int64)
	var resCredit UserData //统计的结构
	utxoinfos := bc.FindMyUtxos(pubKeyHash)
	for _, utxoinfo := range utxoinfos {
		key := string(utxoinfo.TXID)
		if utxoinfo.Output.Value.UserId == userid {
			needUtxos[key] = append(needUtxos[key], utxoinfo.Index)
			for _, needInformation := range utxoinfo.Output.Value.Information {
				resCredit.Information = append(resCredit.Information, needInformation)
			}
			resCredit.UserId = userid
			resCredit.Name = utxoinfo.Output.Value.Name
		}
	}
	return needUtxos, resCredit
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
	//遍历账本找到所有引用的交易
	prevTXs := make(map[string]Transaction)

	for _, input := range tx.TXInputs {
		prevTX := bc.FindTransaction(input.TXID)
		if prevTX == nil {
			fmt.Println("没有找到交易")
		} else {
			prevTXs[string(input.TXID)] = *prevTX
		}
	}
	tx.Sign(privateKey, prevTXs)
}

func (bc *BlockChain) FindTransaction(txid []byte) *Transaction {
	it := bc.NewIterator()
	for {
		block := it.Next()
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.TXid, txid) {
				return tx
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil
}

//矿工校验流程
//找到交易input所引用的所有的交易PrevTXs
//对交易进行校验
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	//遍历账本找到所有引用的交易
	prevTXs := make(map[string]Transaction)

	for _, input := range tx.TXInputs {
		prevTX := bc.FindTransaction(input.TXID)
		if prevTX == nil {
			fmt.Println("没有找到交易")
		} else {
			prevTXs[string(input.TXID)] = *prevTX
		}
	}
	return tx.Verify(prevTXs)
}
