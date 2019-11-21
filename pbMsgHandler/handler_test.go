package pbMsgHandler

import (
	"github.com/graydream/YTHost/pb"
	"testing"
)

func TestRegister(t *testing.T){
	var msg pb.StringMsg
	msg.Value = "111"
	(&MsgRegister{}).Register(&msg)
}
