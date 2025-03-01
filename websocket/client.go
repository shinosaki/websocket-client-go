package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type WebSocketClient struct {
	URL       *url.URL
	conn      *websocket.Conn
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	onOpen    func(ws *WebSocketClient, isReconnecting bool)
	onClose   func(ws *WebSocketClient, isReconnecting bool)
	onMessage func(ws *WebSocketClient, payload []byte)
}

func NewWebSocketClient(
	onOpen func(ws *WebSocketClient, isReconnecting bool),
	onClose func(ws *WebSocketClient, isReconnecting bool),
	onMessage func(ws *WebSocketClient, payload []byte),
) *WebSocketClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketClient{
		ctx:       ctx,
		cancel:    cancel,
		onOpen:    onOpen,
		onClose:   onClose,
		onMessage: onMessage,
	}
}

func (ws *WebSocketClient) Connect(
	webSocketUrl string,
	attempts int,
	interval time.Duration,
	isReconnecting bool,
) (err error) {
	ws.URL, err = url.Parse(webSocketUrl)
	if err != nil {
		return err
	}

	// attempt retry
	for range attempts {
		origin := ws.URL.Scheme + "://" + ws.URL.Host
		ws.conn, err = websocket.Dial(ws.URL.String(), "", origin)
		if err == nil {
			break
		}
		time.Sleep(interval * time.Second)
	}
	if err != nil {
		return fmt.Errorf("websocket dial falied: %v", err)
	}

	if ws.onOpen != nil {
		ws.onOpen(ws, isReconnecting)
	}

	// Message Handler
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		for {
			select {
			case <-ws.ctx.Done():
				// log.Println("websocket receive cancel")
				return
			default:
				var payload []byte
				if err := websocket.Message.Receive(ws.conn, &payload); err != nil {
					// log.Println("receive error", err)
					return
				}
				ws.onMessage(ws, payload)
			}
		}
	}()

	return nil
}

func (ws *WebSocketClient) Disconnect(isReconnecting bool) {
	if ws.conn != nil {
		ws.conn.Close()
	}

	ws.cancel()
	ws.wg.Wait()

	if ws.onClose != nil {
		ws.onClose(ws, isReconnecting)
	}
}

func (ws *WebSocketClient) Reconnect(url string, attempts int, interval time.Duration) error {
	ws.Disconnect(true)
	ws.ctx, ws.cancel = context.WithCancel(context.Background())
	return ws.Connect(url, attempts, interval, true)
}

func (ws *WebSocketClient) SendJSON(v any) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %v", err)
	}
	return websocket.Message.Send(ws.conn, string(bytes))
}
