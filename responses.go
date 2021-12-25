package ledgerx

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Response struct {
	Type string `json:"type"`
}

type Message struct {
	Type string
	Data interface{}
}

type TopBookResponse struct {
	Type       string `json:"type"`
	ContractID int64  `json:"contract_id"`
	Ask        int64  `json:"ask"`
	AskSize    int64  `json:"ask_size"`
	Bid        int64  `json:"bid"`
	BidSize    int64  `json:"bid_size"`
	Clock      int64  `json:"clock"`
}

type ActionReportResponse struct {
	Type                string `json:"type"`
	ContractID          int64  `json:"contract_id"`
	Ask                 int64  `json:"ask"`
	Bid                 int64  `json:"bid"`
	Clock               int64  `json:"clock"`
	CustomerID          int64  `json:"cid"`
	MarketParticipantID int64  `json:"mpid"`
	InsertedTime        int64  `json:"inserted_time"`
	UpdatedTime         int64  `json:"updated_time"`
	Timestamp           int64  `json:"timestamp"`
	Price               int64  `json:"price"`
	OriginalPrice       int64  `json:"original_price"`
	InsertedPrice       int64  `json:"inserted_price"`
	FilledPrice         int64  `json:"filled_price"`
	IsAsk               bool   `json:"is_ask"`
	IsVolatile          bool   `json:"is_volatile"`
	Size                int64  `json:"size"`
	OriginalSize        int64  `json:"original_size"`
	InsertedSize        int64  `json:"inserted_size"`
	FilledSize          int64  `json:"filled_size"`
	OrderType           string `json:"order_type"`
	MessageID           string `json:"mid"`
	StatusType          int    `json:"status_type"`
	StatusReason        int    `json:"status_reason"`
}

func (a *ActionReportResponse) String() string {
	json, _ := json.Marshal(a)
	return string(json)
}

type LedgerTime struct {
	time.Time
}

const ledgerTimeLayout = "2006-01-02 15:04:05+0000"

func (lt *LedgerTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		lt.Time = time.Time{}
		return
	}
	lt.Time, err = time.Parse(ledgerTimeLayout, s)
	return
}

func (lt *LedgerTime) MarshalJSON() ([]byte, error) {
	if lt.Time.UnixNano() == (time.Time{}).UnixNano() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", lt.Time.Format(ledgerTimeLayout))), nil
}

type ListContractsData struct {
	ID              int64      `json:"id"`
	Label           string     `json:"label"`
	Name            string     `json:"name"`
	IsCall          bool       `json:"is_call"`
	Active          bool       `json:"active"`
	StrikePrice     int32      `json:"strike_price"`
	MinIncrement    int32      `json:"min_increment"`
	DateLive        LedgerTime `json:"date_live"`
	DateExpires     LedgerTime `json:"date_expires"`
	DateExercise    LedgerTime `json:"date_exercise"`
	UnderlyingAsset string     `json:"underlying_asset"`
	CollateralAsset string     `json:"collateral_asset"`
	DerivativeType  string     `json:"derivative_type"`
	OpenInterest    int32      `json:"open_interest"`
	IsNextDay       bool       `json:"is_next_day"`
	Multiplier      int32      `json:"multiplier"`
	Type            string     `json:"type"`
}

type ListContractsResponse struct {
	Data     []ListContractsData `json:"data"`
	Metadata []Metadata          `json:"metadata"`
}

type ListOpenOrdersResponse struct {
	Data     []ListOpenOrdersData `json:"data"`
	Metadata []Metadata           `json:"metadata"`
}

type ListOpenOrdersData struct {
	Mid           string `json:"mid"`
	Type          string `json:"type"`
	Mpid          int64  `json:"mpid"`
	Cid           int64  `json:"cid"`
	Timestamp     int64  `json:"timestamp"`
	Ticks         int64  `json:"ticks"`
	ContractID    int64  `json:"contract_id"`
	OriginalPrice int64  `json:"orignal_price"`
	OriginalSize  int64  `json:"orignal_size"`
	InsertedPrice int64  `json:"inserted_price"`
	InsertedSize  int64  `json:"inserted_size"`
	FilledPrice   int64  `json:"filled_price"`
	FilledSize    int64  `json:"filled_size"`
	Vwap          int32  `json:"vwap"`
	StatusType    int32  `json:"status_type"`
	StatusReason  int32  `json:"status_reason"`
	IsAsk         bool   `json:"is_ask"`
	InsertedTime  int64  `json:"inserted_time"`
	UpdatedTime   int64  `json:"updated_time"`
	OrderType     string `json:"order_type"`
	Clock         int64  `json:"clock"`
}

