package supabase

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	supa "github.com/supabase-community/supabase-go"
)

func InitializeDB() (*supa.Client, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
		return nil, err
	}

	apiURL := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_API_KEY")
	if apiURL == "" || apiKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL ou SUPABASE_API_KEY não estão definidas")
	}

	client, err := supa.NewClient(apiURL, apiKey, nil)
	if err != nil {
		return nil, fmt.Errorf("não foi possível inicializar o cliente Supabase: %v", err)
	}

	_, err = client.SignInWithEmailPassword(os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD"))
	if err != nil {
		return nil, fmt.Errorf("falha ao autenticar: %v", err)
	}

	return client, nil
}

func VerifyAPIKey(client *supa.Client, apiKey string) (bool, error) {
    data, _, err := client.From("api_keys").Select("api_key", "exact", false).Eq("api_key", apiKey).Execute()
    if err != nil {
        fmt.Println(err.Error())
        if err.Error() == "JWT expired" {
            err = refreshSession(client)
            if err != nil {
                return false, fmt.Errorf("erro ao renovar a sessão: %v", err)
            }
            data, _, err = client.From("api_keys").Select("api_key", "exact", false).Eq("api_key", apiKey).Execute()
        }
        if err != nil {
            return false, fmt.Errorf("erro ao consultar a tabela api_keys: %v", err)
        }
    }

    var result []map[string]interface{}
    err = json.Unmarshal(data, &result)
    if err != nil {
        return false, fmt.Errorf("erro ao decodificar a resposta JSON: %v", err)
    }

    if len(result) > 0 {
        return true, nil
    }

    return false, nil
}

func GetUserIdByApiKey(client *supa.Client, apiKey string) (string, error) {
	data, _, err := client.From("api_keys").Select("id", "exact", false).Eq("api_key", apiKey).Execute()
	if err != nil {
        if err.Error() == "JWT expired" {
            err = refreshSession(client)
            if err != nil {
                return "", fmt.Errorf("erro ao renovar a sessão: %v", err)
            }
            data, _, err = client.From("api_keys").Select("id", "exact", false).Eq("api_key", apiKey).Execute()
        }
        if err != nil {
            return "", fmt.Errorf("erro ao consultar a tabela api_keys: %v", err)
        }
	}

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar a resposta JSON: %v", err)
	}

	if len(result) > 0 {
		return result[0]["id"].(string), nil
	}
	
	return "", nil
}

func GetApiUsageByUserId(client *supa.Client, userId string) (int, int, error) {
	data, _, err := client.From("api_usage").Select("*", "exact", false).Eq("id", userId).Execute()
	if err != nil {
        if err.Error() == "JWT expired" {
            err = refreshSession(client)
            if err != nil {
                return 0, 0, fmt.Errorf("erro ao renovar a sessão: %v", err)
            }
            data, _, err = client.From("api_usage").Select("*", "exact", false).Eq("id", userId).Execute()
        }
        if err != nil {
            return 0, 0, fmt.Errorf("erro ao consultar a tabela api_usage: %v", err)
        }
	}

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return 0, 0, fmt.Errorf("erro ao decodificar a resposta JSON: %v", err)
	}

	if len(result) > 0 {
		return int(result[0]["max_api_calls"].(float64)), int(result[0]["current_api_calls"].(float64)), nil
	}

	return 0, 0, nil
}

func UpdateApiUserUsage(client *supa.Client, userId string, apiCalls int) error {
	_, _, err := client.From("api_usage").Update(map[string]interface{}{
		"current_api_calls": apiCalls,
	}, "", "").Eq("id", userId).Execute()
	if err != nil {
        if err.Error() == "JWT expired" {
            err = refreshSession(client)
            if err != nil {
                return fmt.Errorf("erro ao renovar a sessão: %v", err)
            }
            _, _, err = client.From("api_usage").Update(map[string]interface{}{
                "current_api_calls": apiCalls,
            }, "", "").Eq("id", userId).Execute()
        }
        if err != nil {
            return fmt.Errorf("erro ao atualizar a tabela api_usage: %v", err)
        }
	}

	return nil
}

func SelectAllIdsFromApiUsage(client *supa.Client) ([]string, error) {
	data, _, err := client.From("api_usage").Select("id", "exact", false).Execute()
	if err != nil {
        if err.Error() == "JWT expired" {
            err = refreshSession(client)
            if err != nil {
                return nil, fmt.Errorf("erro ao renovar a sessão: %v", err)
            }
            data, _, err = client.From("api_usage").Select("id", "exact", false).Execute()
        }
        if err != nil {
            return nil, fmt.Errorf("erro ao consultar a tabela api_usage: %v", err)
        }
	}

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar a resposta JSON: %v", err)
	}

	ids := make([]string, len(result))
	for i, row := range result {
		ids[i] = row["id"].(string)
	}

	return ids, nil
}

func ResetApiUsage(client *supa.Client) error {
	ids, err := SelectAllIdsFromApiUsage(client)
	if err != nil {
		return fmt.Errorf("erro ao selecionar todos os IDs da tabela api_usage: %v", err)
	}

	for _, id := range ids {
		_, _, err = client.From("api_usage").Update(map[string]interface{}{
			"current_api_calls": 0,
		}, "", "").Eq("id", id).Execute()
		if err != nil {
            if err.Error() == "JWT expired" {
                err = refreshSession(client)
                if err != nil {
                    return fmt.Errorf("erro ao renovar a sessão: %v", err)
                }
                _, _, err = client.From("api_usage").Update(map[string]interface{}{
                    "current_api_calls": 0,
                }, "", "").Eq("id", id).Execute()
            }
            if err != nil {
                return fmt.Errorf("erro ao atualizar a tabela api_usage: %v", err)
            }
		}
	}

	return nil
}

func GetSubscriptionStatus(client *supa.Client, userId string) (string, error) {
	data, _, err := client.From("subscriptions").Select("status", "exact", false).Eq("user_id", userId).Execute()
	if err != nil {
		return "", fmt.Errorf("erro ao consultar a tabela subscriptions: %v", err)
	}

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar a resposta JSON: %v", err)
	}

	if len(result) > 0 {
		if status, ok := result[0]["status"].(string); ok {
			return status, nil
		}
		return "", fmt.Errorf("o campo 'status' não é uma string")
	}

	return "", nil
}

func refreshSession(client *supa.Client) error {
    // Implemente a lógica para renovar o token aqui
    err := client.Auth.Reauthenticate() // Este método pode variar
    if err != nil {
        return fmt.Errorf("erro ao tentar renovar a sessão: %v", err)
    }

    return nil
}
