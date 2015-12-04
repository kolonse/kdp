// kdp
/**
*	kdp 协议  全称:kolonse diy proto
*	用于做TCP方式双向通信简单协议
*	主要用于简化TCP通信时使用JSON或者二进制方式导致不方便查看数据内容的问题
 */
package kdp

import (
	"strconv"
	"strings"
)

const (
	KDP_PROTO_METHOD_CONN  = "CONN"
	KDP_PROTO_METHOD_REQ   = "REQ"
	KDP_PROTO_METHOD_RES   = "RES"
	KDP_PROTO_METHOD_CLOSE = "CLOSE"
)

/**
*	TCP 协议
 */
const (
	KDP_PROTO_HEAD_MARK   = "-------kolonse head mark begin-------\r\n"
	KDP_PROTO_BODY_LENGTH = "Content Length:"
	KDP_PROTO_LOCAL_ADDR  = "Local Addr:"
	KDP_PROTO_REMOTE_ADDR = "Remote Addr:"
	KDP_PROTO_LINE_END    = "\r\n"
	KDP_PROTO_HEAD_END    = "-------kolonse head mark end-------\r\n"
)

type KDP struct {
	mark       string
	method     string
	version    string
	localAddr  string
	remoteAddr string
	bodyLength int
	headLength int
	bodyBuff   []byte
	headBuff   []byte
	buff       []byte
	err        Error
}

func (pp *KDP) HeaderString() string {
	return string(pp.headBuff)
}

func (pp *KDP) Parse(buff []byte) *KDP {
	pp.buff = make([]byte, len(buff))
	copy(pp.buff, buff)
	return pp.ParseMark().
		ParseBody().
		ParseProto().
		ParseLocalAddr().
		ParseRemoteAddr()
}

func (pp *KDP) ParseMark() *KDP {
	if len(pp.buff) < len(KDP_PROTO_HEAD_MARK) { // buff 长度不够处理 mark 头标志
		pp.err = NewError(KDP_PROTO_ERROR_LENGTH, "parse mark: length not enougth")
	} else {
		if string(pp.buff[0:len(KDP_PROTO_HEAD_MARK)]) != KDP_PROTO_HEAD_MARK {
			pp.err = NewError(KDP_PROTO_ERROR_NOT_KDP_PROTO, "not found head mark begin")
			return pp
		}
		pp.mark = string(KDP_PROTO_HEAD_MARK)
		// 将 header buff 单独拿出来
		index := strings.Index(string(pp.buff), string(KDP_PROTO_HEAD_END))
		if index == -1 { // 只要没找到协议头结尾标志 那么就认为长度不足
			pp.err = NewError(KDP_PROTO_ERROR_LENGTH, "not found head mark end")
		} else {
			pp.headBuff = make([]byte, index-len(KDP_PROTO_HEAD_MARK))
			copy(pp.headBuff, pp.buff[len(KDP_PROTO_HEAD_MARK):index])
			pp.headLength = index + len(KDP_PROTO_HEAD_END)
		}
	}
	//	if( )
	return pp
}

func (pp *KDP) ParseProto() *KDP { //  协议必定在 mark 之后
	if pp.HaveError() {
		//  读取协议
		index := strings.Index(string(pp.headBuff), " ")
		if index == -1 { // 协议字段就是 第一行空格前的字符串
			pp.err = NewError(KDP_PROTO_ERROR_FORMAT, "not found Proto")
			return pp
		}
		pp.method = string(pp.headBuff[0:index])
		//提取版本头
		indexEnd := strings.Index(string(pp.headBuff), KDP_PROTO_LINE_END)
		pp.version = string(pp.headBuff[index+1 : indexEnd])
	}
	return pp
}

func (pp *KDP) ParseLocalAddr() *KDP {
	if pp.HaveError() {
		index := strings.Index(string(pp.headBuff), string(KDP_PROTO_LOCAL_ADDR))
		if index == -1 { //  如果没有 Content Length 字段 那么就认为是 0
			//pp.err = NewError(KDP_PROTO_ERROR_FORMAT, "not found Content Length")
			return pp
		}

		indexEnd := strings.Index(string(pp.headBuff[index:]), string(KDP_PROTO_LINE_END))
		if indexEnd == -1 {
			pp.err = NewError(KDP_PROTO_ERROR_FORMAT, "not found RemoteAddr's Line End")
			return pp
		}
		pp.remoteAddr = string(pp.headBuff[index+len(KDP_PROTO_LOCAL_ADDR) : index+indexEnd])
	}
	return pp
}

