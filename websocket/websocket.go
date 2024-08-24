package websocket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"wsaetherfy/yatickerpb"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type YahooFinance struct {
	ws        *websocket.Conn
	sub       string
	connected bool
	done      chan bool
}

func New() *YahooFinance {
	return &YahooFinance{}
}

func NewWithSub(subs string) *YahooFinance {
	return &YahooFinance{
		sub: subs,
	}
}

func (yf *YahooFinance) Connect() error {
	u := url.URL{
		Scheme: "wss",
		Host:   "streamer.finance.yahoo.com",
		Path:   "/",
	}

	var err error

	yf.ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("Fail to Dial: ", err)
		return err
	}
	yf.connected = true
	return nil
}

func (yf *YahooFinance) Close() error {
	if !yf.connected {
		return nil
	}
	yf.done = make(chan bool, 2)
	yf.done <- true
	yf.ws.Close()
	yf.connected = false
	return nil
}

func (yf *YahooFinance) Subscribe(subs ...string) error {
	if !yf.connected {
		if err := yf.Connect(); err != nil {
			fmt.Println("Fail to connect to WS: ", err)
			return err
		}
	}
	subs = append(subs, yf.sub)
	body := map[string][]string{
		"subscribe": subs,
	}
	message, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Fail to Unmarshal Body: ", err)
		return err
	}

	err = yf.ws.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		fmt.Println("Fail to Write Message: ", err)
		return err
	}
	return nil
}

func (yf *YahooFinance) Ticker() (chan *yatickerpb.Yaticker, error) {
	if !yf.connected {
		if err := yf.Connect(); err != nil {
			fmt.Println("Fail to connect to WS: ", err)
			return nil, err
		}
	}
	output := make(chan *yatickerpb.Yaticker, 10)
	go func() {
		defer close(output)
		for {
			if err := yf.readMessages(output); err != nil {
				fmt.Println("Fail to Read Message: ", err)
				// Tenta reconectar
				yf.Close()
				time.Sleep(1 * time.Second)
				if err := yf.Connect(); err != nil {
					fmt.Println("Fail to reconnect to WS: ", err)
					break
				}
				yf.Subscribe()
				fmt.Println("WS Conection Restored")
			}
		}
	}()
	return output, nil
}

func (yf *YahooFinance) readMessages(output chan *yatickerpb.Yaticker) error {
	for {
		select {
		case <-yf.done:
			return nil
		default:
			_, message, err := yf.ws.ReadMessage()
			if err != nil {
				return err
			} else {
				decodedMessage := make([]byte, base64.StdEncoding.DecodedLen(len(message)))
				l, err := base64.StdEncoding.Decode(decodedMessage, message)
				if err != nil {
					fmt.Println("Fail to Decode Message: ", err)
					continue
				}

				yaticker := &yatickerpb.Yaticker{}
				if err := proto.Unmarshal(decodedMessage[:l], yaticker); err != nil {
					fmt.Println("Fail to Unmarshal Message: ", err)
					continue
				}

				output <- yaticker
			}
		}
	}
}
