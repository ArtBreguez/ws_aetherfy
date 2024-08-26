package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"wsaetherfy/currency"
	"wsaetherfy/supabase"
	wsy "wsaetherfy/websocket"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var supabaseClient, _ = supabase.InitializeDB()

func wsHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key is required", http.StatusUnauthorized)
		return
	}

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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erro ao criar conexão WebSocket:", err)
		return
	}
	defer conn.Close()

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Println("Erro ao ler mensagem do WebSocket:", err)
		return
	}

	var request map[string]string
	if err := json.Unmarshal(message, &request); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Erro ao decodificar o JSON"))
		return
	}

	pair, ok := request["pair"]
	if !ok {
		conn.WriteMessage(websocket.TextMessage, []byte("Par de moedas é obrigatório"))
		return
	}

	timeframe, ok := request["timeframe"]
	if !ok {
		timeframe = "tick"
	}

	if _, exists := currency.GetAllCurrencyCodes()[pair]; !exists {
		conn.WriteMessage(websocket.TextMessage, []byte("Par de moedas inválido"))
		return
	}

	if _, valid := currency.ValidTimeframes[timeframe]; !valid {
		conn.WriteMessage(websocket.TextMessage, []byte("Timeframe inválido"))
		return
	}

	yf := wsy.NewWithSub(currency.GetAllCurrencyCodes()[pair])
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
			priceData := map[string]interface{}{
				"price":     output.Price,
				"timestamp": time.Now().UTC(),
			}
			err := conn.WriteJSON(priceData)
			if err != nil {
				log.Println("Erro ao enviar mensagem via WebSocket:", err)
				return
			}
		}
	}
}

func priceHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		http.Error(w, "API key is required", http.StatusUnauthorized)
		return
	}

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

	pair := r.URL.Query().Get("pair")
	timeframe := r.URL.Query().Get("timeframe")
	if pair == "" {
		http.Error(w, "Par de moedas é obrigatório", http.StatusBadRequest)
		return
	}

	if timeframe == "" {
		timeframe = "tick"
	}

	prices, found := currency.GetPrices(pair, timeframe)
	if !found {
		http.Error(w, "Par de moedas ou timeframe não encontrado", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"pair":      pair,
		"timeframe": timeframe,
		"prices":    prices,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Inicializar monitoramento de todas as moedas
	go currency.MonitorAllCurrencies()

	// Configurar handlers HTTP
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/prices", priceHandler)

	// Inicializar servidor
	port := ":8080"
	fmt.Println("Servidor WebSocket e HTTP rodando na porta", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
