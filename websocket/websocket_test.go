package websocket

import "testing"

func TestYFTicker(t *testing.T) {
	subs := "EURUSD=X"
	yf := NewWithSub(subs)
	if err := yf.Connect(); err != nil {
		t.Fail()
		return
	}
	if err := yf.Subscribe(); err != nil {
		t.Fail()
		return
	}
	ticker, err := yf.Ticker()
	if err != nil {
		t.Fail()
		return
	}
	output := <-ticker
	yf.Close()
	if output.Id == "" {
		t.Fail()
		return
	}
}
