package repository

import "time"

type Log struct {
	ID            int       `json:"id"`
	TransactionID string    `json:"transaction_id"`
	Amount        string    `json:"amount"`
	Status        string    `json:"status"`
	Time          time.Time `json:"time"`
}

func NewLog(id string, amount string, status string) *Log {
	return &Log{
		TransactionID: id,
		Amount:        amount,
		Status:        status,
	}
}

type LogRepository interface {
	InsertLog(log *Log) error
}

type LogRepositoryImpl struct {
	db *PostgresDB
}

func NewLogRepository(db *PostgresDB) LogRepository {
	return &LogRepositoryImpl{
		db: db,
	}
}

func (l *LogRepositoryImpl) InsertLog(log *Log) error {
	query := `INSERT INTO public.logs (transaction_id, amount, status, time) VALUES ($1, $2, $3, $4) RETURNING id`
	l.db.Lock()
	defer l.db.Unlock()
	err := l.db.db.QueryRow(query, log.TransactionID, log.Amount, log.Status, log.Time).Scan(&log.ID)
	if err != nil {
		return err
	}
	return nil
}
