package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/mr-tron/base58/base58"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type Transaction struct {
	TXid      []byte     //交易id
	TXInputs  []TXInput  //所有的input
	TXOutputs []TXOutput //所有的output
}

//定义input
type TXInput struct {
	TXID  []byte //交易ID
	Index int64  //output的索引
	//Address string //解锁脚本
	Signature []byte //交易签名
	PubKey    []byte //公钥本身，不是公钥哈希
}

//定义output
type TXOutput struct {
	Value UserData //转账数据
	//Address string  //锁定脚本
	PubKeyHash []byte //公钥哈希
}

//给定转账地址找到公钥哈希完成交易锁定
func (output *TXOutput) Lock(address string) {
	decodeInfo, _ := base58.Decode(address)
	pubKeyHash := decodeInfo[1 : len(decodeInfo)-4]
	output.PubKeyHash = pubKeyHash
}

func NewTXOutput(value UserData, address string) TXOutput {
	output := TXOutput{
		Value: value,
	}
	output.Lock(address)
	return output
}

//定义交易的征信人的信息
type UserData struct {
	UserId      int
	Name        string
	Information []UserCredit
}

//定义存取款信息
type UserCredit struct {
	BorDate   uint64 //借款日期
	RepDate   uint64 //还款日期
	Value int64 //借款金额
	IsOverdue string //是否逾期
}

//计算交易ID
func (tx *Transaction) SetTXID() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash := sha256.Sum256(buffer.Bytes())
	tx.TXid = hash[:]
}

//征信数据上传，类似于挖矿交易
func UploadInformation(address string, idCard string) *Transaction {
	//表示Input为挖矿交易，所以Index设置为-1作为标识
	inputs := []TXInput{{
		TXID:      nil,
		Index:     -1,
		Signature: nil, //挖矿交易不需要签名
		PubKey:    nil,
	}}
	UserIdCard, _ := strconv.Atoi(idCard)
	UserID ,UserName  := FindUser(idCard)
	UserCre := FindUserCre(UserID)
	value := UserData{
		UserId: UserIdCard,
		Name:   UserName,
		Information: UserCre,
	}
	output := NewTXOutput(value, address)
	outputs := []TXOutput{output}
	tx := Transaction{
		TXid:      nil,
		TXInputs:  inputs,
		TXOutputs: outputs,
	}
	tx.SetTXID()

	return &tx
}

func FindUserCre(userId int) []UserCredit {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/mytest?charset=utf8")
	if err != nil {
		fmt.Println("连接数据库失败:", err)
		log.Panic(err)
	}
	//使用DB的query方法遍历数据库数据
	rows, err := db.Query("select * from `transaction`")
	//获取完毕释放rows，阻止更多的列举
	defer rows.Close()
	if err != nil {
		fmt.Println("获取错误:", err)
		log.Panic(err)
	}
	//如果有数据记录Next指针就不为true
	finUserCre := []UserCredit{}
	for rows.Next() {
		var creId int
		var BorData uint64
		var RepData uint64
		var IsOverdue string
		var Value int64
		var userid int
		rows.Scan(&creId,&BorData, &RepData, &IsOverdue,&Value,&userid)
		if userId == userid {
			fmt.Println("找到了")
			fmt.Println(BorData,RepData,IsOverdue,Value,userId)
			userCre := UserCredit{
				BorDate:   BorData,
				RepDate:   RepData,
				Value:     Value,
				IsOverdue: IsOverdue,
			}
			finUserCre = append(finUserCre, userCre)
		}
	}
	//Err返回可能的、在迭代时出现的错误。Err需在显式或隐式调用Close方法后调用。
	err = rows.Err()
	if err != nil {
		fmt.Println("other error:", err)
		log.Panic(err)
	}
	return finUserCre
}

