package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/thinkingx/simrs-backend/internal/domain"
)

var (
	ErrInvalidPrescription = errors.New("invalid prescription")
	ErrMedicineNotFound    = errors.New("medicine not found")
)

type MedicalStore interface {
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

type MedicalUsecase struct {
	store MedicalStore
}

func NewMedicalUsecase(store MedicalStore) *MedicalUsecase {
	return &MedicalUsecase{store: store}
}

func (u *MedicalUsecase) CreateMedicalRecord(ctx context.Context, mr domain.MedicalRecord) error {
	if mr.EncounterID == "" {
		return errors.New("encounter_id required")
	}
	if mr.Subjective == "" && mr.Objective == "" && mr.Assessment == "" && mr.Plan == "" {
		return errors.New("at least one SOAP field required")
	}
	if mr.ID == "" {
		mr.ID = "mr-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return u.store.CreateMedicalRecord(ctx, mr)
}

func (u *MedicalUsecase) GetMedicalRecords(ctx context.Context, encounterID string) ([]domain.MedicalRecord, error) {
	if encounterID == "" {
		return nil, errors.New("encounter_id required")
	}
	return u.store.GetMedicalRecordsByEncounter(ctx, encounterID)
}

func (u *MedicalUsecase) CreateMedicine(ctx context.Context, med domain.Medicine) error {
	if med.Name == "" {
		return errors.New("name required")
	}
	if med.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	if med.ID == "" {
		med.ID = "med-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	if err := u.store.CreateMedicine(ctx, med); err != nil {
		return err
	}
	// Log initial stock as an inbound transaction (audit).
	if med.Stock > 0 {
		return u.store.CreateMedicineTransaction(ctx, domain.MedicineTransaction{
			ID:         "trx-" + strconv.FormatInt(time.Now().UnixNano(), 36),
			MedicineID: med.ID,
			Quantity:   med.Stock,
			Type:       "in",
			Note:       "Stok awal",
		})
	}
	return nil
}

func (u *MedicalUsecase) ListMedicines(ctx context.Context) ([]domain.Medicine, error) {
	return u.store.ListMedicines(ctx)
}

func (u *MedicalUsecase) UpdateMedicine(ctx context.Context, med domain.Medicine) error {
	if med.ID == "" || med.Name == "" {
		return errors.New("id and name required")
	}
	if med.Stock < 0 {
		return errors.New("stock cannot be negative")
	}
	return u.store.UpdateMedicine(ctx, med)
}

func (u *MedicalUsecase) DeleteMedicine(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id required")
	}
	return u.store.DeleteMedicine(ctx, id)
}

// TakeMedicine records a stock-out (apoteker mengambil obat) and reduces stock.
func (u *MedicalUsecase) TakeMedicine(ctx context.Context, medicineID string, quantity int, note string) error {
	if medicineID == "" || quantity <= 0 {
		return errors.New("medicine_id and positive quantity required")
	}
	med, err := u.store.GetMedicine(ctx, medicineID)
	if err != nil {
		return err
	}
	if med.Stock < quantity {
		return ErrMedicineNotFound
	}
	if err := u.store.UpdateMedicineStock(ctx, medicineID, -quantity); err != nil {
		return err
	}
	return u.store.CreateMedicineTransaction(ctx, domain.MedicineTransaction{
		ID:         "trx-" + strconv.FormatInt(time.Now().UnixNano(), 36),
		MedicineID: medicineID,
		Quantity:   -quantity,
		Type:       "out",
		Note:       note,
	})
}

func (u *MedicalUsecase) ListMedicineTransactions(ctx context.Context) ([]domain.MedicineTransaction, error) {
	return u.store.ListMedicineTransactions(ctx)
}

func (u *MedicalUsecase) CreatePrescription(ctx context.Context, p domain.Prescription) error {
	if p.EncounterID == "" || p.MedicineID == "" {
		return ErrInvalidPrescription
	}
	if p.Quantity <= 0 {
		return ErrInvalidPrescription
	}
	if p.ID == "" {
		p.ID = "rx-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	if p.Title == "" {
		p.Title = "Resep " + p.ID
	}
	p.Status = "pending"
	return u.store.CreatePrescription(ctx, p)
}

func (u *MedicalUsecase) GetPrescriptions(ctx context.Context, encounterID string) ([]domain.Prescription, error) {
	if encounterID == "" {
		return nil, errors.New("encounter_id required")
	}
	return u.store.GetPrescriptionsByEncounter(ctx, encounterID)
}

// ListPrescriptions powers the pharmacist dispensing queue.
func (u *MedicalUsecase) ListPrescriptions(ctx context.Context) ([]domain.Prescription, error) {
	return u.store.ListPrescriptions(ctx)
}

// SetPrescriptionStatus moves a prescription through pending -> processing -> dispensed.
// On the final dispensed step the stock is deducted and an outbound transaction logged.
func (u *MedicalUsecase) SetPrescriptionStatus(ctx context.Context, id, status string) error {
	if id == "" {
		return ErrInvalidPrescription
	}
	switch status {
	case "processing":
		return u.store.UpdatePrescriptionStatus(ctx, id, "processing")
	case "dispensed":
		// Find the prescription to know medicine + quantity.
		rx, err := u.findPrescription(ctx, id)
		if err != nil {
			return err
		}
		med, err := u.store.GetMedicine(ctx, rx.MedicineID)
		if err != nil {
			return err
		}
		if med.Stock < rx.Quantity {
			return ErrMedicineNotFound
		}
		if err := u.store.UpdateMedicineStock(ctx, rx.MedicineID, -rx.Quantity); err != nil {
			return err
		}
		if err := u.store.CreateMedicineTransaction(ctx, domain.MedicineTransaction{
			ID:         "trx-" + strconv.FormatInt(time.Now().UnixNano(), 36),
			MedicineID: rx.MedicineID,
			Quantity:   -rx.Quantity,
			Type:       "out",
			Note:       "Resep " + rx.ID,
		}); err != nil {
			return err
		}
		return u.store.UpdatePrescriptionStatus(ctx, id, "dispensed")
	default:
		return ErrInvalidPrescription
	}
}

func (u *MedicalUsecase) findPrescription(ctx context.Context, id string) (*domain.Prescription, error) {
	all, err := u.store.ListPrescriptions(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, ErrInvalidPrescription
}