func (pp *KDP) ParseRemoteAddr() *KDP {
	if pp.HaveError() {
		index := strings.Index(string(pp.headBuff), string(KDP_PROTO_REMOTE_ADDR))
		if index == -1 { //  如果没有 Content Length 字段 那么就认为是 0
			//pp.err = NewError(KDP_PROTO_ERROR_FORMAT, "not found Content Length")
			return pp
		}

		indexEnd := strings.Index(string(pp.headBuff[index:]), string(KDP_PROTO_LINE_END))
		if indexEnd == -1 {
			pp.err = NewError(KDP_PROTO_ERROR_FORMAT, "not found RemoteAddr's Line End")
			return pp
		}
		pp.remoteAddr = string(pp.headBuff[index+len(KDP_PROTO_REMOTE_ADDR) : index+indexEnd])
	}
	return pp
}

func (pp *KDP) parseBodyLength() *KDP {
	if pp.HaveError() {
		//  查找 body length 字段
		index := strings.Index(string(pp.headBuff), string(KDP_PROTO_BODY_LENGTH))
		if index == -1 { //  如果没有 Content Length 字段 那么就认为是 0
			//pp.err = NewError(KDP_PROTO_ERROR_FORMAT, "not found Content Length")
			return pp
		}

		indexEnd := strings.Index(string(pp.headBuff[index:]), string(KDP_PROTO_LINE_END))
		if indexEnd == -1 {
			pp.err = NewError(KDP_PROTO_ERROR_FORMAT, "not found Content Length's Line End")
			return pp
		}
		lengthString := string(pp.headBuff[index+len(KDP_PROTO_BODY_LENGTH) : index+indexEnd])
		lengthInt, err := strconv.Atoi(lengthString)
		if err != nil {
			pp.err = NewError(KDP_PROTO_ERROR_FORMAT, err.Error())
			return pp
		}
		pp.bodyLength = lengthInt
	}
	return pp
}

func (pp *KDP) ParseBody() *KDP {
	pp.parseBodyLength()
	if pp.HaveError() {
		if pp.bodyLength == 0 { // 长度为0 就不进行解析BUFF
			return pp
		}
		if len(pp.buff) < pp.bodyLength+pp.headLength { // 长度不足
			pp.err = NewError(KDP_PROTO_ERROR_LENGTH, "parse body: length not enougth")
			return pp
		}

		pp.bodyBuff = make([]byte, pp.bodyLength)
		copy(pp.bodyBuff, pp.buff[pp.headLength:])
	}
	return pp
}

func (pp *KDP) StringifyLocalAddr(addr string) *KDP {
	pp.headBuff = append(pp.headBuff, []byte(KDP_PROTO_LOCAL_ADDR)...)
	pp.headBuff = append(pp.headBuff, []byte(addr)...)
	pp.headBuff = append(pp.headBuff, []byte(KDP_PROTO_LINE_END)...)
	return pp
}

func (pp *KDP) StringifyRemoteAddr(addr string) *KDP {
	pp.headBuff = append(pp.headBuff, []byte(KDP_PROTO_REMOTE_ADDR)...)
	pp.headBuff = append(pp.headBuff, []byte(addr)...)
	pp.headBuff = append(pp.headBuff, []byte(KDP_PROTO_LINE_END)...)
	return pp
}

func (pp *KDP) StringifyBody(body []byte) *KDP {
	if len(body) == 0 { // 如果长度为 0 那么就不进行处理 body
		return pp
	}
	bodyLenString := strconv.Itoa(len(body))
	pp.headBuff = append(pp.headBuff, []byte(KDP_PROTO_BODY_LENGTH)...)
	pp.headBuff = append(pp.headBuff, []byte(bodyLenString)...)
	pp.headBuff = append(pp.headBuff, []byte(KDP_PROTO_LINE_END)...)
	pp.bodyBuff = make([]byte, len(body))
	copy(pp.bodyBuff, body)
	return pp
}

func (pp *KDP) StringifyEnd() *KDP {
	pp.buff = append(pp.buff, []byte(KDP_PROTO_HEAD_MARK)...)
	pp.buff = append(pp.buff, pp.headBuff...)
	pp.buff = append(pp.buff, []byte(KDP_PROTO_HEAD_END)...)
	pp.buff = append(pp.buff, pp.bodyBuff...)
	return pp
}

func (pp *KDP) GetError() Error {
	return pp.err
}

func (pp *KDP) HaveError() bool {
	return pp.err.GetCode() == 0
}

func (pp *KDP) GetBody() []byte {
	return pp.bodyBuff
}

func (pp *KDP) GetMethod() string {
	return pp.method
}

func (pp *KDP) GetVersion() string {
	return pp.version
}

func (pp *KDP) GetLocalAddr() string {
	return pp.localAddr
}

func (pp *KDP) GetRemoteAddr() string {
	return pp.remoteAddr
}

func (pp *KDP) GetBodyLength() int {
	return pp.bodyLength
}

func (pp *KDP) GetBuff() []byte {
	return pp.buff
}

func (pp *KDP) GetProtoLen() int {
	return pp.headLength + pp.bodyLength
}

func NewKDP() *KDP {
	return &KDP{}
}
