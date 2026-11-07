package domain

type MedicalRecord struct {
	ID          string `json:"id"`
	EncounterID string `json:"encounter_id"`
	Subjective  string `json:"subjective"`
	Objective   string `json:"objective"`
	Assessment  string `json:"assessment"`
	Plan        string `json:"plan"`
	CreatedAt   string `json:"created_at"`
}

type Medicine struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Stock       int    `json:"stock"`
	Unit        string `json:"unit"`
	Price       int    `json:"price"`
	ExpiredDate string `json:"expired_date"`
}

// MedicineTransaction records a stock movement (in/out) for audit/history.
type MedicineTransaction struct {
	ID         string `json:"id"`
	MedicineID string `json:"medicine_id"`
	Quantity   int    `json:"quantity"` // positive = masuk, negative = keluar
	Type       string `json:"type"`     // "in", "out"
	Note       string `json:"note"`
	CreatedAt  string `json:"created_at"`
}

type Prescription struct {
	ID          string `json:"id"`
	EncounterID string `json:"encounter_id"`
	MedicineID  string `json:"medicine_id"`
	Title       string `json:"title"`       // nama resep
	Description string `json:"description"` // deskripsi obat / aturan pakai
	Quantity    int    `json:"quantity"`
	Status      string `json:"status"` // pending | processing | dispensed
	CreatedAt   string `json:"created_at"`
}

type SATUSEHATPayload struct {
	EncounterID string `json:"encounter_id"`
	PatientID   string `json:"patient_id"`
	Condition   string `json:"condition"`
	Observation string `json:"observation"`
}
