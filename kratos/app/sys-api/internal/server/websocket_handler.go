package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/force-c/nai-tizi/kratos/app/sys-api/internal/data"
	kerrors "github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/websocket"
)

const operationWebSocket = "/custom.ws/Connect"

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func registerWebSocketEndpoint(srv *khttp.Server, deps *GatewayDeps) {
	if srv == nil || deps == nil || deps.WebSocket == nil {
		return
	}
	srv.Route("/").GET("/ws", wrapOperation(operationWebSocket, func(ctx context.Context, httpCtx khttp.Context) error {
		return serveWebSocket(ctx, httpCtx, deps.WebSocket)
	}))
}

func serveWebSocket(ctx context.Context, httpCtx khttp.Context, hub *WebSocketHub) error {
	req := httpCtx.Request()
	resp := httpCtx.Response()
	userID := resolveWebSocketUserID(ctx, req)
	if userID <= 0 {
		return kerrors.Unauthorized("UNAUTHORIZED", "未登录")
	}
	conn, err := wsUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return err
	}
	client := hub.Register(userID, conn)
	defer conn.Close()
	defer hub.Unregister(userID, conn)

	if client == nil {
		return nil
	}
	_ = client.WriteJSON(map[string]any{
		"type": "connected",
		"data": map[string]any{"userId": userID, "connections": hub.ConnectionCount(userID), "totalConnections": hub.TotalConnections()},
	})

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			return nil
		}
		text := strings.TrimSpace(string(payload))
		if strings.EqualFold(text, "ping") {
			if err = client.WriteJSON(map[string]any{"type": "pong", "data": map[string]any{"userId": userID}}); err != nil {
				return nil
			}
			continue
		}
		if strings.EqualFold(text, "connections") {
			if err = client.WriteJSON(map[string]any{"type": "connections", "data": map[string]any{"userId": userID, "connections": hub.ConnectionCount(userID), "totalConnections": hub.TotalConnections()}}); err != nil {
				return nil
			}
			continue
		}
		if messageType == websocket.TextMessage {
			_ = hub.SendToUser(userID, "echo", map[string]any{"message": text})
			continue
		}
		if err = client.WriteMessage(messageType, payload); err != nil {
			return nil
		}
	}
}

func resolveWebSocketUserID(ctx context.Context, req *http.Request) int64 {
	if userID := data.CurrentUserID(ctx); userID > 0 {
		return userID
	}
	if req == nil {
		return 0
	}
	value := strings.TrimSpace(req.URL.Query().Get("userId"))
	if value == "" {
		return 0
	}
	userID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || userID <= 0 {
		return 0
	}
	return userID
}
