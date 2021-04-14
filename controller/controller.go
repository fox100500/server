package controller

import (
	"errors"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

//Listeners ...
//type Listeners []*websocket.Conn
type Event string

//
const (
	EventInfo      Event = "eventinfo"
	EventCandle    Event = "eventcandle"
	EventOrderBook Event = "eventorderbook"
)

//
type FIGI string

type WebSocArray []*websocket.Conn

//
type Listeners map[FIGI]WebSocArray

//Register ...
type Register struct {
	registerID string
	Sources    map[Event]Listeners
	sendOk     int64
	sendErr    int64
}

//New ...
func New(registerID string) (ctx *Register) {
	ctx = &Register{
		registerID: registerID,
	}

	ctx.Sources = make(map[Event]Listeners)

	ctx.sendOk = 0
	ctx.sendErr = 0

	return ctx
}

func recoverFunc() {
	fmt.Println("Check recover ... ")
	if recoverMsg := recover(); recoverMsg != nil {
		fmt.Println("done", recoverMsg)
	} else {
		fmt.Println("normal exit.")
	}
}

//AddListenerByFigi ...
func (ctx *Register) AddListenerByFigi(eventName string, figi string, wsID *websocket.Conn) {
	defer recoverFunc()

	if _, ok := ctx.Sources[Event(eventName)]; !ok {
		listener := make(map[FIGI]WebSocArray)
		listener = map[FIGI]WebSocArray{}
		ctx.Sources[Event(eventName)] = listener
	}

	listener := ctx.Sources[Event(eventName)]

	wsList := listener[FIGI(figi)]

	if len(wsList) > 0 {
		for _, value := range wsList {
			if value == wsID {
				fmt.Println("AddListener wsID not added")
				return
			}
		}
	}

	listener[FIGI(figi)] = append(listener[FIGI(figi)], wsID)
	fmt.Println("AddListener", eventName, figi, len(listener[FIGI(figi)]))
}

//DispatchEvent ...
func (ctx *Register) DispatchEvent(name string, figi string, event interface{}) error {

	source, ok := ctx.Sources[Event(name)]
	if !ok {
		log.Println("DispatchEvent ", name, "not found")
		fmt.Println("DispatchEvent ", name, "not found")
		return errors.New("event not found")
	}
	listeners, ok := source[FIGI(figi)]
	if !ok {
		log.Println("DispatchEvent ", figi, "in", name, "not found")
		fmt.Println("DispatchEvent ", figi, "in", name, "not found")
		return errors.New("figi not found")
	}

	if len(listeners) > 0 {
		for key, listener := range listeners {
			//fmt.Println("Try send...", name)
			if err := listener.WriteJSON(event); err != nil {
				ctx.sendErr++
				log.Println("DispatchEvent key=", key, "cnt=", ctx.sendErr, err)
				//fmt.Println("DispatchEvent key=", key, "cnt=", ctx.sendErr, err)
				ctx.DeleteListener(name, figi, listener)
			} else {
				ctx.sendOk++
				//log.Println("DispatchEvent Send key=", key, "cnt=", ctx.sendOk, "done")
				//fmt.Println("DispatchEvent Send key=", key, "cnt=", ctx.sendOk, "done")

			}
		}
		return nil
	}
	log.Println("DispatchEvent", name, figi, "No listeners ")
	fmt.Println("DispatchEvent", name, figi, "No listeners ")
	return errors.New("no listeners")
}

//DeleteListener ...
func (ctx *Register) DeleteListener(eventName string, figi string, wsID *websocket.Conn) {

	source, ok := ctx.Sources[Event(eventName)]
	if !ok {
		log.Println("DeleteListener", eventName, "not found")
		fmt.Println("DeleteListener", eventName, "not found")
		return
	}

	listeners, ok := source[FIGI(figi)]
	if !ok {
		log.Println("DeleteListener", figi, "in", eventName, "not found")
		fmt.Println("DeleteListener", figi, "in", eventName, "not found")
		return
	}

	lenListeners := len(listeners)

	log.Println("Del Listener", eventName, figi, lenListeners)
	fmt.Println("Del Listener", eventName, figi, lenListeners)
	if lenListeners > 0 {
		for i := range listeners {
			if listeners[i] == wsID {
				if lenListeners > 1 {
					ctx.Sources[Event(eventName)][FIGI(figi)][i] = ctx.Sources[Event(eventName)][FIGI(figi)][lenListeners-1]
					ctx.Sources[Event(eventName)][FIGI(figi)] = ctx.Sources[Event(eventName)][FIGI(figi)][:lenListeners-1]
					log.Println("Del Listener result: new size", len(ctx.Sources[Event(eventName)][FIGI(figi)]))
					fmt.Println("Del Listener result: new size", len(ctx.Sources[Event(eventName)][FIGI(figi)]))
					return
				}
				ctx.Sources[Event(eventName)][FIGI(figi)][i] = nil
				ctx.Sources[Event(eventName)][FIGI(figi)] = ctx.Sources[Event(eventName)][FIGI(figi)][:lenListeners-1]
				log.Println("Del Listener result: new size", len(ctx.Sources[Event(eventName)][FIGI(figi)]))
				fmt.Println("Del Listener result: new size", len(ctx.Sources[Event(eventName)][FIGI(figi)]))
				return
			}
		}
		log.Println("Del Listener wsID not found ")
		fmt.Println("Del Listener wsID not found ")
		return
	}
	log.Println("Del Listener lenListeners==0")
	fmt.Println("Del Listener lenListeners==0")

	delete(ctx.Sources[Event(eventName)], FIGI(figi))

	listeners, ok = ctx.Sources[Event(eventName)][FIGI(figi)]
	if !ok {
		log.Println("DeleteListener check delete", figi, "in", eventName, "not found")
		fmt.Println("DeleteListener check delete", figi, "in", eventName, "not found")
		return
	}
}
