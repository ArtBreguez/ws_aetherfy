package currency

import (
	"log"
	"sync"
	"wsaetherfy/websocket"
)

var currencyMap = map[string]string{
	"EUR/USD": "EURUSD=X",
	"USD/JPY": "JPY=X",
	"GBP/USD": "GBPUSD=X",
	"AUD/USD": "AUDUSD=X",
	"NZD/USD": "NZDUSD=X",
	"EUR/JPY": "EURJPY=X",
	"GBP/JPY": "GBPJPY=X",
	"EUR/GBP": "EURGBP=X",
	"EUR/CAD": "EURCAD=X",
	"EUR/SEK": "EURSEK=X",
	"EUR/CHF": "EURCHF=X",
	"EUR/HUF": "EURHUF=X",
	"USD/CNY": "CNY=X",
	"USD/HKD": "HKD=X",
	"USD/SGD": "SGD=X",
	"USD/INR": "INR=X",
	"USD/MXN": "MXN=X",
	"USD/PHP": "PHP=X",
	"USD/IDR": "IDR=X",
	"USD/THB": "THB=X",
	"USD/MYR": "MYR=X",
	"USD/ZAR": "ZAR=X",
	"USD/RUB": "RUB=X",
	"BTC/USD": "BTC-USD",
	"ETH/USD": "ETH-USD",
	"USDT/USD": "USDT-USD",
	"BNB/USD": "BNB-USD",
	"SOL/USD": "SOL-USD",
	"USDC/USD": "USDC-USD",
	"XRP/USD": "XRP-USD",
	"STETH/USD": "STETH-USD",
	"DOGE/USD": "DOGE-USD",
	"TON11419/USD": "TON11419-USD",
	"ADA/USD": "ADA-USD",
	"WTRX/USD": "WTRX-USD",
	"TRX/USD": "TRX-USD",
	"WSTETH/USD": "WSTETH-USD",
	"AVAX/USD": "AVAX-USD",
	"WBTC/USD": "WBTC-USD",
	"WETH/USD": "WETH-USD",
	"SHIB/USD": "SHIB-USD",
	"LINK/USD": "LINK-USD",
	"DOT/USD": "DOT-USD",
	"BCH/USD": "BCH-USD",
	"EDLC/USD": "EDLC-USD",
	"MATIC/USD": "MATIC-USD",
	"NEAR/USD": "NEAR-USD",
	"LEO/USD": "LEO-USD",
}

func GetCurrencyCode(pair string) (string, bool) {
	code, exists := currencyMap[pair]
	return code, exists
}

func GetAllCurrencyCodes() map[string]string {
	return currencyMap
}

type PriceData struct {
	Price     float64
	Timestamp int64
}

type PriceStore struct {
	sync.RWMutex
	prices map[string][]PriceData
}

var (
	priceStore = &PriceStore{prices: make(map[string][]PriceData)}
)

// Função para monitorar todas as moedas
func MonitorAllCurrencies() {
	for pair, subs := range GetAllCurrencyCodes() {
		go func(pair, subs string) {
			yf := websocket.NewWithSub(subs)
			if err := yf.Connect(); err != nil {
				log.Printf("Erro ao conectar para %s: %v", pair, err)
				return
			}
			defer yf.Close()

			if err := yf.Subscribe(); err != nil {
				log.Printf("Erro ao assinar %s: %v", pair, err)
				return
			}

			ticker, err := yf.Ticker()
			if err != nil {
				log.Printf("Erro ao obter ticker para %s: %v", pair, err)
				return
			}

			for {
				select {
				case output := <-ticker:
					priceStore.Lock()
					priceData := PriceData{
						Price:     float64(output.Price),
						Timestamp: output.Time,  // Adicionando o timestamp
					}
					if prices, found := priceStore.prices[pair]; found {
						priceStore.prices[pair] = append(prices, priceData)
					} else {
						priceStore.prices[pair] = []PriceData{priceData}
					}
					priceStore.Unlock()
				}
			}
			
		}(pair, subs)
	}
}

func GetPrices(pair string) (PriceData, bool) {
	priceStore.RLock()
	defer priceStore.RUnlock()
	prices, found := priceStore.prices[pair]
	if !found || len(prices) == 0 {
		return PriceData{}, false
	}
	return prices[len(prices)-1], found  // Retornar o último preço com timestamp
}

