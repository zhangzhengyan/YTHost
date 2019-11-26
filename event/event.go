package event

import (
	"github.com/google/uuid"
)

type Event struct {
	Name string
	Msg  interface{}
}

type EventHandleFunc func(evt Event)

type EventTrigger struct {
	eventMap map[string]map[uuid.UUID]EventHandleFunc
}

func (et *EventTrigger) Subscribe(name string, handler EventHandleFunc) uuid.UUID {

	if et.eventMap == nil {
		et.eventMap = make(map[string]map[uuid.UUID]EventHandleFunc)
	}
	if _, ok := et.eventMap[name]; !ok {
		et.eventMap[name] = make(map[uuid.UUID]EventHandleFunc)
	}
	id := uuid.New()
	et.eventMap[name][id] = handler
	return id
}

func (et *EventTrigger) Unsubscribe(id uuid.UUID) {
	for _, hdmap := range et.eventMap {
		delete(hdmap, id)
	}
}

func (et *EventTrigger) Emit(evt Event) {
	if hds, ok := et.eventMap[evt.Name]; ok {
		for _, handler := range hds {
			go handler(evt)
		}
	}
}
