package ledgerx

import (
	"fmt"
	"os"
	"testing"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrderFlow(t *testing.T) {
	return
	// if testing.Short() {
	// 	t.Skip("Skipping Create Order Flow Test")
	// }
	ledgerWebClient := NewLedgerX("", StagingRestBaseURL, StagingTradingBaseURL, getApiKey())

	assert.Equal(t, fetchOpenOrders(t, ledgerWebClient), 0, "should not any open orders")

	contractID, err := getBtcContractId(ledgerWebClient)
	assert.Nil(t, err, "should not error when fetching contracts")
	assert.NotEqual(t, contractID, 0, "should find contract ID")

	quantity := decimal.NewFromInt(1)
	price := decimal.NewFromInt(1)

	createOrderRequest := &CreateOrderRequest{
		OrderType:   "limit",
		ContractID:  int32(contractID),
		IsAsk:       false,
		SwapPurpose: "undisclosed",
		Size:        int32(quantity.BigInt().Int64()),
		Price:       int32(price.Mul(decimal.NewFromInt(100)).BigInt().Int64()),
		Volatile:    false,
	}

	createOrderResponse, err := ledgerWebClient.CreateOrder(createOrderRequest)
	assert.Nil(t, err, "should not error when creating new order")
	orderID := createOrderResponse.Data.Mid

	assert.Equal(t, fetchOpenOrders(t, ledgerWebClient), 1, "should not any open orders")

	cancelErr := ledgerWebClient.CancelOrder(orderID, int32(contractID))
	assert.Nil(t, cancelErr, "should not error when cancelling order")

	assert.Equal(t, fetchOpenOrders(t, ledgerWebClient), 0, "should not any open orders")
}

func getApiKey() string {
	apiKey := os.Getenv("LEDGER_API_KEY")
	if apiKey == "" {
		log.Fatalf("LEDGER_API_KEY required to use Ledger Exchange.")
	}
	return apiKey
}

func getBtcContractId(client *LedgerX) (int64, error) {
	listContractsResponse, err := client.ListContracts()
	if err != nil {
		return 0, fmt.Errorf("Error fetching contracts: %s", err.Error())
	}

	log.Println(listContractsResponse)

	for _, contract := range listContractsResponse.Data {
		if contract.UnderlyingAsset == "CBTC" {
			log.Println("FOUND CONTRACT")
			return contract.ID, nil
		}
	}

	return 0, fmt.Errorf("contract not found")
}

func fetchOpenOrders(t *testing.T, client *LedgerX) int {
	openOrdersResponse, err := client.ListOpenOrders()
	assert.Nil(t, err, "should not return error on listOpenOrders")
	return len(openOrdersResponse.Data)
}
