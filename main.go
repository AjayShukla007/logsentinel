package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userrepo "github.com/AjayShukla007/logsentinel/internal/repository/user"
	cronservice "github.com/AjayShukla007/logsentinel/internal/service/cron"
	logservice "github.com/AjayShukla007/logsentinel/internal/service/log"
	projectservice "github.com/AjayShukla007/logsentinel/internal/service/project"
	userservice "github.com/AjayShukla007/logsentinel/internal/service/user"

	// teamservice "github.com/AjayShukla007/logsentinel/internal/service/team"

	projectrepo "github.com/AjayShukla007/logsentinel/internal/repository/project"
	// teamrepo "github.com/AjayShukla007/logsentinel/internal/repository/team"

	pb "github.com/AjayShukla007/logsentinel/proto/gen/proto"
)

const (
	port = ":50051"
)

func getDatabaseURL() string {
    host := getEnvWithDefault("DB_HOST", "db")
    dbPortStr := getEnvWithDefault("DB_PORT", "5432")
    dbPort, err := strconv.Atoi(dbPortStr)
    if err != nil {
        log.Fatalf("Invalid DB_PORT: %v", err)
    }
    user := getEnvWithDefault("DB_USER", "postgres")
    dbName := getEnvWithDefault("DB_NAME", "logsentinel")

    passwordBytes, err := os.ReadFile("/run/secrets/db-password")
    if err != nil {
        log.Fatalf("failed to read database password: %v", err)
    }
    password := strings.TrimSpace(string(passwordBytes))

    dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
        user, password, host, dbPort, dbName)

    log.Printf("Connecting to database at %s:%d", host, dbPort)
    return dsn
}

func getEnvWithDefault(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func getLocalDatabaseURL() string {
	host := os.Getenv("LOCAL_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("LOCAL_DB_PORT")
	if portStr == "" {
		portStr = "5432"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid LOCAL_DB_PORT: %v", err)
	}

	user := os.Getenv("LOCAL_DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("LOCAL_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbName := os.Getenv("LOCAL_DB_NAME")
	if dbName == "" {
		dbName = "logsentinel"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, dbName)

	log.Printf("Connecting to local database at %s:%d", host, port)
	return dsn
}

func main() {
    log.Println("Starting LogSentinel server...")
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: Error loading .env file, using defaults")
    }
	var dsn string
    if os.Getenv("DB_HOST") != "" {
        log.Println("Initializing Docker/Compose database connection...")
        dsn = getDatabaseURL()
    } else {
        log.Println("Initializing local database connection...")
        dsn = getLocalDatabaseURL()
    }
    dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer dbpool.Close()

	log.Println("Testing database connection...")
	var version string
	err = dbpool.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query database: %v\n", err)
	}
	log.Printf("Connected to PostgreSQL: %s", version)

	// teamRepository := teamrepo.NewPostgresRepository(dbpool)
	projectRepository := projectrepo.NewPostgresRepository(dbpool)
	userRepository := userrepo.NewPostgresRepository(dbpool)

	// teamSvc := teamservice.NewTeamService(teamRepository)
	projectSvc := projectservice.NewProjectService(projectRepository, userRepository)
	userSvc := userservice.NewUserService(userRepository)
	logSvc := logservice.NewLogService(dbpool)
	cronSvc := cronservice.NewCronService(dbpool)
	cronSvc.Start()

	log.Printf("Initializing gRPC server on port %s...", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("unable to create connection pool: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterLogServiceServer(s, logSvc)
	// pb.RegisterTeamServiceServer(s, teamSvc)
	pb.RegisterProjectServiceServer(s, projectSvc)
	pb.RegisterUserServiceServer(s, userSvc)

    if os.Getenv("ENABLE_GRPC_REFLECTION") == "true" {
        reflection.Register(s)
        log.Println("gRPC reflection enabled")
    }

	log.Printf("Server listening on port %s\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
