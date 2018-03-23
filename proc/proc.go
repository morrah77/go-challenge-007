// Concurrent in-memory key-value store
// Helen Lazar (https://github.com/morrah77) 2018-03-24
package proc

import (
	"errors"
	"time"
)

const channelsCapacity = 10

type inputValue struct {
	key    string
	value  interface{}
	chResp chan outputValue
}

type outputValue struct {
	value interface{}
	err   error
}

type storedValue struct {
	value   interface{}
	created time.Time
}

var ErrorKeyExists = errors.New(`Key already exists`)
var ErrorKeyNotExists = errors.New(`Key does not exist`)

type ChannelProcessor struct {
	chCreate chan inputValue
	chGet    chan inputValue
	chUpdate chan inputValue
	chRemove chan inputValue
	chList   chan inputValue
	store    map[string]storedValue
	ttl      time.Duration
}

func NewChannelProcessor(ttl time.Duration) *ChannelProcessor {
	return &ChannelProcessor{
		chCreate: make(chan inputValue, channelsCapacity),
		chUpdate: make(chan inputValue, channelsCapacity),
		chGet:    make(chan inputValue, channelsCapacity),
		chRemove: make(chan inputValue, channelsCapacity),
		chList:   make(chan inputValue, channelsCapacity),
		store:    make(map[string]storedValue),
		ttl:      ttl,
	}
}

func (cp *ChannelProcessor) Start() {
	go func() {
		for {
			select {
			case v := <-cp.chCreate:
				if _, ok := cp.valueExistsAndNotOutdated(v.key); ok {
					v.chResp <- outputValue{
						value: nil,
						err:   ErrorKeyExists,
					}
				} else {
					cp.store[v.key] = storedValue{
						value:   v.value,
						created: time.Now(),
					}
					v.chResp <- outputValue{
						value: nil,
						err:   nil,
					}
				}
			case v := <-cp.chUpdate:
				if _, ok := cp.valueExistsAndNotOutdated(v.key); !ok {
					v.chResp <- outputValue{
						value: nil,
						err:   ErrorKeyNotExists,
					}
				} else {
					cp.store[v.key] = storedValue{
						value:   v.value,
						created: time.Now(),
					}
					v.chResp <- outputValue{
						value: nil,
						err:   nil,
					}
				}
			case v := <-cp.chRemove:
				if _, ok := cp.valueExistsAndNotOutdated(v.key); !ok {
					v.chResp <- outputValue{
						value: nil,
						err:   ErrorKeyNotExists,
					}
				} else {
					delete(cp.store, v.key)
					v.chResp <- outputValue{
						value: nil,
						err:   nil,
					}
				}
			case v := <-cp.chGet:
				value, ok := cp.valueExistsAndNotOutdated(v.key)
				if !ok {
					v.chResp <- outputValue{
						value: nil,
						err:   ErrorKeyNotExists,
					}
				} else {
					v.chResp <- outputValue{
						value: value.value,
						err:   nil,
					}
				}
			case v := <-cp.chList:
				var keys []string
				for key, sv := range cp.store {
					if !sv.created.Add(cp.ttl).After(time.Now()) {
						delete(cp.store, key)
					} else {
						keys = append(keys, key)
					}
				}
				v.chResp <- outputValue{
					value: keys,
					err:   nil,
				}
			}
		}
	}()
}

func (cp *ChannelProcessor) valueExistsAndNotOutdated(key string) (storedValue, bool) {
	sv, ok := cp.store[key]
	if !ok {
		return sv, false
	}
	if !sv.created.Add(cp.ttl).After(time.Now()) {
		delete(cp.store, key)
		return sv, false
	}
	return sv, true
}

func (cp *ChannelProcessor) Create(key string, value interface{}) error {
	input := inputValue{
		key:    key,
		value:  value,
		chResp: make(chan outputValue),
	}
	cp.chCreate <- input
	resp := <-input.chResp
	return resp.err
}

func (cp *ChannelProcessor) Get(key string) (interface{}, error) {
	input := inputValue{
		key:    key,
		chResp: make(chan outputValue),
	}
	cp.chGet <- input
	resp := <-input.chResp
	return resp.value, resp.err
}

func (cp *ChannelProcessor) Update(key string, value interface{}) error {
	input := inputValue{
		key:    key,
		value:  value,
		chResp: make(chan outputValue),
	}
	cp.chUpdate <- input
	resp := <-input.chResp
	return resp.err
}

func (cp *ChannelProcessor) Remove(key string) error {
	input := inputValue{
		key:    key,
		chResp: make(chan outputValue),
	}
	cp.chRemove <- input
	resp := <-input.chResp
	return resp.err
}

func (cp *ChannelProcessor) KeyList() ([]string, error) {
	input := inputValue{
		chResp: make(chan outputValue),
	}
	cp.chList <- input
	println(`List will wait for resp`)
	resp := <-input.chResp
	return resp.value.([]string), resp.err
}

func (cp *ChannelProcessor) Stop() {
	close(cp.chCreate)
	close(cp.chUpdate)
	close(cp.chRemove)
	close(cp.chGet)
	close(cp.chList)
}
