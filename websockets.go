package ledgerx

import (
	"encoding/json"

	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Connect to the Kraken API, this should only be called once.
func (l *LedgerX) Connect() error {
	l.wg.Add(1)
	go l.managerThread()

	if err := l.dial(); err != nil {
		return err
	}

	l.wg.Add(1)
	go l.listenSocket()

	return nil
}

func (l *LedgerX) Listen() <-chan Message {
	return l.msg
}

func (l *LedgerX) Close() error {
	for i := 0; i < 2; i++ {
		l.stop <- struct{}{}
	}
	l.wg.Wait()

	if l.conn != nil {
		if err := l.conn.Close(); err != nil {
			return err
		}
	}

	close(l.stop)
	close(l.msg)
	close(l.connect)
	return nil
}

func (l *LedgerX) getWebsocketUrl() string {
	if l.token != "" {
		return fmt.Sprintf("%s?token=%s", l.websocketUrl, l.token)
	}
	return l.websocketUrl
}

func (l *LedgerX) dial() error {
	c, resp, err := websocket.DefaultDialer.Dial(l.getWebsocketUrl(), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	l.conn = c
	return nil
}

func (l *LedgerX) managerThread() {
	defer l.wg.Done()

	heartbeat := time.NewTicker(l.heartbeatTimeout)
	defer heartbeat.Stop()

	for {
		select {
		case <-l.stop:
			return
		case <-l.connect:
			time.Sleep(l.reconnectTimeout)

			log.Warnf("reconnecting ledgerx...")

			if err := l.dial(); err != nil {
				log.Error(err)
				l.connect <- struct{}{}
				continue
			}

			l.wg.Add(1)
			go l.listenSocket()
		case <-heartbeat.C:
			err := l.conn.WriteMessage(websocket.TextMessage, []byte("pong"))
			if err != nil {
				log.Println(err)
				l.connect <- struct{}{}
			}
		}
	}
}

func (l *LedgerX) listenSocket() {
	defer l.wg.Done()

	if l.conn == nil {
		return
	}

	if err := l.conn.SetReadDeadline(time.Now().Add(l.readTimeout)); err != nil {
		log.Error(err)
		return
	}

	for {
		select {
		case <-l.stop:
			return
		default:
			_, msg, err := l.conn.ReadMessage()
			if err != nil {
				log.Error(err)
				l.connect <- struct{}{}
				return
			}

			if err := l.conn.SetReadDeadline(time.Now().Add(l.readTimeout)); err != nil {
				log.Error(err)
				l.connect <- struct{}{}
				return
			}

			log.Tracef("server->client: %s", string(msg))

			if err := l.handleMessage(msg); err != nil {
				log.Error(err)
			}
		}
	}
}

func (l *LedgerX) handleMessage(data []byte) error {
	if len(data) == 0 {
		return errors.Errorf("Empty response: %s", string(data))
	}

	var res Response
	err := json.Unmarshal(data, &res)
	if err != nil {
		return errors.Errorf("Error during unmarshal response: %s", string(data))
	}

	switch res.Type {
	case ChanBookTop:
		var jsonRes TopBookResponse
		err := json.Unmarshal(data, &jsonRes)
		if err != nil {
			return errors.Errorf("Error during unmarshal TopBookResponse: %s", string(data))
		}
		l.msg <- Message{
			Type: ChanBookTop,
			Data: jsonRes,
		}
	case ChanActionReport:
		var jsonRes ActionReportResponse
		err := json.Unmarshal(data, &jsonRes)
		if err != nil {
			return errors.Errorf("Error during unmarshal ActionReportResponse: %s", string(data))
		}
		//log.Println(string(data))
		l.msg <- Message{
			Type: ChanActionReport,
			Data: jsonRes,
		}
	case ChanBalanceUpdate:
		var jsonRes BalanceUpdateMessage
		//log.Println(string(data))
		err := json.Unmarshal(data, &jsonRes)
		if err != nil {
			return errors.Errorf("Error during unmarshal BalanceUpdateMessage: %s", string(data))
		}
		l.msg <- Message{
			Type: ChanBalanceUpdate,
			Data: jsonRes,
		}
	case ChanOpenPositionsUpdate:
		var jsonRes OpenPositionsMessage
		err := json.Unmarshal(data, &jsonRes)
		if err != nil {
			return errors.Errorf("Error during unmarshal OpenPositionsMessage: %s", string(data))
		}
		l.msg <- Message{
			Type: ChanOpenPositionsUpdate,
			Data: jsonRes,
		}
	case ChanHeartbeat:
		var jsonRes HeartbeatMessage
		err := json.Unmarshal(data, &jsonRes)
		if err != nil {
			return errors.Errorf("Error during unmarshal HeartbeatMessage: %s", string(data))
		}
		l.msg <- Message{
			Type: ChanHeartbeat,
			Data: jsonRes,
		}
	default:
		if handleInfoMessage(res.Type, data) == false {
			return errors.Errorf("Unexpected message: %s", string(data))
		}
	}
	return nil
}

func handleInfoMessage(messageType string, data []byte) bool {
	switch messageType {
	case ChanAuthSuccess:
		//log.Warn("Ledger authentication Successful")
		return true
	case ChanAuthFailure:
		//log.Warn("Ledger authentication Failure")
		return true
	case ChanMeta:
		//log.Warn("Ledger received session ID.")
		return true
	case ChanStateManifest:
		//log.Warn("Ledger received state manifest.")
		return true
	default:
		return false
	}
}
