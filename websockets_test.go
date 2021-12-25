package ledgerx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type TestHandler struct {
	t *testing.T
}

func NewTestHandler(t *testing.T) *TestHandler {
	return &TestHandler{
		t: t,
	}
}

func (s *TestHandler) BookTopMessage(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	counter := 0
	for {
		if counter == 0 {
			bookTopResponse := &TopBookResponse{
				Type:       "book_top",
				ContractID: 123,
				Ask:        123,
				Bid:        123,
				Clock:      123,
			}

			bytes, err := json.Marshal(bookTopResponse)
			if err != nil {
				s.t.Fatalf("Error marshaling bookTopMessage")
			}
			err = c.WriteMessage(1, bytes)
			if err != nil {
				s.t.Fatalf("Error writing message")
			}
		}
	}
}

func (s *TestHandler) ActionReportMessage(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	counter := 0
	for {
		if counter == 0 {
			actionReportResponse := &ActionReportResponse{
				Type:                "action_report",
				ContractID:          123,
				Ask:                 123,
				Bid:                 123,
				Clock:               123,
				CustomerID:          123,
				MarketParticipantID: 123,
				Price:               123,
				Size:                123,
			}

			bytes, err := json.Marshal(actionReportResponse)
			if err != nil {
				s.t.Fatalf("Error marshaling bookTopMessage")
			}
			err = c.WriteMessage(1, bytes)
			if err != nil {
				s.t.Fatalf("Error writing message")
			}
		}
	}
}

func TestBookTopHandling(t *testing.T) {
	testHandler := NewTestHandler(t)
	s := httptest.NewServer(http.HandlerFunc(testHandler.BookTopMessage))
	defer s.Close()
	websocketUrl := "ws" + strings.TrimPrefix(s.URL, "http")

	ledgerClient := NewLedgerX(websocketUrl, "", "", "")

	if err := ledgerClient.Connect(); err != nil {
		t.Fatalf("Error connecting to web socket: %s", err.Error())
	}
	for {
		select {
		case message := <-ledgerClient.Listen():
			msg := message.Data.(TopBookResponse)
			assert.Equal(t, ChanBookTop, msg.Type, "should be correct types")
			assert.Equal(t, int64(123), msg.Bid, "Bids should match")
			assert.Equal(t, int64(123), msg.Ask, "ask should match")
			return
		}
	}
}

func TestActionReportHandling(t *testing.T) {
	testHandler := NewTestHandler(t)
	s := httptest.NewServer(http.HandlerFunc(testHandler.ActionReportMessage))
	defer s.Close()
	websocketUrl := "ws" + strings.TrimPrefix(s.URL, "http")

	ledgerClient := NewLedgerX(websocketUrl, "", "", "")

	if err := ledgerClient.Connect(); err != nil {
		t.Fatalf("Error connecting to web socket: %s", err.Error())
	}
	for {
		select {
		case message := <-ledgerClient.Listen():
			msg := message.Data.(ActionReportResponse)
			assert.Equal(t, ChanActionReport, msg.Type, "should be correct types")
			assert.Equal(t, int64(123), msg.Bid, "Bids should match")
			assert.Equal(t, int64(123), msg.Ask, "ask should match")
			assert.Equal(t, int64(123), msg.CustomerID, "CustomerID should match")
			assert.Equal(t, int64(123), msg.MarketParticipantID, "MarketParticipantID should match")
			assert.Equal(t, int64(123), msg.Price, "Price should match")
			assert.Equal(t, int64(123), msg.Size, "Size should match")
			return
		}
	}
}
