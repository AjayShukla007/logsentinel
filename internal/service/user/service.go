package user

import (
	"context"
	"time"

	userrepo "github.com/AjayShukla007/logsentinel/internal/repository/user"
	pb "github.com/AjayShukla007/logsentinel/proto/gen/proto"
	"github.com/google/uuid"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	repo userrepo.Repository
}

func NewUserService(repo userrepo.Repository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	
	return &pb.User{
		Id:          uuid.New().String(),
		ClientId:    req.ClientId,
		AccountType: req.AccountType,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {

	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	if len(users) > 0 {
		return &pb.User{
			Id:          users[0].ID.String(),
			ClientId:    users[0].ClientID,
			AccountType: users[0].AccountType,
			CreatedAt:   users[0].CreatedAt.Format(time.RFC3339),
		}, nil
	}

	return &pb.User{
		Id:          req.UserId,
		ClientId:    "demo-client-id",
		AccountType: "pro",
		CreatedAt:   time.Now().AddDate(0, 0, -30).Format(time.RFC3339),
	}, nil
}

func (s *UserService) UpdateUserAccountType(ctx context.Context, req *pb.UpdateUserAccountTypeRequest) (*pb.User, error) {
	return &pb.User{
		Id:          req.UserId,
		ClientId:    "demo-client-id",
		AccountType: req.AccountType,
		CreatedAt:   time.Now().AddDate(0, 0, -30).Format(time.RFC3339),
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully (demo mode)",
	}, nil
}

func (s *UserService) ValidateUserCredentials(ctx context.Context, clientID, projectID, apiKey string) (bool, string, error) {
	// TODO: demo mode
	return true, uuid.New().String(), nil
}

func (s *UserService) GetUserProjectCount(ctx context.Context, userID string) (int, error) {
	return 3, nil
}

func (s *UserService) CheckUserQuota(ctx context.Context, req *pb.GetUserRequest) (*pb.QuotaResponse, error) {
	return &pb.QuotaResponse{
		AccountType:        "pro",
		ProjectCount:       int32(3),
		ProjectLimit:       5,
		LogCountLastMinute: int32(10),
		LogLimitPerMinute:  100,
		LastLogTime:        time.Now().Format(time.RFC3339),
	}, nil
}
