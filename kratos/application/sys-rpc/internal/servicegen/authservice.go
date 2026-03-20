package service

import (
	"context"

	pb "github.com/force-c/nai-tizi/kratos/api/system/v1"
)

type AuthServiceService struct {
	pb.UnimplementedAuthServiceServer
}

func NewAuthServiceService() *AuthServiceService {
	return &AuthServiceService{}
}

func (s *AuthServiceService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	return &pb.LoginReply{}, nil
}
func (s *AuthServiceService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.MessageReply, error) {
	return &pb.MessageReply{}, nil
}
func (s *AuthServiceService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenReply, error) {
	return &pb.RefreshTokenReply{}, nil
}
func (s *AuthServiceService) Me(ctx context.Context, req *pb.MeRequest) (*pb.MeReply, error) {
	return &pb.MeReply{}, nil
}
