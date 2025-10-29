package project

import (
	"context"
	"time"

	"github.com/AjayShukla007/logsentinel/internal/repository/project"
	userrepo "github.com/AjayShukla007/logsentinel/internal/repository/user"
	pb "github.com/AjayShukla007/logsentinel/proto/gen/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectService struct {
	pb.UnimplementedProjectServiceServer
	repo     project.Repository
	userRepo userrepo.Repository
}

func NewProjectService(repo project.Repository, userRepo userrepo.Repository) *ProjectService {
	return &ProjectService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.Project, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	dummyID := uuid.New()
	return &pb.Project{
		Id:        dummyID.String(),
		Name:      req.Name,
		UserId:    req.UserId,
		ApiKey:    "demo-api-key-" + dummyID.String()[:8],
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *ProjectService) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.Project, error) {
	dummyID := uuid.New()
	userID := uuid.New()

	return &pb.Project{
		Id:        dummyID.String(),
		Name:      "Demo Project",
		UserId:    userID.String(),
		ApiKey:    "demo-api-key-12345678",
		CreatedAt: time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
	}, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	return &pb.DeleteProjectResponse{
		Success: true,
		Message: "Project deleted successfully (demo mode)",
	}, nil
}
