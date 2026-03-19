package biz

import "context"

// HealthUsecase exposes the minimal rpc business contract for bootstrapping the Kratos project.
type HealthUsecase struct{}

func NewHealthUsecase() *HealthUsecase {
	return &HealthUsecase{}
}

func (uc *HealthUsecase) Ping(_ context.Context, name string) string {
	if name == "" {
		return "pong from sys-rpc"
	}
	return "pong from sys-rpc: " + name
}
