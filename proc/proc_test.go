package proc

import (
	"reflect"
	"testing"
	"time"
)

func TestNewChannelProcessor(t *testing.T) {
	cp := NewChannelProcessor(1000000000)
	if cp.store == nil {
		t.Error(`Store is not created`)
	}
	if cp.ttl != time.Second {
		t.Error(`ttl is incorrect: expected 1 but got &#v`, cp.ttl)
	}
}

func TestChannelProcessor_valueExistsAndNotOutdated(t *testing.T) {
	var ok bool
	cp := NewChannelProcessor(1000000000)
	cp.store[`foo`] = storedValue{
		value:   1,
		created: time.Now().Add(-time.Second * 2),
	}
	_, ok = cp.valueExistsAndNotOutdated(`foo`)
	if ok {
		t.Error(`Incorrest outdate value recognition`)
	}
	_, ok = cp.valueExistsAndNotOutdated(`moo`)
	if ok {
		t.Error(`Incorrest inexisting value recognition`)
	}
	cp.store[`foo`] = storedValue{
		value:   1,
		created: time.Now(),
	}
	_, ok = cp.valueExistsAndNotOutdated(`foo`)
	if !ok {
		t.Error(`Incorrest up-to-date value recognition`)
	}
}

func TestChannelProcessor_Create(t *testing.T) {
	var (
		cp  *ChannelProcessor
		err error
	)
	cp = NewChannelProcessor(1000000000)
	cp.Start()
	defer cp.Stop()
	err = cp.Create(`foo`, `bar`)
	if err != nil {
		t.Error(`Error on value creation`)
	}
	err = cp.Create(`foo`, 1)
	if err == nil {
		t.Error(`No error on existing value creation`)
	}
	v, err := cp.Get(`foo`)
	if err != nil {
		t.Error(`Error on value get`)
	}
	if v.(string) != `bar` {
		t.Error(`Incorrect value returned`)
	}
}

func TestChannelProcessor_Update(t *testing.T) {
	var (
		cp  *ChannelProcessor
		err error
	)
	cp = NewChannelProcessor(1000000000)
	cp.Start()
	defer cp.Stop()
	err = cp.Update(`foo`, 1)
	if err == nil {
		t.Error(`No error on inexisting value update`)
	}
	err = cp.Create(`foo`, `bar`)
	if err != nil {
		t.Error(`Error on value creation`)
	}
	err = cp.Update(`foo`, 1)
	if err != nil {
		t.Error(`Error on existing value update`)
	}
	v, err := cp.Get(`foo`)
	if err != nil {
		t.Error(`Error on value get`)
	}
	if v.(int) != 1 {
		t.Error(`Incorrect value returned`)
	}
}

func TestChannelProcessor_KeyList(t *testing.T) {
	var (
		cp  *ChannelProcessor
		err error
	)
	cp = NewChannelProcessor(1000000000)
	cp.Start()
	defer cp.Stop()
	err = cp.Update(`foo`, 1)
	if err == nil {
		t.Error(`No error on inexisting value update`)
	}
	err = cp.Create(`foo`, `bar`)
	if err != nil {
		t.Error(`Error on value creation`)
	}
	err = cp.Create(`meeeeeow`, `mooooo`)
	if err != nil {
		t.Error(`Error on value creation`)
	}
	err = cp.Create(`wtf`, 42)
	if err != nil {
		t.Error(`Error on value creation`)
	}
	err = cp.Update(`foo`, 1)
	if err != nil {
		t.Error(`Error on existing value update`)
	}
	keys, err := cp.KeyList()
	if err != nil {
		t.Error(`Error on value get`)
	}
	if !reflect.DeepEqual(keys, []string{`foo`, `meeeeeow`, `wtf`}) {
		t.Error(`Incorrect KeyList returned value`)
	}
}
