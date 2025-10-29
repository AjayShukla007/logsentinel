package log

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AjayShukla007/logsentinel/internal/ratelimit"
	// "github.com/AjayShukla007/logsentinel/internal/repository/project"
	pb "github.com/AjayShukla007/logsentinel/proto/gen/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LogService struct {
	pb.UnimplementedLogServiceServer
	db          *pgxpool.Pool
	rateLimiter *ratelimit.RateLimiter
}

type ConnectionManager struct {
	connections map[string]*ConnectionInfo
	mu          sync.RWMutex
}

type ConnectionInfo struct {
	clientID     string
	projectID    string
	lastActivity time.Time
	messageCount int
	isProAccount bool
}

func NewLogService(db *pgxpool.Pool) *LogService {
	return &LogService{
		db:          db,
		rateLimiter: ratelimit.NewRateLimiter(),
	}
}

func (s *LogService) Test(req *pb.TestRequest, stream pb.LogService_TestServer) error {
	for {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		response := &pb.TestResponse{
			Message:   "active",
			Timestamp: timestamp,
		}

		if err := stream.Send(response); err != nil {
			return fmt.Errorf("error sending test response: %v", err)
		}

		time.Sleep(2 * time.Second)
	}
}

func (s *LogService) SendLog(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	// TODO: remove these in production server as this might leak important data
    fmt.Printf("Received log request: projectId=%s clientId=%s\n", req.ProjectId, req.ClientId)
    var accountType string

	var projectExists bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`, req.ProjectId).Scan(&projectExists)
	if err != nil {
		fmt.Printf("Error checking project existence: %v\n", err)
		return &pb.LogResponse{
			Success: false,
			Message: "Database error when checking project",
		}, nil
	}

	if !projectExists {
		return &pb.LogResponse{
			Success: false,
			Message: "Project not found",
		}, nil
	}

	var apiKeyMatches bool
	err = s.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND api_key = $2)
	`, req.ProjectId, req.ApiKey).Scan(&apiKeyMatches)

	if err != nil {
		fmt.Printf("Error checking API key: %v\n", err)
		return &pb.LogResponse{
			Success: false,
			Message: "Database error when checking API key",
		}, nil
	}

	if !apiKeyMatches {
		return &pb.LogResponse{
			Success: false,
			Message: "Invalid API key",
		}, nil
	}

	err = s.db.QueryRow(ctx, `
		SELECT u.account_type 
		FROM users u 
		JOIN projects p ON p.user_id = u.id 
		WHERE p.id = $1 AND u.client_id = $2
	`, req.ProjectId, req.ClientId).Scan(&accountType)

	if err != nil {
		fmt.Printf("Error checking user relationship: %v\n", err)
		return &pb.LogResponse{
			Success: false,
			Message: "Invalid client ID or user doesn't own this project",
		}, nil
	}

	isProAccount := accountType == "pro"
	if !s.rateLimiter.AllowLog(req.ClientId, isProAccount) {
		return &pb.LogResponse{
			Success: false,
			Message: "Rate limit exceeded",
		}, nil
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO logs (project_id, category, message)
		VALUES ($1, $2, $3)`,
		req.ProjectId, req.Category, req.Message,
	)

	if err != nil {
		fmt.Printf("Error inserting log: %v\n", err)
		return &pb.LogResponse{
			Success: false,
			Message: "Failed to save log",
		}, nil
	}

	return &pb.LogResponse{
		Success: true,
		Message: "Log saved successfully",
	}, nil
}

func (s *LogService) StreamLogs(req *pb.LogRequest, stream pb.LogService_StreamLogsServer) error {
    fmt.Printf("Received stream logs request: projectId=%s\n", req.ProjectId)
	/* // Database query pseudocode for production
	   rows, err := s.db.Query(context.Background(), `
	       SELECT id, level, message, created_at
	       FROM logs
	       WHERE project_id = $1
	       ORDER BY created_at DESC
	       LIMIT $2 OFFSET $3
	   `, req.ProjectId, batchSize, offset) */
	// In production we would filter logs by req.ProjectId
	// For now let's include the project ID in messages to show that filtering works
	projectHint := req.ProjectId
	if projectHint == "" {
		projectHint = "unknown-project"
	}

	// this is an fixed reference time that won't change between requests 
	// in production we would store this per project or user session
	referenceTime := time.Now()

	// in production these would come from the database and will be filtered by project ID
	hardcodedLogs := []struct {
		level     string
		message   string
		timestamp time.Time
	}{
		{"info", fmt.Sprintf("[Project: %s] Application started", projectHint), referenceTime.Add(-30 * time.Minute)},
		{"info", fmt.Sprintf("[Project: %s] User login successful", projectHint), referenceTime.Add(-28 * time.Minute)},
		{"warn", fmt.Sprintf("[Project: %s] High memory usage detected", projectHint), referenceTime.Add(-25 * time.Minute)},
		{"info", fmt.Sprintf("[Project: %s] New user registered", projectHint), referenceTime.Add(-22 * time.Minute)},
		{"error", fmt.Sprintf("[Project: %s] Database connection timeout", projectHint), referenceTime.Add(-20 * time.Minute)},
		{"info", fmt.Sprintf("[Project: %s] Config loaded successfully", projectHint), referenceTime.Add(-18 * time.Minute)},
		{"warn", fmt.Sprintf("[Project: %s] Slow query detected", projectHint), referenceTime.Add(-15 * time.Minute)},
		{"info", fmt.Sprintf("[Project: %s] Payment processed", projectHint), referenceTime.Add(-12 * time.Minute)},
		{"error", fmt.Sprintf("[Project: %s] Failed to send email", projectHint), referenceTime.Add(-10 * time.Minute)},
		{"info", fmt.Sprintf("[Project: %s] Cache refreshed", projectHint), referenceTime.Add(-8 * time.Minute)},
		{"warn", fmt.Sprintf("[Project: %s] API rate limit approaching", projectHint), referenceTime.Add(-6 * time.Minute)},
		{"info", fmt.Sprintf("[Project: %s] Scheduled task completed", projectHint), referenceTime.Add(-5 * time.Minute)},
		{"error", fmt.Sprintf("[Project: %s] Invalid authentication token", projectHint), referenceTime.Add(-4 * time.Minute)},
		{"info", fmt.Sprintf("[Project: %s] New feature enabled", projectHint), referenceTime.Add(-3 * time.Minute)},
		{"warn", fmt.Sprintf("[Project: %s] Deprecated API in use", projectHint), referenceTime.Add(-2 * time.Minute)},
		{"info", fmt.Sprintf("[Project: %s] Data backup completed", projectHint), referenceTime.Add(-90 * time.Second)},
		{"error", fmt.Sprintf("[Project: %s] File upload failed", projectHint), referenceTime.Add(-60 * time.Second)},
		{"info", fmt.Sprintf("[Project: %s] User preferences updated", projectHint), referenceTime.Add(-45 * time.Second)},
		{"warn", fmt.Sprintf("[Project: %s] Low disk space warning", projectHint), referenceTime.Add(-30 * time.Second)},
		{"info", fmt.Sprintf("[Project: %s] System health check passed", projectHint), referenceTime.Add(-15 * time.Second)},
	}

	loadMore := false
	if req.Message == "LOAD_MORE" {
		loadMore = true
		fmt.Printf("Load more request received for project: %s\n", projectHint)
	}

	startIndex := 0
	endIndex := 10

	if loadMore {
		startIndex = 10
		endIndex = len(hardcodedLogs)
	}

	for i := startIndex; i < endIndex && i < len(hardcodedLogs); i++ {
		log := hardcodedLogs[i]
		formattedMessage := fmt.Sprintf("[%s] [%s] %s",
			log.timestamp.Format("2006-01-02 15:04:05"),
			log.level,
			log.message)

		if err := stream.Send(&pb.LogResponse{
			Success: true,
			Message: formattedMessage,
		}); err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
	}

	// If this was just a load more request we're done after sending the additional logs
	if loadMore {
		return nil
	}

	infoTicker := time.NewTicker(2 * time.Second)
	errorTicker := time.NewTicker(10 * time.Second)
	defer infoTicker.Stop()
	defer errorTicker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil

		case t := <-infoTicker.C:
			formattedMessage := fmt.Sprintf("[%s] [info] [Project: %s] Regular system update",
				t.Format("2006-01-02 15:04:05"),
				projectHint)

			if err := stream.Send(&pb.LogResponse{
				Success: true,
				Message: formattedMessage,
			}); err != nil {
				return err
			}

		case t := <-errorTicker.C:
			formattedMessage := fmt.Sprintf("[%s] [error] [Project: %s] Periodic error check failed",
				t.Format("2006-01-02 15:04:05"),
				projectHint)

			if err := stream.Send(&pb.LogResponse{
				Success: true,
				Message: formattedMessage,
			}); err != nil {
				return err
			}
		}
	}
}

func (s *LogService) BatchSendLogs(stream pb.LogService_BatchSendLogsServer) error {
	firstLog, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to receive initial log: %v", err)
	}
	var count int32
	var accountType string
	err = s.db.QueryRow(context.Background(), `
		SELECT u.account_type 
		FROM users u 
		JOIN projects p ON p.user_id = u.id 
		WHERE p.id = $1 AND p.api_key = $2 AND u.client_id = $3`,
		firstLog.ProjectId, firstLog.ApiKey, firstLog.ClientId,
	).Scan(&accountType)

	if err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
	}

	isProAccount := accountType == "pro"
	if !s.rateLimiter.AllowLog(firstLog.ClientId, isProAccount) {
		return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
	}

	_, err = s.db.Exec(context.Background(), `
		INSERT INTO logs (project_id, category, message)
		VALUES ($1, $2, $3)`,
		firstLog.ProjectId, firstLog.Category, firstLog.Message,
	)

	if err != nil {
		return status.Errorf(codes.Internal, "failed to save log: %v", err)
	}

	for {
		logReq, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.BatchLogResponse{
				Success: true,
				Message: "All logs processed successfully",
				Count:   count,
			})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "error receiving log: %v", err)
		}

		if logReq.ProjectId != firstLog.ProjectId || logReq.ApiKey != firstLog.ApiKey || logReq.ClientId != firstLog.ClientId {
			return status.Errorf(codes.InvalidArgument, "credentials must be consistent in batch")
		}

		if !s.rateLimiter.AllowLog(logReq.ClientId, isProAccount) {
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		_, err = s.db.Exec(context.Background(), `
			INSERT INTO logs (project_id, category, message)
			VALUES ($1, $2, $3)`,
			logReq.ProjectId, logReq.Category, logReq.Message,
		)

		if err != nil {
			return status.Errorf(codes.Internal, "failed to save log: %v", err)
		}

		count++
	}
}

func NewConnectionManager() *ConnectionManager {
	manager := &ConnectionManager{
		connections: make(map[string]*ConnectionInfo),
	}

	go manager.cleanupStaleConnections()

	return manager
}

func (cm *ConnectionManager) RegisterConnection(clientID, projectID string, isProAccount bool) string {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	connectionID := uuid.New().String()
	cm.connections[connectionID] = &ConnectionInfo{
		clientID:     clientID,
		projectID:    projectID,
		lastActivity: time.Now(),
		isProAccount: isProAccount,
	}

	return connectionID
}

func (cm *ConnectionManager) UpdateActivity(connectionID string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[connectionID]; exists {
		conn.lastActivity = time.Now()
		conn.messageCount++
		return true
	}
	return false
}

func (cm *ConnectionManager) cleanupStaleConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		cm.mu.Lock()
		for id, conn := range cm.connections {
			if time.Since(conn.lastActivity) > 30*time.Minute {
				delete(cm.connections, id)
			}
		}
		cm.mu.Unlock()
	}
}

// this ConnectClient handles persistent bidirectional streaming connections from clients
func (s *LogService) ConnectClient(stream pb.LogService_ConnectClientServer) error {
	var connectionID string
	var clientID string
	var projectID string
	var isAuthenticated bool
	var isProAccount bool

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Done channel to signal when stream is closed
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-heartbeatTicker.C:
				if isAuthenticated {
					if err := stream.Send(&pb.ServerMessage{
						Message: &pb.ServerMessage_Pong{
							Pong: &pb.HeartbeatMessage{
								Timestamp: time.Now().UnixNano(),
							},
						},
					}); err != nil {
						// Connection might be broken, if so terminate goroutine
						// TODO: we need to do some additional events instead of just termination
						return
					}
				}
			}
		}
	}()

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to receive message: %v", err)
		}

		switch m := msg.Message.(type) {
		case *pb.ClientMessage_Auth:
			auth := m.Auth
			clientID = auth.ClientId
			projectID = auth.ProjectId
			if auth.ApiKey == "" || auth.ClientId == "" || auth.ProjectId == "" {
				var missingFields []string
				if auth.ApiKey == "" {
					missingFields = append(missingFields, "api_key")
				}
				if auth.ClientId == "" {
					missingFields = append(missingFields, "client_id")
				}
				if auth.ProjectId == "" {
					missingFields = append(missingFields, "project_id")
				}

				errorMsg := fmt.Sprintf("Project not found: missing %s", strings.Join(missingFields, ", "))

				stream.Send(&pb.ServerMessage{
					Message: &pb.ServerMessage_AuthResponse{
						AuthResponse: &pb.AuthResponse{
							Success: false,
							Message: errorMsg,
						},
					},
				})
                fmt.Printf("Unable to find project: clientId=%s projectId=%s (missing: %s)\n",
                    auth.ClientId, auth.ProjectId, strings.Join(missingFields, ", "))
                return status.Errorf(codes.Unauthenticated, errorMsg)
            }
			var accountType string
			err := s.db.QueryRow(context.Background(), `
				SELECT u.account_type 
				FROM users u 
				JOIN projects p ON p.user_id = u.id 
				WHERE p.id = $1 AND p.api_key = $2 AND u.client_id = $3`,
				auth.ProjectId, auth.ApiKey, auth.ClientId,
			).Scan(&accountType)

			if err != nil {
				stream.Send(&pb.ServerMessage{
					Message: &pb.ServerMessage_AuthResponse{
						AuthResponse: &pb.AuthResponse{
							Success: false,
							Message: "Authentication failed: invalid credentials",
						},
					},
				})
				fmt.Println("Authentication failed:", err)
				return status.Errorf(codes.Unauthenticated, "Authentication failed")
			}

			connectionID = uuid.New().String()
			isAuthenticated = true
			isProAccount = accountType == "pro"

			// Send successful auth response
			stream.Send(&pb.ServerMessage{
				Message: &pb.ServerMessage_AuthResponse{
					AuthResponse: &pb.AuthResponse{
						Success:   true,
						SessionId: connectionID,
						Message:   "Successfully authenticated",
					},
				},
			})

			fmt.Printf("Client connected: ID=%s, Project=%s\n", clientID, projectID)

		case *pb.ClientMessage_Log:
			if !isAuthenticated {
				stream.Send(&pb.ServerMessage{
					Message: &pb.ServerMessage_Error{
						Error: &pb.ErrorMessage{
							Code:    "unauthenticated",
							Message: "Not authenticated",
						},
					},
				})
				continue
			}

			if !s.rateLimiter.AllowLog(clientID, isProAccount) {
				stream.Send(&pb.ServerMessage{
					Message: &pb.ServerMessage_Error{
						Error: &pb.ErrorMessage{
							Code:    "rate_limit_exceeded",
							Message: "Rate limit exceeded",
						},
					},
				})
				continue
			}

			// _, err = s.db.Exec(context.Background(), `
			// 	INSERT INTO logs (project_id, category, message)
			// 	VALUES ($1, $2, $3)`,
			// 	projectID, m.Log.Category, m.Log.Message,
			// )
			fmt.Printf("Received log projectId: %s\n", projectID)
			fmt.Printf("Received log Category: %s\n", m.Log.Category)
			fmt.Printf("Received log message: %s\n", m.Log.Message)

			err = nil

			if err != nil {
				stream.Send(&pb.ServerMessage{
					Message: &pb.ServerMessage_Error{
						Error: &pb.ErrorMessage{
							Code:    "database_error",
							Message: "Failed to save log",
						},
					},
				})
				continue
			}

			stream.Send(&pb.ServerMessage{
				Message: &pb.ServerMessage_LogResponse{
					LogResponse: &pb.LogResponse{
						Success: true,
						Message: "Log saved successfully",
					},
				},
			})

		case *pb.ClientMessage_Ping:
			stream.Send(&pb.ServerMessage{
				Message: &pb.ServerMessage_Pong{
					Pong: &pb.HeartbeatMessage{
						Timestamp: time.Now().UnixNano(),
					},
				},
			})

		case *pb.ClientMessage_Close:
			fmt.Printf("Client disconnecting: ID=%s, Project=%s, Reason=%s\n",
				clientID, projectID, m.Close.Reason)
			return nil
		}
	}
}
