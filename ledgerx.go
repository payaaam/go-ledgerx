package ledgerx

import (
	"encoding/json"

	"bytes"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	//log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type clientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type LedgerX struct {
	websocketUrl string
	restUrl      string
	tradingUrl   string
	token        string
	conn         *websocket.Conn
	client       clientInterface

	reconnectTimeout time.Duration
	readTimeout      time.Duration
	heartbeatTimeout time.Duration

	msg     chan Message
	connect chan struct{}
	stop    chan struct{}

	wg sync.WaitGroup
}

func NewLedgerX(websocketUrl string, restUrl string, tradingUrl string, apiKey string) *LedgerX {
	return &LedgerX{
		websocketUrl:     websocketUrl,
		restUrl:          restUrl,
		tradingUrl:       tradingUrl,
		token:            apiKey,
		reconnectTimeout: 5 * time.Second,
		readTimeout:      15 * time.Second,
		heartbeatTimeout: 6 * time.Second,
		connect:          make(chan struct{}, 1),
		msg:              make(chan Message, 1024),
		stop:             make(chan struct{}, 1),
		client:           http.DefaultClient,
	}
}

func (l *LedgerX) ListContracts() (*ListContractsResponse, error) {
	currentTime := time.Now()
	beforeTimestamp := fmt.Sprintf("%sT00:00", currentTime.AddDate(0, 0, 2).Format("2006-01-02"))
	afterTimestamp, _ := time.Parse("2006-01-02T15:04", beforeTimestamp)
	afterTimestamp = afterTimestamp.AddDate(0, 0, -1*ListContractLookback)

	url := fmt.Sprintf("%s/trading/contracts?derivative_type=day_ahead_swap&before_ts=%s&after_ts=%s", l.restUrl, beforeTimestamp, afterTimestamp.Format("2006-01-02T15:04"))
	req, err := l.makeRequest("GET", url, true, nil)
	if err != nil {
		return nil, fmt.Errorf("Error during request creation: %s", err.Error())
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error during request execution: %s", err.Error())
	}
	defer resp.Body.Close()

	listContractsResponse := &ListContractsResponse{}
	parseErr := l.parseResponse(resp, listContractsResponse)
	if parseErr != nil {
		return nil, parseErr
	}

	return listContractsResponse, nil
}

func (l *LedgerX) ListOpenOrders() (*ListOpenOrdersResponse, error) {
	url := fmt.Sprintf("%s/api/open-orders", l.tradingUrl)
	req, err := l.makeRequest("GET", url, true, nil)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error during request creation: %s", err.Error())
	}
	defer resp.Body.Close()

	listOpenOrdersResponse := &ListOpenOrdersResponse{}
	parseErr := l.parseResponse(resp, listOpenOrdersResponse)
	if parseErr != nil {
		return nil, parseErr
	}

	return listOpenOrdersResponse, nil
}

func (l *LedgerX) ListTrades(derivativeType string, lookbackDays int, asset string, offset int32) (*ListTradesResponse, error) {
	if derivativeType == "" {
		derivativeType = "day_ahead_swap"
	}

	if lookbackDays == 0 {
		lookbackDays = -2
	}

	afterTimestamp := fmt.Sprintf("%sT00:00", time.Now().AddDate(0, 0, lookbackDays).Format("2006-01-02"))
	url := fmt.Sprintf("%s/trading/trades?derivative_type=%s&after_ts=%s&limit=%d&offset=%d", l.restUrl, derivativeType, afterTimestamp, DefaultPageSize, offset)
	if asset != "" {
		url += fmt.Sprintf("&asset=%s", asset)
	}

	req, err := l.makeRequest("GET", url, true, nil)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error during request creation: %s", err.Error())
	}
	defer resp.Body.Close()

	listTradesResponse := &ListTradesResponse{}
	parseErr := l.parseResponse(resp, listTradesResponse)
	if parseErr != nil {
		return nil, parseErr
	}

	return listTradesResponse, nil

}

