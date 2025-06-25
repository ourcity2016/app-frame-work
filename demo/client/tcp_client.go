package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// EncodeMessage 将消息编码为协议格式
func EncodeMessage(data []byte) []byte {
	length := uint32(len(data))

	// 创建缓冲区：4字节长度 + 数据长度
	buf := make([]byte, 4+len(data))

	// 写入长度字段（小端序）
	binary.LittleEndian.PutUint32(buf[:4], length)

	// 写入数据
	copy(buf[4:], data)

	return buf
}
func decodeMessage(conn net.Conn) ([]byte, error) {
	// 使用更大的初始缓冲区(或可配置)
	s := bufio.NewReaderSize(conn, 16*1024)

	// 读取并消费长度前缀
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(s, lengthBytes); err != nil {
		return nil, fmt.Errorf("读取长度前缀失败: %w", err)
	}

	length := binary.LittleEndian.Uint32(lengthBytes)

	// 更严格的大小限制检查
	const maxMessageSize = 10 * 1024 * 1024
	if length == 0 {
		return nil, errors.New("零长度消息")
	}
	if length > maxMessageSize {
		return nil, fmt.Errorf("消息长度 %d 超过限制 %d", length, maxMessageSize)
	}

	// 读取消息体
	message := make([]byte, length)
	if _, err := io.ReadFull(s, message); err != nil {
		return nil, fmt.Errorf("读取消息体失败: %w", err)
	}
	return message[4:len(message)], nil
}
func mainx() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
	}
	//go func() {
	//	for {
	//		conn.Write(EncodeMessage([]byte("{\"cmd\":\"request\",\"router\":\"xxx\"}")))
	//		time.Sleep(time.Second * 10)
	//	}
	//}()
	go func() {
		for {
			msg, err := decodeMessage(conn)
			if err != nil {
				log.Printf(err.Error())
			}
			log.Print(string(msg))
		}
	}()
	//select {}
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		fmt.Println(input)
		inputByte := []byte(input)
		data := EncodeMessage(inputByte)
		conn.Write(data)
	}
}
