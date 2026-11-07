package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/thinkingx/simrs-backend/internal/domain"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)

type PatientStore interface {
	CreatePatient(ctx context.Context, p domain.Patient) error
	CreateEncounter(ctx context.Context, e domain.Encounter) error
	NextQueueNumber(ctx context.Context, polyclinic string) (int, error)
	ListPatients(ctx context.Context) ([]domain.Patient, error)
	ListEncounters(ctx context.Context, polyclinic string) ([]domain.Encounter, error)
	ListPolyclinics(ctx context.Context) ([]domain.Polyclinic, error)
	ListDoctors(ctx context.Context, poli string) ([]domain.Doctor, error)
	CreateDoctor(ctx context.Context, d domain.Doctor) error
	UpdateDoctor(ctx context.Context, d domain.Doctor) error
	DeleteDoctor(ctx context.Context, id string) error
	GetEncounter(ctx context.Context, id string) (*domain.Encounter, error)
	UpdateEncounterStatus(ctx context.Context, id, status string) error
}

type PatientUsecase struct {
	store PatientStore
}

func NewPatientUsecase(store PatientStore) *PatientUsecase {
	return &PatientUsecase{store: store}
}

func (u *PatientUsecase) RegisterPatient(ctx context.Context, p domain.Patient) error {
	if p.NIK == "" {
		return errors.New("nik required")
	}
	if p.FullName == "" {
		return errors.New("full_name required")
	}
	return u.store.CreatePatient(ctx, p)
}

func (u *PatientUsecase) CreateVisit(ctx context.Context, patientID, polyclinic, doctorID string) (*domain.Encounter, error) {
	if polyclinic == "" {
		return nil, errors.New("polyclinic required")
	}
	q, err := u.store.NextQueueNumber(ctx, polyclinic)
	if err != nil {
		return nil, err
	}
	e := domain.Encounter{
		ID:          "enc-" + strconv.FormatInt(time.Now().UnixNano(), 36),
		PatientID:   patientID,
		DoctorID:    doctorID,
		Polyclinic:  polyclinic,
		QueueNumber: q,
		Status:      "waiting",
		VisitDate:   time.Now().Format(time.RFC3339),
	}
	if err := u.store.CreateEncounter(ctx, e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (u *PatientUsecase) ListPatients(ctx context.Context) ([]domain.Patient, error) {
	return u.store.ListPatients(ctx)
}

func (u *PatientUsecase) ListEncounters(ctx context.Context, polyclinic string) ([]domain.Encounter, error) {
	return u.store.ListEncounters(ctx, polyclinic)
}

func (u *PatientUsecase) ListPolyclinics(ctx context.Context) ([]domain.Polyclinic, error) {
	return u.store.ListPolyclinics(ctx)
}

func (u *PatientUsecase) ListDoctors(ctx context.Context, poli string) ([]domain.Doctor, error) {
	return u.store.ListDoctors(ctx, poli)
}

func (u *PatientUsecase) CreateDoctor(ctx context.Context, d domain.Doctor) error {
	if d.Name == "" {
		return errors.New("name required")
	}
	if d.Poli == "" {
		return errors.New("poli required")
	}
	if d.ID == "" {
		return errors.New("id required")
	}
	return u.store.CreateDoctor(ctx, d)
}

func (u *PatientUsecase) UpdateDoctor(ctx context.Context, d domain.Doctor) error {
	if d.ID == "" {
		return errors.New("id required")
	}
	if d.Name == "" || d.Poli == "" {
		return errors.New("name and poli required")
	}
	return u.store.UpdateDoctor(ctx, d)
}

func (u *PatientUsecase) DeleteDoctor(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id required")
	}
	return u.store.DeleteDoctor(ctx, id)
}

// validEncounterStatus is the allowed lifecycle per SDD §3.3.
var validEncounterStatus = map[string]bool{
	"waiting": true, "in_examination": true, "completed": true, "cancelled": true,
}

func (u *PatientUsecase) GetEncounter(ctx context.Context, id string) (*domain.Encounter, error) {
	return u.store.GetEncounter(ctx, id)
}

func (u *PatientUsecase) UpdateEncounterStatus(ctx context.Context, id, status string) error {
	if !validEncounterStatus[status] {
		return errors.New("invalid status")
	}
	return u.store.UpdateEncounterStatus(ctx, id, status)
}
