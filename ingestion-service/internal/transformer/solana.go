package transformer

import (
	"fmt"
	"time"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/kafka"
)

func TransformSolanaTransactions(data interface{}) ([]kafka.Transaction, error) {
	txsRaw, ok := data.([]interface{})
	if !ok {
		return nil, nil
	}

	var transactions []kafka.Transaction
	for _, txRaw := range txsRaw {
		tx, ok := txRaw.(map[string]interface{})
		if !ok {
			continue
		}

		hash, _ := tx["signature"].(string)
		if hash == "" {
			hash, _ = tx["hash"].(string)
		}

		var timestamp time.Time
		if blockTime, ok := tx["blockTime"].(float64); ok {
			timestamp = time.Unix(int64(blockTime), 0).UTC()
		}
		if timestamp.IsZero() {
			timestamp = time.Now().UTC()
		}

		var from, to, amount, symbol string
		
		if meta, ok := tx["meta"].(map[string]interface{}); ok {
			if preBalances, ok := meta["preBalances"].([]interface{}); ok && len(preBalances) > 0 {
				if postBalances, ok := meta["postBalances"].([]interface{}); ok && len(postBalances) > 0 {
					if len(preBalances) == len(postBalances) {
						pre, _ := preBalances[0].(float64)
						post, _ := postBalances[0].(float64)
						amount = fmt.Sprintf("%.0f", post-pre)
					}
				}
			}
		}

		if transaction, ok := tx["transaction"].(map[string]interface{}); ok {
			if message, ok := transaction["message"].(map[string]interface{}); ok {
				if accountKeys, ok := message["accountKeys"].([]interface{}); ok {
					if len(accountKeys) > 0 {
						from, _ = accountKeys[0].(string)
					}
					if len(accountKeys) > 1 {
						to, _ = accountKeys[1].(string)
					}
				}
			}
		}

		symbol = "SOL"

		transactions = append(transactions, kafka.Transaction{
			Hash:      hash,
			From:      from,
			To:        to,
			Amount:    amount,
			Symbol:    symbol,
			Timestamp: timestamp,
			Type:      "transfer",
		})
	}

	return transactions, nil
}
