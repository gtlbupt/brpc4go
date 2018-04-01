package baidu_std

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	BAIDU_STD_RPC_MSG_HEADER_LEN = 12
	BAIDU_STD_MAGIC_STRING       = "PRPC"
)

type BaiduRpcStdProtocolHeader struct {
	bodySize uint32
	metaSize uint32
}

func (h *BaiduRpcStdProtocolHeader) GetHeaderLen() int {
	return BAIDU_STD_RPC_MSG_HEADER_LEN
}

func (h *BaiduRpcStdProtocolHeader) GetMetaSize() int {
	return int(h.metaSize)
}

func (h *BaiduRpcStdProtocolHeader) SetMetaSize(size int) {
	h.metaSize = uint32(size)
}

func (h *BaiduRpcStdProtocolHeader) GetBodySize() int {
	return int(h.bodySize)
}

func (h *BaiduRpcStdProtocolHeader) SetBodySize(size int) {
	h.bodySize = uint32(size)
}

func (h *BaiduRpcStdProtocolHeader) Marshal() ([]byte, error) {
	var buf = bytes.NewBuffer(make([]byte, 0, BAIDU_STD_RPC_MSG_HEADER_LEN))

	binary.Write(buf, binary.LittleEndian, []byte(BAIDU_STD_MAGIC_STRING))
	binary.Write(buf, binary.BigEndian, h.bodySize)
	binary.Write(buf, binary.BigEndian, h.metaSize)

	return buf.Bytes(), nil
}

func (h *BaiduRpcStdProtocolHeader) Unmarshal(data []byte) error {
	if len(data) < BAIDU_STD_RPC_MSG_HEADER_LEN {
		return errors.New("Bad RPC Header Length")
	}

	if string(data[:4]) != BAIDU_STD_MAGIC_STRING {
		return errors.New("Bad Magic String")
	}

	h.bodySize = binary.BigEndian.Uint32(data[4:8])
	h.metaSize = binary.BigEndian.Uint32(data[8:12])

	return nil
}

func (h *BaiduRpcStdProtocolHeader) String() string {
	return fmt.Sprintf("[meta_size:%d][body_size:%d]", h.metaSize, h.bodySize)
}
