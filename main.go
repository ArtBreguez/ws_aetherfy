package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	wsy "wsaetherfy/websocket" 
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erro ao criar conexão WebSocket:", err)
		return
	}
	defer conn.Close()

	// Inicialize a conexão com o ticker
	subs := "EURUSD=X" // Ou qualquer outro ticker desejado
	yf := wsy.NewWithSub(subs)
	if err := yf.Connect(); err != nil {
		log.Println("Erro ao conectar:", err)
		return
	}
	defer yf.Close()

	if err := yf.Subscribe(); err != nil {
		log.Println("Erro ao assinar:", err)
		return
	}

	ticker, err := yf.Ticker()
	if err != nil {
		log.Println("Erro ao obter ticker:", err)
		return
	}

	for {
		select {
		case output := <-ticker:
			err := conn.WriteJSON(output)
			if err != nil {
				log.Println("Erro ao enviar mensagem via WebSocket:", err)
				return
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	port := ":8080"
	fmt.Println("Servidor WebSocket rodando na porta", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
