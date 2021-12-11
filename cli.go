package main

import (
	"fmt"
)

//1. 所有的支配动作交给命令行来做
//2. 主函数只需要调用命令行结构即可
//3. 根据输入的不同命令，命令行做相应动作
//	1. addBlock
//	2. printChain

const Usage = `
	./CreateBlockChain 创建区块链
	./printChain 打印区块链
	./getInformation 获取地址所存信息
	./upload "上传数据"
	./Send "传递数据"
	./listAddress "打印所有地址"
	./createWallet  "创建地址"
	./printTX "打印交易"
`

//CLI: command line的缩写
type CLI struct {
	//bc *BlockChain//CLI中不需要保存区块链实例了，所有的命令在自己调用之前自己获取区块链实例
}

//给CLI提供一个方法，进行命令解析，从而执行调度
func (cli *CLI) Run(myId string) {
	fmt.Println(Usage)
	var cmds string
	fmt.Println("请输入指定命令")
	fmt.Scanln(&cmds)
	switch cmds {
	case "CreateBlockChain":
		fmt.Printf("创建区块链命令被调用\n")
		var addr string
		fmt.Println("请输入地址，返回请输入exit")
		fmt.Scanln(&addr)
		if addr == "exit" {
			cli.Run(myId)
		}
		cli.CreateBlockChain(addr)
		cli.Run(myId)
	case "printChain":
		fmt.Printf("打印区块链命令被调用\n")
		cli.PrintChain()
		cli.Run(myId)
	case "getInformation":
		fmt.Printf("打印地址所存信息命令被调用\n")
		var addr string
		fmt.Println("请输入地址,返回请输入exit")
		fmt.Scanln(&addr)
		if addr == "exit" {
			cli.Run(myId)
		}
		cli.GetInformation(addr, myId)
		cli.Run(myId)
	case "upload":
		fmt.Printf("上传数据命令被调用\n")
		var myAddress , userID string
		fmt.Println("请输入用户本地钱包地址,返回请输入exit")
		fmt.Scanln(&myAddress)
		if myAddress == "exit" {
			cli.Run(myId)
		}
		fmt.Println("请输入交易用户ID,返回请输入exit")
		fmt.Scanln(&userID)
		if userID == "exit" {
			cli.Run(myId)
		}
		cli.Upload(myAddress,userID)
		cli.Run(myId)
	case "Send":
		fmt.Printf("传递数据命令被调用\n")
		var from, to string
		var userid int
		fmt.Println("请输入本地钱包地址，返回请输入exit")
		fmt.Scanln(&from)
		if from == "exit" {
			cli.Run(myId)
		}
		fmt.Println("请输入目的交易地址，返回请输入exit")
		fmt.Scanln(&to)
		if to == "exit" {
			cli.Run(myId)
		}
		fmt.Println("请输入需要交易的用户ID，返回请输入-1")
		fmt.Scanln(&userid)
		if userid == -1 {
			cli.Run(myId)
		}
		cli.Send(from, to, userid, myId)
		cli.Run(myId)
	case "createWallet":
		fmt.Printf("创建地址命令被调用\n")
		cli.CreateWallet(myId)
		cli.Run(myId)
	case "listAddress":
		fmt.Printf("打印地址命令被调用\n")
		cli.listAddresses(myId)
		cli.Run(myId)
	case "printTX":
		fmt.Printf("打印交易命令被调用\n")
		cli.printTX()
		cli.Run(myId)
	default:
		fmt.Printf("无效的命令，请检查\n")
		cli.Run(myId)
	}
	//添加区块的时候：bc.addBlock(data) ，data通过os.Args拿回来
	//打印区块的时候：遍历区块链，不需要外部输入数据
}