type Metadata struct {
	TotalCount int64  `json:"total_count"`
	Next       string `json:"next"`
	Previous   string `json:"previous"`
	Limit      int64  `json:"limit"`
	Offset     int64  `json:"offset"`
}

func (m *Metadata) String() string {
	json, _ := json.Marshal(m)
	return string(json)
}

type ListTradesResponse struct {
	Data     []ListTradeData `json:"data"`
	Metadata Metadata        `json:"meta"`
}

type ListTradeData struct {
	ID            int64  `json:"id"`
	ContractID    int64  `json:"contract_id,string"`
	ContractLabel string `json:"contract_label"`
	FilledPrice   int64  `json:"filled_price"`
	FilledSize    int64  `json:"filled_size"`
	Fee           int64  `json:"fee"`
	OrderType     string `json:"order_type"`
	OrderID       string `json:"order_id"`
	StatusType    string `json:"status_type"`
	Created       string `json:"created"`
	Timestamp     string `json:"timestamp"`
}

type CreateOrderRequest struct {
	OrderType   string `json:"order_type"`
	ContractID  int32  `json:"contract_id"`
	IsAsk       bool   `json:"is_ask"`
	SwapPurpose string `json:"swap_purpose"`
	Size        int32  `json:"size"`
	Price       int32  `json:"price"`
	Volatile    bool   `json:"volatile"`
}

func (c *CreateOrderRequest) String() string {
	json, _ := json.Marshal(c)
	return string(json)
}

type CreateOrderData struct {
	Mid string `json:"mid"`
}

type CreateOrderResponse struct {
	Data CreateOrderData `json:"data"`
}

type CancelOrderRequest struct {
	ContractID int32 `json:"contract_id"`
}

type CancelAndReplaceRequest struct {
	ContractID int32 `json:"contract_id"`
	Size       int32 `json:"size"`
	Price      int32 `json:"price"`
}

type InvalidTokenErrorResponse struct {
	Error string `json:"error"`
}

type TradeErrorResponse struct {
	Error TradeErrorObject `json:"error"`
}

type TradeErrorObject struct {
	Message string `json:"message"`
	Code    int32  `json:"code"`
}

type OpenPositionsMessage struct {
	Type      string     `json:"type"`
	Positions []Position `json:"positions"`
}

type HeartbeatMessage struct {
	Type       string `json:"type"`
	Timestamp  int64  `json:"timestamp"`
	Ticks      int64  `json:"ticks"`
	RunID      int64  `json:"run_id"`
	IntervalMS int64  `json:"interval_ms"`
}

func (o *OpenPositionsMessage) String() string {
	json, _ := json.Marshal(o)
	return string(json)
}

type Position struct {
	ContractID           int64 `json:"contract_id"`
	ExerciseSize         int64 `json:"exercise_size"`
	MarketPariticipantID int64 `json:"mpid"`
	Size                 int64 `json:"size"`
}

type BalanceUpdateMessage struct {
	Collateral Collateral `json:"collateral"`
}

func (b *BalanceUpdateMessage) String() string {
	json, _ := json.Marshal(b)
	return string(json)
}

type Collateral struct {
	AvailableBalances         LedgerBalance `json:"available_balances"`
	DeliverableLockedBalances LedgerBalance `json:"deliverable_locked_balances"`
	FeeLockedBalances         LedgerBalance `json:"fee_locked_balances"`
	OrderLockedBalances       LedgerBalance `json:"order_locked_balances"`
	PositionLockedBalances    LedgerBalance `json:"position_locked_balances"`
}

type LedgerBalance struct {
	BTC  int64 `json:"BTC"`
	CBTC int64 `json:"CBTC"`
	USD  int64 `json:"USD"`
	ETH  int64 `json:"ETH"`
}
