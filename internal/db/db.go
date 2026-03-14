package db

import (
	"database/sql"
	"fmt"
	"time"

	"metric-collector/internal/models"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// WAL mode for better concurrent read/write
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			cpu_percent REAL,
			memory_percent REAL,
			disk_percent REAL,
			net_bytes_sent INTEGER,
			net_bytes_recv INTEGER,
			load_avg REAL
		)
	`); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Insert(m models.Metrics) error {
	_, err := s.db.Exec(`
		INSERT INTO metrics (timestamp, cpu_percent, memory_percent, disk_percent, net_bytes_sent, net_bytes_recv, load_avg)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.Timestamp.UTC(), m.CPUPercent, m.MemoryPercent, m.DiskPercent, m.NetBytesSent, m.NetBytesRecv, m.LoadAvg,
	)
	return err
}

func (s *Store) Query(minutes int) ([]models.Metrics, error) {
	since := time.Now().UTC().Add(-time.Duration(minutes) * time.Minute)
	rows, err := s.db.Query(`
		SELECT id, timestamp, cpu_percent, memory_percent, disk_percent, net_bytes_sent, net_bytes_recv, load_avg
		FROM metrics WHERE timestamp >= ? ORDER BY timestamp ASC`, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Metrics
	for rows.Next() {
		var m models.Metrics
		var ts string
		if err := rows.Scan(&m.ID, &ts, &m.CPUPercent, &m.MemoryPercent, &m.DiskPercent, &m.NetBytesSent, &m.NetBytesRecv, &m.LoadAvg); err != nil {
			return nil, err
		}
		m.Timestamp, _ = time.Parse("2006-01-02 15:04:05+00:00", ts)
		if m.Timestamp.IsZero() {
			m.Timestamp, _ = time.Parse("2006-01-02T15:04:05Z", ts)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (s *Store) Close() error {
	return s.db.Close()
}
