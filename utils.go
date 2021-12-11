package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

//这是一个工具函数文件
//将数字转成byte类型
func uintToByte(num uint64) []byte {
	//TODO
	//使用binary.Write来进行编码
	var buffer bytes.Buffer
	//编码要进行校验
	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buffer.Bytes()
}
//也可使用binary.Read()进行解码

func IsFileExit(fileName string) bool {
	//使用os.Stat来判断文件是否存在
	_ , err := os.Stat(fileName)
	if os.IsNotExist(err){
		return false
	}
	return true
}