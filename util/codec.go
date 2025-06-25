package util

import "encoding/binary"

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
