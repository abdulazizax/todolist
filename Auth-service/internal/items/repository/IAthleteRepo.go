package repository

import (
	"context"

	pb "auth-service/genproto/auth"
)

type (
	IAuthRepo interface {
		Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error)
		Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error)
		Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.LogoutResponse, error)
		CreateAdmin(ctx context.Context, in *pb.CreateAdminRequest) (*pb.CreateAdminResponse, error)
		DeleteUser(ctx context.Context, in *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error)
		GetUserByEmail(ctx context.Context, in *pb.GetUserByEmailRequest) (*pb.RegisterResponse, error)
		UpdatePasswordService(ctx context.Context, req *pb.UpdateUserPasswordRequest) (*pb.RawResponse, error)
		SendVerificationCode(ctx context.Context, req *pb.SendVerificationCodeRequest) (*pb.RawResponse, error)
	}
)
