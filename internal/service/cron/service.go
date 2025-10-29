package cron

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CronService struct {
	db *pgxpool.Pool
}

func NewCronService(db *pgxpool.Pool) *CronService {
	return &CronService{
		db: db,
	}
}

func (s *CronService) Start() {
	go s.scheduleCleanup()
}

func (s *CronService) scheduleCleanup() {
	// For testing, using a shorter interval
	// ticker := time.NewTicker(1 * time.Minute) // Run every minute instead of every 24 hours
	//TODO: change the interval value to 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		s.deleteOldLogs()
		<-ticker.C
	}
}

func (s *CronService) deleteOldLogs() {
	query := `
        DELETE FROM logs l
        USING projects p, users u
        WHERE l.project_id = p.id 
        AND p.user_id = u.id 
        AND u.account_type = 'free'
        AND l.created_at < NOW() - INTERVAL '7 days'
    `

	result, err := s.db.Exec(context.Background(), query)
	if err != nil {
		log.Printf("Error deleting old logs: %v", err)
		return
	}

	rowsAffected := result.RowsAffected()
	log.Printf("Deleted %d old logs from free tier users", rowsAffected)
}
