package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"wsaetherfy/currency"
	"wsaetherfy/supabase"

	"github.com/gorilla/websocket"
	wsy "wsaetherfy/websocket"
	supa "github.com/supabase-community/supabase-go"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var supabaseClient *supa.Client

// Função para reiniciar a conexão com Supabase a cada 1 hora
func initializeAndRefreshSupabaseConnection() {
	for {
		var err error
		supabaseClient, err = supabase.InitializeDB()
		if err != nil {
			log.Fatalf("Erro ao inicializar o Supabase: %v", err)
		}
		log.Println("Conexão com Supabase reiniciada com sucesso.")
		
		// Espera por 1 hora antes de reiniciar a conexão
		time.Sleep(1 * time.Hour)
	}
}

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

	pair := string(message)
	_, exists := currency.GetAllCurrencyCodes()[pair]
	if !exists {
		log.Println("Par de moedas não encontrado:", pair)
		conn.WriteMessage(websocket.TextMessage, []byte("Par de moedas inválido"))
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
			err := conn.WriteJSON(output)
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
	if pair == "" {
		http.Error(w, "Par de moedas é obrigatório", http.StatusBadRequest)
		return
	}

	priceData, found := currency.GetPrices(pair)
	if !found {
		http.Error(w, "Par de moedas não encontrado", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"pair":      pair,
		"price":     priceData.Price,
		"timestamp": priceData.Timestamp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Inicializar e reiniciar a conexão Supabase a cada 1 hora em uma goroutine separada
	go initializeAndRefreshSupabaseConnection()

	// Inicializar monitoramento de todas as moedas
	go currency.MonitorAllCurrencies()

	// Configurar handlers HTTP
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/prices", priceHandler)

	// Inicializar servidor
	port := ":8081"
	fmt.Println("Servidor WebSocket e HTTP rodando na porta", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
