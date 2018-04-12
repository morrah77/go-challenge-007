package proc

import (
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

func TestChannelProcessor_SetTtl(t *testing.T) {
	var (
		cp  *ChannelProcessor
		err error
	)
	cp = NewChannelProcessor(1000000000)
	cp.Start()
	defer cp.Stop()
	err = cp.SetTtl(`foo`, 1)
	if err == nil {
		t.Error(`No error on inexisting value TTL set`)
	}
	err = cp.Create(`foo`, `bar`)
	if err != nil {
		t.Error(`Error on value creation`)
	}
	err = cp.SetTtl(`foo`, `meeeeooowww`)
	if err == nil {
		t.Error(`No error on incorrect TTL set`)
	}
	err = cp.SetTtl(`foo`, `10s`)
	if err != nil {
		t.Errorf(`Error on existing value correct TTL set with time string, got %#v`, err)
	}
	// a great opportunity of tests stored in the same package, other way is to have additional exposed method
	if cp.store[`foo`].Ttl != time.Second*10 {
		t.Errorf(`Incorrect TTL set with time string, got %#v`, cp.store[`foo`].Ttl)
	}
	err = cp.SetTtl(`foo`, `1`)
	if err != nil {
		t.Errorf(`Error on existing value correct TTL set with digit-only string, got %#v`, err)
	}
	if cp.store[`foo`].Ttl != time.Second {
		t.Errorf(`Incorrect TTL set with digit-only string, got %#v`, cp.store[`foo`].Ttl)
	}
	err = cp.SetTtl(`foo`, time.Millisecond)
	if err != nil {
		t.Errorf(`Error on existing value correct TTL set with integer, got %#v`, err)
	}
	if cp.store[`foo`].Ttl != time.Millisecond {
		t.Errorf(`Incorrect TTL set with integer, got %#v`, cp.store[`foo`].Ttl)
	}
	v, err := cp.Get(`foo`)
	if err != nil {
		t.Error(`Error on value get`)
	}
	if v.(string) != `bar` {
		t.Error(`Incorrect value returned`)
	}
	time.Sleep(time.Millisecond * 2)
	v, err = cp.Get(`foo`)
	if err == nil {
		t.Error(`Error on outdated value get`)
	}
}

func TestChannelProcessor_Remove(t *testing.T) {
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
	err = cp.Create(`meeeeeow`, `mooooo`)
	if err != nil {
		t.Error(`Error on value creation`)
	}
	v, err := cp.Get(`foo`)
	if err != nil {
		t.Error(`Error on value get`)
	}
	if v.(string) != `bar` {
		t.Error(`Incorrect value returned`)
	}
	err = cp.Remove(`foo`)
	if err != nil {
		t.Error(`Error on Remove existing key`)
	}
	v, err = cp.Get(`foo`)
	if err == nil {
		t.Error(`Did not remove key`)
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
		t.Error(`Error on KeyList`)
	}
	if !compareSrtingSlice(keys, []string{`foo`, `meeeeeow`, `wtf`}) {
		t.Errorf(`Incorrect KeyList returned value, got %#v`, keys)
	}
}

func compareSrtingSlice(got []string, expected []string) bool {
	if len(got) != len(expected) {
		return false
	}
	m := 0
	for _, s := range got {
		for j, z := range expected {
			if s == z {
				m++
				if j == len(expected)-1 {
					expected = expected[:j]
				} else {
					expected = append(expected[:j], expected[j+1:]...)
				}
				break
			}
		}
	}
	return len(got) == m
}
