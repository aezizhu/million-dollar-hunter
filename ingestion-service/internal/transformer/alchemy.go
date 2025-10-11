package transformer

import (
	"time"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/kafka"
)

type AlchemyTransfer struct {
	UniqueId  string  `json:"uniqueId"`
	Hash      string  `json:"hash"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Value     float64 `json:"value"`
	Asset     string  `json:"asset"`
	Category  string  `json:"category"`
	BlockNum  string  `json:"blockNum"`
	Metadata  Metadata `json:"metadata"`
	RawContract RawContract `json:"rawContract"`
}

type Metadata struct {
	BlockTimestamp string `json:"blockTimestamp"`
}

type RawContract struct {
	Value   string `json:"value"`
	Address string `json:"address"`
	Decimal int    `json:"decimal"`
}

type AlchemyResponse struct {
	Transfers []AlchemyTransfer `json:"transfers"`
}

func TransformAlchemyTransfers(data interface{}) ([]kafka.Transaction, error) {
	response, ok := data.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	transfersRaw, ok := response["transfers"].([]interface{})
	if !ok {
		return nil, nil
	}

	var transactions []kafka.Transaction
	for _, tr := range transfersRaw {
		transfer, ok := tr.(map[string]interface{})
		if !ok {
			continue
		}

		hash, _ := transfer["hash"].(string)
		if hash == "" {
			continue
		}
		
		from, _ := transfer["from"].(string)
		to, _ := transfer["to"].(string)
		category, _ := transfer["category"].(string)

		var amount string
		var symbol string
		var tokenAddress string

		if rawContract, ok := transfer["rawContract"].(map[string]interface{}); ok {
			if val, ok := rawContract["value"].(string); ok {
				amount = val
			}
			if addr, ok := rawContract["address"].(string); ok {
				tokenAddress = addr
			}
		}

		if asset, ok := transfer["asset"].(string); ok {
			symbol = asset
		}
		
		if tokenAddress == "" && category == "external" {
			tokenAddress = "0x0000000000000000000000000000000000000000"
		}

		var timestamp time.Time
		if metadata, ok := transfer["metadata"].(map[string]interface{}); ok {
			if blockTimestamp, ok := metadata["blockTimestamp"].(string); ok {
				parsed, err := time.Parse(time.RFC3339, blockTimestamp)
				if err == nil {
					timestamp = parsed
				}
			}
		}

		if timestamp.IsZero() {
			timestamp = time.Now().UTC()
		}

		txType := "transfer"
		if category != "" {
			txType = category
		}

		transactions = append(transactions, kafka.Transaction{
			Hash:         hash,
			From:         from,
			To:           to,
			Amount:       amount,
			Symbol:       symbol,
			TokenAddress: tokenAddress,
			Timestamp:    timestamp,
			Type:         txType,
		})
	}

	return transactions, nil
}