func FindUser(idCard string) (userID int , username string) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/mytest?charset=utf8")
	if err != nil {
		fmt.Println("连接数据库失败:", err)
		return
	}
	//使用DB的query方法遍历数据库数据
	rows, err := db.Query("select * from `user`")
	//获取完毕释放rows，阻止更多的列举
	defer rows.Close()
	if err != nil {
		fmt.Println("获取错误:", err)
		return
	}
	var finUserID int
	var finUserName string
	//如果有数据记录Next指针就不为true
	for rows.Next() {
		var userid int
		var idcard string
		var username string
		rows.Scan(&userid, &idcard, &username)
		if idcard == idCard {
			finUserID , finUserName =userid, username
			break
		}
	}
	//Err返回可能的、在迭代时出现的错误。Err需在显式或隐式调用Close方法后调用。
	err = rows.Err()
	if err != nil {
		fmt.Println("other error:", err)
		return
	}
	return finUserID , finUserName
}


//传递的是挖矿的人
func NewCoinbaseTx(miner string, UserId int, UserName string) *Transaction {
	//TODO
	//表示Input为挖矿交易，所以Index设置为-1作为标识
	inputs := []TXInput{{
		TXID:      nil,
		Index:     -1,
		Signature: nil, //挖矿交易不需要签名
		PubKey:    nil,
	}}
	value := UserData{
		UserId: UserId,
		Name:   UserName,
		Information: []UserCredit{{
			BorDate:   uint64(time.Now().Unix()),
			RepDate:   uint64(time.Now().Unix()),
			IsOverdue: "NO",
			Value : 0,
		}},
	}
	output := NewTXOutput(value, miner)
	outputs := []TXOutput{output}
	tx := Transaction{
		TXid:      nil,
		TXInputs:  inputs,
		TXOutputs: outputs,
	}
	tx.SetTXID()

	return &tx
}

//判断是否是挖矿交易
func (tx *Transaction) IsCoinbase() bool {
	inputs := tx.TXInputs
	if len(inputs) == 1 && inputs[0].TXID == nil && inputs[0].Index == -1 {
		return true
	}
	return false
}

