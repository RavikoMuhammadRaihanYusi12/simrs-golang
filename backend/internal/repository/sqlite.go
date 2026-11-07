// Package repository implements SQLite storage using modernc.org/sqlite.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"

	"github.com/thinkingx/simrs-backend/internal/domain"
)

var ErrConstraint = errors.New("constraint violation")

type SQLitePatientStore struct {
	db *sql.DB
}

func NewSQLitePatientStore(dsn string) (*SQLitePatientStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	s := &SQLitePatientStore{db: db}
	if err := s.initSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *SQLitePatientStore) initSchema() error {
	scripts := []string{
		`CREATE TABLE IF NOT EXISTS patients (
			id TEXT PRIMARY KEY,
			nik TEXT UNIQUE NOT NULL,
			no_rm TEXT UNIQUE NOT NULL,
			full_name TEXT NOT NULL,
			birth_date TEXT NOT NULL,
			gender TEXT NOT NULL,
			phone TEXT NOT NULL,
			address TEXT NOT NULL,
			complaint TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS encounters (
			id TEXT PRIMARY KEY,
			patient_id TEXT NOT NULL,
			doctor_id TEXT NOT NULL DEFAULT '',
			polyclinic TEXT NOT NULL,
			queue_number INTEGER NOT NULL,
			status TEXT NOT NULL,
			visit_date TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS polyclinics (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS doctors (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			poli TEXT NOT NULL
		)`,
	}
	for _, script := range scripts {
		if _, err := s.db.Exec(script); err != nil {
			return err
		}
	}
	return s.seedMaster()
}

// seedMaster inserts demo polylinic + doctor rows once.
func (s *SQLitePatientStore) seedMaster() error {
	polys := []domain.Polyclinic{
		{ID: "poli-umum", Name: "Poli Umum"},
		{ID: "poli-anak", Name: "Poli Anak"},
		{ID: "poli-gigi", Name: "Poli Gigi"},
		{ID: "igd", Name: "IGD"},
	}
	for _, p := range polys {
		if _, err := s.db.Exec("INSERT OR IGNORE INTO polyclinics (id, name) VALUES (?, ?)", p.ID, p.Name); err != nil {
			return err
		}
	}
	docs := []domain.Doctor{
		{ID: "doc-1", Name: "dr. Andi Pratama", Poli: "poli-umum"},
		{ID: "doc-2", Name: "dr. Siti Rahayu, Sp.A", Poli: "poli-anak"},
		{ID: "doc-3", Name: "drg. Budi Santoso", Poli: "poli-gigi"},
		{ID: "doc-4", Name: "dr. Maya Lestari, Sp.EM", Poli: "igd"},
	}
	for _, d := range docs {
		if _, err := s.db.Exec("INSERT OR IGNORE INTO doctors (id, name, poli) VALUES (?, ?, ?)", d.ID, d.Name, d.Poli); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLitePatientStore) CreatePatient(ctx context.Context, p domain.Patient) error {
	const query = "INSERT INTO patients (id, nik, no_rm, full_name, birth_date, gender, phone, address, complaint, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, query, p.ID, p.NIK, p.NoRM, p.FullName, p.BirthDate, p.Gender, p.Phone, p.Address, p.Complaint, p.Description, now, now)
	if err != nil {
		return ErrConstraint
	}
	return nil
}

func (s *SQLitePatientStore) CreateEncounter(ctx context.Context, e domain.Encounter) error {
	const query = "INSERT INTO encounters (id, patient_id, doctor_id, polyclinic, queue_number, status, visit_date) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err := s.db.ExecContext(ctx, query, e.ID, e.PatientID, e.DoctorID, e.Polyclinic, e.QueueNumber, e.Status, e.VisitDate)
	if err != nil {
		return ErrConstraint
	}
	return nil
}

func (s *SQLitePatientStore) NextQueueNumber(ctx context.Context, polyclinic string) (int, error) {
	const query = "SELECT COALESCE(MAX(queue_number), 0) FROM encounters WHERE polyclinic = ?"
	var max int
	if err := s.db.QueryRowContext(ctx, query, polyclinic).Scan(&max); err != nil {
		return 0, err
	}
	return max + 1, nil
}

func (s *SQLitePatientStore) ListPatients(ctx context.Context) ([]domain.Patient, error) {
	const query = "SELECT id, nik, no_rm, full_name, birth_date, gender, phone, address, complaint, description FROM patients ORDER BY created_at DESC"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Patient{}
	for rows.Next() {
		var p domain.Patient
		if err := rows.Scan(&p.ID, &p.NIK, &p.NoRM, &p.FullName, &p.BirthDate, &p.Gender, &p.Phone, &p.Address, &p.Complaint, &p.Description); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *SQLitePatientStore) ListEncounters(ctx context.Context, polyclinic string) ([]domain.Encounter, error) {
	query := "SELECT id, patient_id, doctor_id, polyclinic, queue_number, status, visit_date FROM encounters"
	args := []any{}
	if polyclinic != "" {
		query += " WHERE polyclinic = ?"
		args = append(args, polyclinic)
	}
	query += " ORDER BY queue_number"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Encounter{}
	for rows.Next() {
		var e domain.Encounter
		if err := rows.Scan(&e.ID, &e.PatientID, &e.DoctorID, &e.Polyclinic, &e.QueueNumber, &e.Status, &e.VisitDate); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *SQLitePatientStore) GetEncounter(ctx context.Context, id string) (*domain.Encounter, error) {
	const query = "SELECT id, patient_id, doctor_id, polyclinic, queue_number, status, visit_date FROM encounters WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)
	var e domain.Encounter
	if err := row.Scan(&e.ID, &e.PatientID, &e.DoctorID, &e.Polyclinic, &e.QueueNumber, &e.Status, &e.VisitDate); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *SQLitePatientStore) UpdateEncounterStatus(ctx context.Context, id, status string) error {
	res, err := s.db.ExecContext(ctx, "UPDATE encounters SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrConstraint
	}
	return nil
}

func (s *SQLitePatientStore) ListPolyclinics(ctx context.Context) ([]domain.Polyclinic, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name FROM polyclinics ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Polyclinic{}
	for rows.Next() {
		var p domain.Polyclinic
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *SQLitePatientStore) ListDoctors(ctx context.Context, poli string) ([]domain.Doctor, error) {
	query := "SELECT id, name, poli FROM doctors"
	args := []any{}
	if poli != "" {
		query += " WHERE poli = ?"
		args = append(args, poli)
	}
	query += " ORDER BY name"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Doctor{}
	for rows.Next() {
		var d domain.Doctor
		if err := rows.Scan(&d.ID, &d.Name, &d.Poli); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *SQLitePatientStore) CreateDoctor(ctx context.Context, d domain.Doctor) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO doctors (id, name, poli) VALUES (?, ?, ?)", d.ID, d.Name, d.Poli)
	if err != nil {
		return ErrConstraint
	}
	return nil
}

func (s *SQLitePatientStore) UpdateDoctor(ctx context.Context, d domain.Doctor) error {
	res, err := s.db.ExecContext(ctx, "UPDATE doctors SET name = ?, poli = ? WHERE id = ?", d.Name, d.Poli, d.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConstraint
	}
	return nil
}

func (s *SQLitePatientStore) DeleteDoctor(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM doctors WHERE id = ?", id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConstraint
	}
	return nil
}

func (s *SQLitePatientStore) Close() error {
	return s.db.Close()
}
