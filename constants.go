package ledgerx

const (
	ProdWebSocketBaseURL    = "wss://api.ledgerx.com/ws"
	ProdRestBaseURL         = "https://api.ledgerx.com"
	ProdTradingBaseURL      = "https://trade.ledgerx.com"
	StagingWebSocketBaseURL = "wss://api-staging.ledgerx.com/ws"
	StagingRestBaseURL      = "https://api-staging.ledgerx.com"
	StagingTradingBaseURL   = "https://staging.ledgerx.com"
)

// Available channels
const (
	ChanBookTop             = "book_top"
	ChanActionReport        = "action_report"
	ChanBalanceUpdate       = "collateral_balance_update"
	ChanOpenPositionsUpdate = "open_positions_update"
	ChanHeartbeat           = "heartbeat"
	ChanMeta                = "meta"
	ChanAuthSuccess         = "auth_success"
	ChanAuthFailure         = "unauth_success"
	ChanStateManifest       = "state_manifest"
)

// Contract IDs
const (
	BtcUsdPair = 22220309
)

var ContractIDToPairs = map[int]string{
	22220309: "XBT/USD",
}

const (
	ListContractLookback = 4 // days
	DefaultPageSize      = 100
)

const (
	StatusCodeOrderInserted             = 200
	StatusCodeTradeOccured              = 201
	StatusCodeMarketOrderNotFilled      = 202
	StatusCodeOrderCancelled            = 203
	StatusCodeOrderCancelledAndReplaced = 204
	StatusCodeContractNotFound          = 600
	StatusCodeOrderNotFound             = 601
	StatusCodeInvalidOrder              = 602
	StatusCodeOrderRejected             = 607
	StatusCodeNoFunds                   = 608
	StatusCodeContractExpired           = 610
)

const (
	ReasonCodeFullFill            = 52
	ReasonCodeCancelledByExchange = 53
)
