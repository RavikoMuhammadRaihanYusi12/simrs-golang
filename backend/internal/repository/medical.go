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
	ErrMedicalRecordNotFound = errors.New("medical record not found")
	ErrMedicineNotFound      = errors.New("medicine not found")
	ErrPrescriptionNotFound  = errors.New("prescription not found")
	ErrMedicineStockError    = errors.New("medicine stock insufficient")
)

type MedicalRepository interface {
	CreateMedicalRecord(ctx context.Context, mr domain.MedicalRecord) error
	GetMedicalRecordsByEncounter(ctx context.Context, encounterID string) ([]domain.MedicalRecord, error)
	CreateMedicine(ctx context.Context, med domain.Medicine) error
	GetMedicine(ctx context.Context, id string) (*domain.Medicine, error)
	ListMedicines(ctx context.Context) ([]domain.Medicine, error)
	UpdateMedicine(ctx context.Context, med domain.Medicine) error
	DeleteMedicine(ctx context.Context, id string) error
	UpdateMedicineStock(ctx context.Context, id string, delta int) error
	CreateMedicineTransaction(ctx context.Context, t domain.MedicineTransaction) error
	ListMedicineTransactions(ctx context.Context) ([]domain.MedicineTransaction, error)
	CreatePrescription(ctx context.Context, p domain.Prescription) error
	GetPrescriptionsByEncounter(ctx context.Context, encounterID string) ([]domain.Prescription, error)
	ListPrescriptions(ctx context.Context) ([]domain.Prescription, error)
	UpdatePrescriptionStatus(ctx context.Context, id, status string) error
}

type SQLiteMedicalRepository struct {
	db *sql.DB
}

