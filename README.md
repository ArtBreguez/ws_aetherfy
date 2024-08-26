API

import requests

def test_price_api(api_key, pair):
    # URL do endpoint de preços
    url = "http://localhost:8080/prices"

    # Headers com a chave da API
    headers = {
        "X-API-Key": api_key
    }

    # Parâmetros da requisição
    params = {
        "pair": pair
    }

    # Fazer a requisição GET para o endpoint de preços
    response = requests.get(url, headers=headers, params=params)

    # Verificar o status da resposta
    if response.status_code == 200:
        # Se a resposta for bem-sucedida, exibir os dados
        data = response.json()
        print("Dados recebidos:", data)
    else:
        # Caso contrário, exibir o erro
        print("Erro ao obter preços:", response.status_code, response.text)

if __name__ == "__main__":
    # Substitua por sua chave API e par de moedas desejado
    api_key = "sk_test_wbuumnff6b_5jirnl7h8l9"
    pair = "GBP/USD"

    # Testar a API de preços
    test_price_api(api_key, pair)


WS

import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    print("Received ticker data:", data)

def on_error(ws, error):
    print("Error:", error)

def on_close(ws, close_status_code, close_msg):
    print("Connection closed")

def on_open(ws):
    print("Connection established")
    # Enviar o par de moedas "BTC/USD" ao servidor WebSocket
    ws.send("BTC/USD")

if __name__ == "__main__":
    # URL do WebSocket server
    ws_url = "ws://localhost:8080/ws"

    # Cabeçalho de autenticação com a chave API
    headers = {
        "X-API-Key": "sk_test_wbuumnff6b_5jirnl7h8l9"  # Substitua por sua chave API real
    }

    # Criar conexão WebSocket
    ws = websocket.WebSocketApp(ws_url,
                                header=headers,
                                on_open=on_open,
                                on_message=on_message,
                                on_error=on_error,
                                on_close=on_close)

    # Iniciar conexão WebSocket
    ws.run_forever()
