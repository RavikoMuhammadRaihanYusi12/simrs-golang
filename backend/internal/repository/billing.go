package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"

	"github.com/thinkingx/simrs-backend/internal/domain"
)

var (
	ErrBillingNotFound   = errors.New("billing not found")
	ErrClaimNotFound     = errors.New("claim not found")
	ErrPaymentNotFound   = errors.New("payment not found")
	ErrStockInsufficient = errors.New("stock insufficient")
)

type BillingRepository interface {
	CreateBilling(ctx context.Context, b domain.Billing) error
	GetBilling(ctx context.Context, id string) (*domain.Billing, error)
	GetBillingsByPatient(ctx context.Context, patientID string) ([]domain.Billing, error)
	GetBillingsByEncounter(ctx context.Context, encounterID string) ([]domain.Billing, error)
	UpdateBillingStatus(ctx context.Context, id string, status string) error
	CreatePayment(ctx context.Context, p domain.Payment) error
	GetPaymentsByBilling(ctx context.Context, billingID string) ([]domain.Payment, error)
	CreateClaim(ctx context.Context, c domain.BPJSClaim) error
	GetClaimStatus(ctx context.Context, noSEP string) (*domain.BPJSClaim, error)
}

type SQLiteBillingRepository struct {
	db *sql.DB
}

func NewSQLiteBillingRepository(dsn string) (*SQLiteBillingRepository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	s := &SQLiteBillingRepository{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteBillingRepository) initSchema() error {
	scripts := []string{
		`CREATE TABLE IF NOT EXISTS billings (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			encounter_id TEXT NOT NULL,
			total INTEGER NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS invoices (
			id TEXT PRIMARY KEY,
			billing_id TEXT NOT NULL,
			amount INTEGER NOT NULL,
			paid_at TEXT,
			FOREIGN KEY (billing_id) REFERENCES billings(id)
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id TEXT PRIMARY KEY,
			billing_id TEXT NOT NULL,
			amount INTEGER NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS bpjs_claims (
			id TEXT PRIMARY KEY,
			no_sep TEXT NOT NULL,
			patient_id TEXT NOT NULL,
			total INTEGER NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
	}
	for _, script := range scripts {
		if _, err := s.db.Exec(script); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteBillingRepository) CreateBilling(ctx context.Context, b domain.Billing) error {
	const query = "INSERT INTO billings (id, patient_id, encounter_id, total, status, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, query, b.ID, b.PatientID, b.EncounterID, b.Total, b.Status, now)
	if err != nil {
		return ErrBillingNotFound
	}
	return nil
}

func (s *SQLiteBillingRepository) GetBilling(ctx context.Context, id string) (*domain.Billing, error) {
	const query = "SELECT id, patient_id, encounter_id, total, status, created_at FROM billings WHERE id = ?"
	row, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var b domain.Billing
	if row.Next() {
		if err := row.Scan(&b.ID, &b.PatientID, &b.EncounterID, &b.Total, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
	}
	return &b, row.Err()
}

func (s *SQLiteBillingRepository) CreateClaim(ctx context.Context, c domain.BPJSClaim) error {
	const query = "INSERT INTO bpjs_claims (id, no_sep, patient_id, total, status, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, query, c.ID, c.NoSEP, c.PatientID, c.Amount, c.Status, now)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteBillingRepository) GetClaimStatus(ctx context.Context, noSEP string) (*domain.BPJSClaim, error) {
	const query = "SELECT id, no_sep, patient_id, total, status, created_at FROM bpjs_claims WHERE no_sep = ?"
	row, err := s.db.QueryContext(ctx, query, noSEP)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if row.Next() {
		var c domain.BPJSClaim
		if err := row.Scan(&c.ID, &c.NoSEP, &c.PatientID, &c.Amount, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		return &c, nil
	}
	return nil, ErrClaimNotFound
}

func (s *SQLiteBillingRepository) GetBillingsByPatient(ctx context.Context, patientID string) ([]domain.Billing, error) {
	const query = "SELECT id, patient_id, encounter_id, total, status, created_at FROM billings WHERE patient_id = ? ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Billing{}
	for rows.Next() {
		var b domain.Billing
		if err := rows.Scan(&b.ID, &b.PatientID, &b.EncounterID, &b.Total, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *SQLiteBillingRepository) GetBillingsByEncounter(ctx context.Context, encounterID string) ([]domain.Billing, error) {
	const query = "SELECT id, patient_id, encounter_id, total, status, created_at FROM billings WHERE encounter_id = ? ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Billing{}
	for rows.Next() {
		var b domain.Billing
		if err := rows.Scan(&b.ID, &b.PatientID, &b.EncounterID, &b.Total, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *SQLiteBillingRepository) UpdateBillingStatus(ctx context.Context, id string, status string) error {
	const query = "UPDATE billings SET status = ? WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLiteBillingRepository) Close() error {
	return s.db.Close()
}

func (s *SQLiteBillingRepository) CreatePayment(ctx context.Context, p domain.Payment) error {
	const query = "INSERT INTO payments (id, billing_id, amount, created_at) VALUES (?, ?, ?, ?)"
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, query, p.ID, p.BillingID, p.Amount, now)
	if err != nil {
		return ErrPaymentNotFound
	}
	return nil
}

func (s *SQLiteBillingRepository) GetPaymentsByBilling(ctx context.Context, billingID string) ([]domain.Payment, error) {
	const query = "SELECT id, billing_id, amount, created_at FROM payments WHERE billing_id = ? ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query, billingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Payment{}
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.BillingID, &p.Amount, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
