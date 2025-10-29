package team

// import (
// 	"context"
// 	"time"

// 	"github.com/AjayShukla007/logsentinel/internal/repository/team"
// 	pb "github.com/AjayShukla007/logsentinel/proto/gen/proto"
// 	"github.com/google/uuid"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// type TeamService struct {
// 	pb.UnimplementedTeamServiceServer
// 	repo team.Repository
// }

// func NewTeamService(repo team.Repository) *TeamService {
// 	return &TeamService{repo: repo}
// }

// func (s *TeamService) CreateTeam(ctx context.Context, req *pb.CreateTeamRequest) (*pb.CreateTeamResponse, error) {
// 	if req.Name == "" {
// 		return nil, status.Error(codes.InvalidArgument, "team name is required")
// 	}
// 	if req.OwnerId == "" {
// 		return nil, status.Error(codes.InvalidArgument, "owner ID is required")
// 	}

// 	team, err := s.repo.Create(ctx, req.Name, req.OwnerId)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to create team: %v", err)
// 	}

// 	return &pb.CreateTeamResponse{
// 		TeamId:    team.ID.String(),
// 		Name:      team.Name,
// 		OwnerId:   team.OwnerID,
// 		CreatedAt: team.CreatedAt.Format(time.RFC3339),
// 	}, nil
// }

// func (s *TeamService) GetTeam(ctx context.Context, req *pb.GetTeamRequest) (*pb.GetTeamResponse, error) {
// 	if req.TeamId == "" {
// 		return nil, status.Error(codes.InvalidArgument, "team ID is required")
// 	}

// 	teamID, err := uuid.Parse(req.TeamId)
// 	if err != nil {
// 		return nil, status.Error(codes.InvalidArgument, "invalid team ID")
// 	}

// 	team, err := s.repo.GetByID(ctx, teamID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.NotFound, "team not found: %v", err)
// 	}

// 	return &pb.GetTeamResponse{
// 		TeamId:    team.ID.String(),
// 		Name:      team.Name,
// 		OwnerId:   team.OwnerID,
// 		CreatedAt: team.CreatedAt.Format(time.RFC3339),
// 	}, nil
// }

// func (s *TeamService) DeleteTeam(ctx context.Context, req *pb.DeleteTeamRequest) (*pb.DeleteTeamResponse, error) {
// 	if req.TeamId == "" {
// 		return nil, status.Error(codes.InvalidArgument, "team ID is required")
// 	}

// 	teamID, err := uuid.Parse(req.TeamId)
// 	if err != nil {
// 		return nil, status.Error(codes.InvalidArgument, "invalid team ID")
// 	}

// 	err = s.repo.Delete(ctx, teamID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to delete team: %v", err)
// 	}

// 	return &pb.DeleteTeamResponse{Success: true}, nil
// }

func NewTeamService() {
	println("NewTeamService")
	return
}