//普通交易
func NewTransaction(from, to string, userid int, bc *BlockChain, myId string) *Transaction {
	//打开钱包
	ws := NewWallets(myId)
	Wallet := ws.WalletMap[from]
	if Wallet == nil {
		fmt.Println("这个地址的私钥不存在，交易创建失败")
	}

	privateKey := Wallet.PrivateKey
	publicKey := Wallet.PublicKey
	publicKeyHash := HashPubKey(Wallet.PublicKey)
	utxos /*标识能用的utxo*/, resCredit /*这些utxo返回的数据*/ := bc.FindNeedUtxos(publicKeyHash, userid)
	if resCredit.UserId != userid {
		fmt.Printf("交易失败，没有找到%d的信息", userid)
		return nil
	}
	if resCredit.Information == nil {
		fmt.Println("没有要交易的信息")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput
	//将output转换成input
	for txid, indexes := range utxos {
		for _, i := range indexes {
			input := TXInput{
				TXID:      []byte(txid),
				Index:     i,
				Signature: nil,
				PubKey:    publicKey,
			}
			inputs = append(inputs, input)
		}
	}

	//创建输出
	value := UserData{
		UserId:      resCredit.UserId,
		Name:        resCredit.Name,
		Information: resCredit.Information,
	}
	output := NewTXOutput(value, to)
	outputs = append(outputs, output)

	tx := Transaction{
		TXid:      nil,
		TXInputs:  inputs,
		TXOutputs: outputs,
	}
	//设置交易id
	tx.SetTXID()
	//签名
	bc.SignTransaction(&tx, privateKey)
	//返回交易结构
	return &tx
}

//签名
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	fmt.Println("对交易进行签名")
	if tx.IsCoinbase() {
		return
	}
	//1. 拷贝一份交易txcopy，做相应裁剪，把每一个Input的pubkey和sig设置为空，output不做改变
	txCopy := tx.TrimmedCopy()
	//2. 遍历inputs，找到公钥哈希赋值给pubkey
	for i, input := range txCopy.TXInputs {
		//找到引用交易
		preTX := prevTXs[string(input.TXID)]
		output := preTX.TXOutputs[input.Index]
		txCopy.TXInputs[i].PubKey = output.PubKeyHash
		//3.生成要签名的数据哈希
		txCopy.SetTXID()
		signData := txCopy.TXid
		//请理以供下笔交易使用
		txCopy.TXInputs[i].PubKey = nil
		fmt.Printf("要签名的数据, signData : %x\n", signData)
		//4. 对数据签名
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, signData)
		if err != nil {
			fmt.Println("交易签名失败")
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TXInputs[i].Signature = signature
	}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	fmt.Println("广播交易,对交易进行IPBFT算法验证")
	clientSendMessageAndListen(prevTXs)
	//1. 拷贝修剪的副本
	txCopy := tx.TrimmedCopy()
	//2. 遍历原始交易
	for i, input := range tx.TXInputs {
		//3. 遍历原始交易的inputs所引用的前交易prevTX
		prevTX := prevTXs[string(input.TXID)]
		output := prevTX.TXOutputs[input.Index]
		//4. 找到output的公钥哈希，赋值给txCopy对应的input
		txCopy.TXInputs[i].PubKey = output.PubKeyHash
		//5. 还原签名数据
		txCopy.SetTXID()
		//清理动作
		txCopy.TXInputs[i].PubKey = nil
		verifyData := txCopy.TXid
		fmt.Printf("verifyData : %x\n", verifyData)
		//6. 校验
		//公钥字节流
		signature := input.Signature
		pubKeyBytes := input.PubKey
		//还原签名为r,s
		r := big.Int{}
		s := big.Int{}
		rData := signature[:len(signature)/2]
		sData := signature[len(signature)/2:]
		r.SetBytes(rData)
		s.SetBytes(sData)
		//还原公钥为Cruve,X,Y
		x := big.Int{}
		y := big.Int{}
		xData := pubKeyBytes[:len(pubKeyBytes)/2]
		yData := pubKeyBytes[len(pubKeyBytes)/2:]
		x.SetBytes(xData)
		y.SetBytes(yData)
		curve := elliptic.P256()
		publicKey := ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}
		//数据、签名、公钥准备完毕，开始校验
		if !ecdsa.Verify(&publicKey, verifyData, &r, &s) {
			fmt.Println(ecdsa.Verify(&publicKey, verifyData, &r, &s))
			return false
		}
	}
	return true
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.TXInputs {
		input1 := TXInput{
			TXID:      input.TXID,
			Index:     input.Index,
			Signature: nil,
			PubKey:    nil,
		}
		inputs = append(inputs, input1)
	}
	outputs = tx.TXOutputs
	tx1 := Transaction{
		TXid:      nil,
		TXInputs:  inputs,
		TXOutputs: outputs,
	}
	return tx1
}

func (tx *Transaction) String() string {
	var lines []string
	fmt.Println("--------------------------------------------------")
	lines = append(lines, fmt.Sprintf("---Transaction %x", tx.TXid))
	for i, input := range tx.TXInputs {
		lines = append(lines, fmt.Sprintf("Input %d", i))
		lines = append(lines, fmt.Sprintf(" 	TXID %x", input.TXID))
		lines = append(lines, fmt.Sprintf(" 	Out %d", input.Index))
		lines = append(lines, fmt.Sprintf("	Signature %x", input.Signature))
		lines = append(lines, fmt.Sprintf("	PubKey %x", input.PubKey))
	}
	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("Output %d", i))
		lines = append(lines, fmt.Sprintf("		借款人姓名 %s", output.Value.Name))
		lines = append(lines, fmt.Sprintf("		借款人ID  %d", output.Value.UserId))
		for _, v := range output.Value.Information {
			lines = append(lines, fmt.Sprintf("			借款日期 %s", time.Unix(int64(v.BorDate), 0).Format("2006-01-02 15:04:05")))
			lines = append(lines, fmt.Sprintf("			还款日期 %s", time.Unix(int64(v.RepDate), 0).Format("2006-01-02 15:04:05")))
			lines = append(lines, fmt.Sprintf("			是否逾期 %s", v.IsOverdue))
			lines = append(lines, fmt.Sprintf("			借款金额 %d", v.Value))
		}
		lines = append(lines, fmt.Sprintf(" 	PubKeyHash %x", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}
