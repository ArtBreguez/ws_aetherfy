package currency

import (
	"log"
	"sync"
	"time"
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

var ValidTimeframes = map[string]bool{
	"tick": true,
	"m1":   true,
	"m5":   true,
	"m15":  true,
	"m30":  true,
	"h1":   true,
	"h4":   true,
	"D":    true,
}

// Estrutura para armazenar preços
type PriceStore struct {
	sync.RWMutex
	prices map[string][]TickData // [pair][]TickData
}

// Estrutura para armazenar dados de tick
type TickData struct {
	Price     float64
	Timestamp time.Time
}

var (
	priceStore = &PriceStore{prices: make(map[string][]TickData)}
)

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
					priceStore.prices[pair] = append(priceStore.prices[pair], TickData{
						Price:     float64(output.Price),
						Timestamp: time.Now(),
					})
					priceStore.Unlock()
				}
			}
		}(pair, subs)
	}
}

func GetPrices(pair string, timeframe string) ([]float64, bool) {
	priceStore.RLock()
	defer priceStore.RUnlock()

	ticks, found := priceStore.prices[pair]
	if !found || len(ticks) == 0 {
		return nil, false
	}

	// Se o timeframe for "tick", retornamos o último preço diretamente
	if timeframe == "tick" {
		lastTick := ticks[len(ticks)-1]
		return []float64{lastTick.Price}, true
	}

	// Caso contrário, calculamos o OHLC para o timeframe solicitado
	ohlc := calculateOHLC(ticks, timeframe)
	return ohlc, true
}

func calculateOHLC(ticks []TickData, timeframe string) []float64 {
	var open, high, low, close float64

	if len(ticks) == 0 {
		return nil
	}

	alignedTime := alignToTimeframe(ticks[0].Timestamp, timeframe)
	open = ticks[0].Price
	high = ticks[0].Price
	low = ticks[0].Price
	close = ticks[0].Price

	for _, tick := range ticks {
		if alignToTimeframe(tick.Timestamp, timeframe) != alignedTime {
			break
		}

		if tick.Price > high {
			high = tick.Price
		}
		if tick.Price < low {
			low = tick.Price
		}
		close = tick.Price
	}

	return []float64{open, high, low, close}
}

// Função para determinar o timestamp inicial do candle com base no timeframe
func alignToTimeframe(timestamp time.Time, timeframe string) time.Time {
	switch timeframe {
	case "m1":
		return timestamp.Truncate(time.Minute)
	case "m5":
		return timestamp.Truncate(5 * time.Minute)
	case "m15":
		return timestamp.Truncate(15 * time.Minute)
	case "m30":
		return timestamp.Truncate(30 * time.Minute)
	case "h1":
		return timestamp.Truncate(time.Hour)
	case "h4":
		// Trunca para as 00h, 04h, 08h, 12h, 16h, ou 20h
		h := timestamp.Hour() / 4 * 4
		return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), h, 0, 0, 0, timestamp.Location())
	case "D":
		return timestamp.Truncate(24 * time.Hour)
	default:
		// Para timeframes não reconhecidos, retornamos o timestamp original
		return timestamp
	}
}

func GetCurrencyCode(pair string) (string, bool) {
	code, exists := currencyMap[pair]
	return code, exists
}

func GetAllCurrencyCodes() map[string]string {
	return currencyMap
}