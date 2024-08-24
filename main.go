package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"wsaetherfy/currency"
	"wsaetherfy/supabase"
	wsy "wsaetherfy/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Inicializa o cliente Supabase globalmente
var supabaseClient, _ = supabase.InitializeDB()

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Obter a chave API do cabeçalho HTTP
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key is required", http.StatusUnauthorized)
		return
	}

	// Verificar a chave API usando o cliente Supabase
	valid, err := supabase.VerifyAPIKey(supabaseClient, apiKey)
	if err != nil {
		log.Println("Erro ao verificar a chave API:", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	if !valid {
		http.Error(w, "Chave API inválida", http.StatusUnauthorized)
		return
	}

	// Estabelecer a conexão WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erro ao criar conexão WebSocket:", err)
		return
	}
	defer conn.Close()

	// Ler a mensagem inicial do cliente que deve conter o par de moedas no formato "EUR/USD"
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Println("Erro ao ler mensagem do WebSocket:", err)
		return
	}

	// Converter a mensagem para string e buscar o código da moeda correspondente
	pair := string(message)
	subs, exists := currency.GetCurrencyCode(pair)
	if !exists {
		log.Println("Par de moedas não encontrado:", pair)
		conn.WriteMessage(websocket.TextMessage, []byte("Par de moedas inválido"))
		return
	}

	// Conectar-se ao WebSocket usando o código da moeda obtido
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

	// Enviar os dados do ticker para o cliente via WebSocket
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
