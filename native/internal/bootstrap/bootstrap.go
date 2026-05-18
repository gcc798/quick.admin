package bootstrap

import (
	"github.com/gcc798/nai-tizi/internal/container"
	"github.com/gcc798/nai-tizi/internal/jobs"
	"github.com/gcc798/nai-tizi/internal/logger"
	"github.com/gcc798/nai-tizi/internal/messaging/websocket"
)

// Bootstrap wires optional scaffold components without binding concrete business logic.
type Bootstrap struct{}

// New creates the bootstrap coordinator for optional runtime integrations.
func New(c container.Container) (*Bootstrap, error) {
	b := &Bootstrap{}
	if sched := c.GetScheduler(); sched != nil {
		if err := jobs.RegisterJobs(sched, c.GetDB(), c.GetRedis(), nil, c.GetLogger()); err != nil {
			return nil, err
		}
	}
	return b, nil
}

// ConfigureWebSocketHandler applies websocket runtime options.
func (b *Bootstrap) ConfigureWebSocketHandler(handler *websocket.Handler, cfg any, log logger.Logger) {
	// The scaffold keeps this hook for business projects to extend websocket behavior.
}