func NewSQLiteMedicalRepository(dsn string) (*SQLiteMedicalRepository, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	s := &SQLiteMedicalRepository{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLiteMedicalRepository) initSchema() error {
	scripts := []string{
		`CREATE TABLE IF NOT EXISTS medical_records (
			id TEXT PRIMARY KEY,
			encounter_id TEXT NOT NULL,
			subjective TEXT NOT NULL,
			objective TEXT NOT NULL,
			assessment TEXT NOT NULL,
			plan TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (encounter_id) REFERENCES encounters(id)
		)`,
		`CREATE TABLE IF NOT EXISTS medicines (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			stock INTEGER NOT NULL,
			unit TEXT NOT NULL,
			price DECIMAL NOT NULL,
			expired_date TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS medicine_transactions (
			id TEXT PRIMARY KEY,
			medicine_id TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			type TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			FOREIGN KEY (medicine_id) REFERENCES medicines(id)
		)`,
		`CREATE TABLE IF NOT EXISTS prescriptions (
			id TEXT PRIMARY KEY,
			encounter_id TEXT NOT NULL,
			medicine_id TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			quantity INTEGER NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (encounter_id) REFERENCES encounters(id),
			FOREIGN KEY (medicine_id) REFERENCES medicines(id)
		)`,
	}
	for _, script := range scripts {
		if _, err := s.db.Exec(script); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteMedicalRepository) CreateMedicalRecord(ctx context.Context, mr domain.MedicalRecord) error {
	const query = "INSERT INTO medical_records (id, encounter_id, subjective, objective, assessment, plan, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, query, mr.ID, mr.EncounterID, mr.Subjective, mr.Objective, mr.Assessment, mr.Plan, now)
	if err != nil {
		return ErrMedicalRecordNotFound
	}
	return nil
}

func (s *SQLiteMedicalRepository) GetMedicalRecordsByEncounter(ctx context.Context, encounterID string) ([]domain.MedicalRecord, error) {
	const query = "SELECT id, subjective, objective, assessment, plan, created_at FROM medical_records WHERE encounter_id = ? ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.MedicalRecord{}
	for rows.Next() {
		var mr domain.MedicalRecord
		if err := rows.Scan(&mr.ID, &mr.Subjective, &mr.Objective, &mr.Assessment, &mr.Plan, &mr.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, mr)
	}
	return out, rows.Err()
}

func (s *SQLiteMedicalRepository) CreateMedicine(ctx context.Context, med domain.Medicine) error {
	const query = "INSERT INTO medicines (id, name, description, stock, unit, price, expired_date) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err := s.db.ExecContext(ctx, query, med.ID, med.Name, med.Description, med.Stock, med.Unit, med.Price, med.ExpiredDate)
	if err != nil {
		return ErrMedicalRecordNotFound
	}
	return nil
}

func (s *SQLiteMedicalRepository) GetMedicine(ctx context.Context, id string) (*domain.Medicine, error) {
	const query = "SELECT id, name, description, stock, unit, price, expired_date FROM medicines WHERE id = ?"
	row, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if row.Next() {
		var med domain.Medicine
		if err := row.Scan(&med.ID, &med.Name, &med.Description, &med.Stock, &med.Unit, &med.Price, &med.ExpiredDate); err != nil {
			return nil, err
		}
		return &med, nil
	}
	return nil, ErrMedicineNotFound
}

func (s *SQLiteMedicalRepository) ListMedicines(ctx context.Context) ([]domain.Medicine, error) {
	const query = "SELECT id, name, description, stock, unit, price, expired_date FROM medicines ORDER BY name"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Medicine{}
	for rows.Next() {
		var m domain.Medicine
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Stock, &m.Unit, &m.Price, &m.ExpiredDate); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *SQLiteMedicalRepository) UpdateMedicineStock(ctx context.Context, id string, delta int) error {
	const query = "UPDATE medicines SET stock = stock + ? WHERE id = ?"
	_, err := s.db.ExecContext(ctx, query, delta, id)
	if err != nil {
		return ErrMedicalRecordNotFound
	}
	return nil
}

func (s *SQLiteMedicalRepository) UpdateMedicine(ctx context.Context, med domain.Medicine) error {
	const query = "UPDATE medicines SET name = ?, description = ?, stock = ?, unit = ?, price = ?, expired_date = ? WHERE id = ?"
	res, err := s.db.ExecContext(ctx, query, med.Name, med.Description, med.Stock, med.Unit, med.Price, med.ExpiredDate, med.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrMedicineNotFound
	}
	return nil
}

func (s *SQLiteMedicalRepository) DeleteMedicine(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM medicines WHERE id = ?", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrMedicineNotFound
	}
	return nil
}

func (s *SQLiteMedicalRepository) CreateMedicineTransaction(ctx context.Context, t domain.MedicineTransaction) error {
	const query = "INSERT INTO medicine_transactions (id, medicine_id, quantity, type, note, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, query, t.ID, t.MedicineID, t.Quantity, t.Type, t.Note, now)
	return err
}

func (s *SQLiteMedicalRepository) ListMedicineTransactions(ctx context.Context) ([]domain.MedicineTransaction, error) {
	const query = "SELECT id, medicine_id, quantity, type, note, created_at FROM medicine_transactions ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MedicineTransaction{}
	for rows.Next() {
		var t domain.MedicineTransaction
		if err := rows.Scan(&t.ID, &t.MedicineID, &t.Quantity, &t.Type, &t.Note, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *SQLiteMedicalRepository) CreatePrescription(ctx context.Context, p domain.Prescription) error {
	const query = "INSERT INTO prescriptions (id, encounter_id, medicine_id, title, description, quantity, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, query, p.ID, p.EncounterID, p.MedicineID, p.Title, p.Description, p.Quantity, p.Status, now)
	if err != nil {
		return ErrPrescriptionNotFound
	}
	return nil
}

func (s *SQLiteMedicalRepository) GetPrescriptionsByEncounter(ctx context.Context, encounterID string) ([]domain.Prescription, error) {
	const query = "SELECT id, medicine_id, title, description, quantity, status, created_at FROM prescriptions WHERE encounter_id = ? ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Prescription{}
	for rows.Next() {
		var p domain.Prescription
		if err := rows.Scan(&p.ID, &p.MedicineID, &p.Title, &p.Description, &p.Quantity, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *SQLiteMedicalRepository) ListPrescriptions(ctx context.Context) ([]domain.Prescription, error) {
	const query = "SELECT id, encounter_id, medicine_id, title, description, quantity, status, created_at FROM prescriptions ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Prescription{}
	for rows.Next() {
		var p domain.Prescription
		if err := rows.Scan(&p.ID, &p.EncounterID, &p.MedicineID, &p.Title, &p.Description, &p.Quantity, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *SQLiteMedicalRepository) UpdatePrescriptionStatus(ctx context.Context, id, status string) error {
	res, err := s.db.ExecContext(ctx, "UPDATE prescriptions SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrPrescriptionNotFound
	}
	return nil
}

func (s *SQLiteMedicalRepository) Close() error {
	return s.db.Close()
}