func (l *LedgerX) ListPositions(offset int32) (*ListTradesResponse, error) {
	url := fmt.Sprintf("%s/trading/positions?liimt=%v&offset=%v", l.restUrl, DefaultPageSize, offset)

	req, err := l.makeRequest("GET", url, true, nil)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error during request creation: %s", err.Error())
	}
	defer resp.Body.Close()
	// buf := new(bytes.Buffer)
	// buf.ReadFrom(resp.Body)
	// newStr := buf.String()

	// fmt.Println(newStr)

	listTradesResponse := &ListTradesResponse{}
	// parseErr := l.parseResponse(resp, listTradesResponse)
	// if parseErr != nil {
	// 	return nil, parseErr
	// }

	return listTradesResponse, nil

}

func (l *LedgerX) CreateOrder(request *CreateOrderRequest) (*CreateOrderResponse, error) {
	requestUrl := fmt.Sprintf("%s/api/orders", l.tradingUrl)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json, %s", err.Error())
	}

	//log.Println(requestBody)
	req, err := l.makeRequest("POST", requestUrl, true, bytes.NewReader(requestBody))

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error during request creation: %s", err.Error())
	}
	defer resp.Body.Close()

	createOrderResponse := &CreateOrderResponse{}
	parseErr := l.parseResponse(resp, createOrderResponse)
	if parseErr != nil {
		return nil, parseErr
	}

	return createOrderResponse, nil

}

func (l *LedgerX) CancelOrder(mid string, contractID int32) error {
	requestUrl := fmt.Sprintf("%s/api/orders/%s", l.tradingUrl, mid)

	requestBody, err := json.Marshal(&CancelOrderRequest{
		ContractID: contractID,
	})
	if err != nil {
		return fmt.Errorf("error marshaling json, %s", err.Error())
	}

	req, err := l.makeRequest("DELETE", requestUrl, true, bytes.NewReader(requestBody))

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error during request creation: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return l.parseResponse(resp, nil)
	}
	return nil
}

func (l *LedgerX) CancelAndReplaceOrder(mid string, request *CancelAndReplaceRequest) error {
	requestUrl := fmt.Sprintf("%s/api/orders/%s/edit", l.tradingUrl, mid)

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling json, %s", err.Error())
	}
	req, err := l.makeRequest("POST", requestUrl, true, bytes.NewReader(requestBody))

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error during request creation: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return l.parseResponse(resp, nil)
	}
	return nil
}

func (l *LedgerX) makeRequest(method string, requestUrl string, requiresAuth bool, data io.Reader) (*http.Request, error) {
	//log.Println(strings.NewReader(data.Encode()))
	// log.Println(requestUrl)
	// log.Println(l.token)

	req, err := http.NewRequest(method, requestUrl, data)
	if err != nil {
		return nil, fmt.Errorf("Error during request creation: %s", err.Error())
	}
	if requiresAuth == true {
		req.Header.Add("Authorization", fmt.Sprintf("JWT %s", l.token))
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func (l *LedgerX) parseResponse(response *http.Response, resType interface{}) error {
	if response.StatusCode != 200 {
		invalidTokenErrorResponse := InvalidTokenErrorResponse{}
		tradeErrorResponse := TradeErrorResponse{}

		invalidTokenErr := parseJson(response, &invalidTokenErrorResponse)
		tradeErr := parseJson(response, &tradeErrorResponse)

		var message string
		if invalidTokenErr == nil && tradeErr != nil {
			// Invalid token parsing was not nil, error was invalid token
			message = invalidTokenErrorResponse.Error
		} else if invalidTokenErr != nil && tradeErr == nil {
			// Invalid token parsing was not nil, error was invalid token
			message = tradeErrorResponse.Error.Message
		}

		body, _ := ioutil.ReadAll(response.Body)
		logrus.Errorf("LedgerX Error: %s", string(body))

		switch message {
		case "INVALID_TOKEN":
			return fmt.Errorf("ledgerx api error: invalid token error")
		case "":
			return fmt.Errorf("ledgerx api error (check for correct env token): %s", "unknown error")
		default:
			return fmt.Errorf("ledgerx api error (check for correct env token): %s", message)
		}
	}

	return parseJson(response, resType)
}

func parseJson(response *http.Response, resType interface{}) error {
	if response.Body == nil {
		return fmt.Errorf("Error during response parsing: can not read response body")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error during response parsing: can not read response body (%s)", err.Error())
	}

	response.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	//log.Println(string(body))
	if err = json.Unmarshal(body, &resType); err != nil {
		return fmt.Errorf("Error during response parsing: json marshalling (%s)", err.Error())
	}
	return nil
}
