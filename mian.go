package main

import (
	"fmt"
	"log"
	"os"
)
const nodeCount = 4

//客户端的监听地址
var clientAddr = "127.0.0.1:8888"

//节点池，主要用来存储监听地址
var nodeTable map[string]string
func main() {
	//为四个节点生成公私钥
	genRsaKeys()
	nodeTable = map[string]string{
		"N0": "127.0.0.1:8000",
		"N1": "127.0.0.1:8001",
		"N2": "127.0.0.1:8002",
		"N3": "127.0.0.1:8003",
	}
	cmds := os.Args
	if len(cmds) != 2{
		fmt.Println("命令输入错误，请输入blockchain+用户名")
		os.Exit(1)
	}
	nodeID := os.Args[1]
	if addr, ok := nodeTable[nodeID]; ok {
		os.Mkdir(cmds[1]+"wallet", os.ModePerm)
		cli := CLI{}
		go cli.Run(cmds[1]+"wallet")
		p := NewPBFT(nodeID, addr)
		go p.tcpListen() //启动节点
	} else {
		log.Fatal("无此节点编号！")
	}
	select {}



}