package database

import (
	"database/sql"
	"time"

	"lan-relay/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func Init(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}

	if err := db.createTables(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS log_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		source_ip TEXT,
		method TEXT,
		target_host TEXT,
		target_port TEXT,
		path TEXT,
		status_code INTEGER,
		duration_ms INTEGER,
		error TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_timestamp ON log_entries(timestamp);
	CREATE INDEX IF NOT EXISTS idx_target_host ON log_entries(target_host);

	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		ngrok_token TEXT DEFAULT '',
		ngrok_domain TEXT DEFAULT '',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	INSERT OR IGNORE INTO settings (id, ngrok_token, ngrok_domain) VALUES (1, '', '');
	`

	_, err := db.conn.Exec(query)
	return err
}

func (db *DB) InsertLogEntry(entry *models.LogEntry) error {
	query := `
	INSERT INTO log_entries (timestamp, source_ip, method, target_host, target_port, path, status_code, duration_ms, error)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query,
		entry.Timestamp,
		entry.SourceIP,
		entry.Method,
		entry.TargetHost,
		entry.TargetPort,
		entry.Path,
		entry.StatusCode,
		entry.Duration,
		entry.Error,
	)

	return err
}

func (db *DB) GetLogs(limit, offset int) ([]models.LogEntry, error) {
	query := `
	SELECT id, timestamp, source_ip, method, target_host, target_port, path, status_code, duration_ms, COALESCE(error, '')
	FROM log_entries
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?
	`

	rows, err := db.conn.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]models.LogEntry, 0)
	for rows.Next() {
		var log models.LogEntry
		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.SourceIP,
			&log.Method,
			&log.TargetHost,
			&log.TargetPort,
			&log.Path,
			&log.StatusCode,
			&log.Duration,
			&log.Error,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

func (db *DB) ClearLogs() error {
	_, err := db.conn.Exec("DELETE FROM log_entries")
	return err
}

func (db *DB) GetLogCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM log_entries").Scan(&count)
	return count, err
}

func (db *DB) GetRecentLogCount(since time.Time) (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM log_entries WHERE timestamp >= ?", since).Scan(&count)
	return count, err
}

func (db *DB) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	query := `SELECT id, ngrok_token, ngrok_domain, updated_at FROM settings WHERE id = 1`

	err := db.conn.QueryRow(query).Scan(
		&settings.ID,
		&settings.NgrokToken,
		&settings.NgrokDomain,
		&settings.UpdatedAt,
	)

	return &settings, err
}

func (db *DB) UpdateSettings(settings *models.Settings) error {
	query := `
	UPDATE settings 
	SET ngrok_token = ?, ngrok_domain = ?, updated_at = CURRENT_TIMESTAMP 
	WHERE id = 1
	`

	_, err := db.conn.Exec(query, settings.NgrokToken, settings.NgrokDomain)
	return err
}
