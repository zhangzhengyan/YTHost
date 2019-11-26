package event

import (
	"fmt"
	"testing"
)

func TestEventTrigger_Emit(t *testing.T) {
	et := new(EventTrigger)
	et.Subscribe("test", func(evt Event) {
		fmt.Println("111", evt.Name, evt.Msg.(string))
	})
	id := et.Subscribe("test", func(evt Event) {
		fmt.Println("222", evt.Name, evt.Msg.(string))
	})
	et.Subscribe("test", func(evt Event) {
		fmt.Println("333", evt.Name, evt.Msg.(string))
	})
	et.Emit(Event{"test", "测试"})
	et.Unsubscribe(id)
	et.Emit(Event{"test", "测试2"})
}
