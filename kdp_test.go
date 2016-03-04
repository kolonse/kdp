package kdp

import (
	"testing"
)

func TestKDP(t *testing.T) {
	k := NewKDP()
	err := k.Parse([]byte("fdsafsdf")).GetError()
	if err.GetCode() != KDP_PROTO_ERROR_FORMAT {
		t.Error("错误应该是 长度错误")
	} else {
		t.Log(err.Error())
	}
}
