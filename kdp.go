// kdp
/**
*	kdp 协议  全称:kolonse diy proto
*	用于做TCP方式双向通信简单协议
*	主要用于简化TCP通信时使用JSON或者二进制方式导致不方便查看数据内容的问题
*	@2016.3.8 by jszhou2
*	将 KDP_PROTO_HEAD_MARK 和 KDP_PROTO_HEAD_END 长度改小
*	khmb和khme表示 kolonse head mark begin/end
 */
package kdp

import (
	"strconv"
	"strings"
)

/**
*	TCP 协议
 */

const (
	KDP_PROTO_HEAD_MARK   = "-khmb-\r\n"
	KDP_PROTO_BODY_LENGTH = "Content Length:"
	KDP_PROTO_LOCAL_ADDR  = "Local Addr:"
	KDP_PROTO_REMOTE_ADDR = "Remote Addr:"
	KDP_PROTO_LINE_END    = "\r\n"
	KDP_PROTO_HEAD_END    = "-khme-\r\n"
)

type KDP struct {
	mark       string
	bodyLength int
	headLength int
	bodyBuff   []byte
	headBuff   []byte
	buff       []byte
	err        Error
	heads      map[string]string
}

func (pp *KDP) Add(key, value string) *KDP {
	pp.heads[key] = value
	return pp
}

func (pp *KDP) Get(key string) (string, bool) {
	value, ok := pp.heads[key]
	return value, ok
}

func (pp *KDP) ParseHead() *KDP {
	headString := pp.HeaderString()
	heads := strings.Split(headString, "\r\n")
	for _, data := range heads {
		data = strings.Trim(data, " ")
		if len(data) == 0 {
			continue
		}
		group := strings.Split(data, ":")
		if len(group) > 2 { //如果不等 2 说明 value中存在 : 需要将他们拼接起来
			for i := 2; i < len(group); i++ {
				group[1] = group[1] + ":" + group[i]
			}
		} else if len(group) < 2 {
			continue
		}
		pp.heads[group[0]] = group[1]
	}
	return pp
}

func (pp *KDP) HeaderString() string {
	return string(pp.headBuff)
}

func (pp *KDP) Parse(buff []byte) *KDP {
	pp.buff = make([]byte, len(buff))
	copy(pp.buff, buff)
	return pp.ParseMark().
		ParseHead().
		ParseBody()
}

func (pp *KDP) ParseMark() *KDP {
	if len(pp.buff) < len(KDP_PROTO_HEAD_MARK) { // buff 长度不够处理 mark 头标志
		// 预先判断收到的头是否和  mark 头匹配
		if string(pp.buff) == string(KDP_PROTO_HEAD_MARK[:len(pp.buff)]) {
			pp.err = NewError(KDP_PROTO_ERROR_LENGTH, "parse mark: mark length not enougth")
		} else {
			pp.err = NewError(KDP_PROTO_ERROR_FORMAT, "parse mark: not kdp proto")
		}
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
	return pp
}

func (pp *KDP) parseBodyLength() *KDP {
	if pp.NotHaveError() {
		value, ok := pp.heads["Content Length"]
		if !ok {
			pp.bodyLength = 0
			return pp
		}
		lengthInt, err := strconv.Atoi(value)
		if err != nil {
			pp.err = NewError(KDP_PROTO_ERROR_FORMAT, err.Error())
			return pp
		}
		pp.bodyLength = lengthInt
	}
	return pp
}

func (pp *KDP) ParseBody() *KDP {
	if pp.NotHaveError() {
		pp.parseBodyLength()
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

func (pp *KDP) StringifyBody(body []byte) *KDP {
	bodyLenString := strconv.Itoa(len(body))
	pp.Add("Content Length", bodyLenString)
	pp.bodyBuff = make([]byte, len(body))
	copy(pp.bodyBuff, body)
	return pp
}

func (pp *KDP) StringifyHead() *KDP {
	for key, value := range pp.heads {
		pp.headBuff = append(pp.headBuff, []byte(key+":")...)
		pp.headBuff = append(pp.headBuff, []byte(value)...)
		pp.headBuff = append(pp.headBuff, []byte(KDP_PROTO_LINE_END)...)
	}
	return pp
}

func (pp *KDP) Stringify() *KDP {
	pp.StringifyHead()
	pp.buff = append(pp.buff, []byte(KDP_PROTO_HEAD_MARK)...)
	pp.buff = append(pp.buff, pp.headBuff...)
	pp.buff = append(pp.buff, []byte(KDP_PROTO_HEAD_END)...)
	pp.buff = append(pp.buff, pp.bodyBuff...)
	return pp
}

func (pp *KDP) GetError() Error {
	return pp.err
}

func (pp *KDP) NotHaveError() bool {
	return pp.err.GetCode() == 0
}

func (pp *KDP) GetBody() []byte {
	return pp.bodyBuff
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
	return &KDP{
		heads: make(map[string]string),
	}
}
