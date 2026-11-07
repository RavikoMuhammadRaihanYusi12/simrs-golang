package domain

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleDoctor   Role = "doctor"
	RoleNurse    Role = "nurse"
	RoleCashier  Role = "cashier"
	RolePharmacy Role = "pharmacy"
	RolePatient  Role = "patient"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     Role   `json:"role"`
}

type Patient struct {
	ID          string `json:"id"`
	NIK         string `json:"nik"`
	NoRM        string `json:"no_rm"`
	FullName    string `json:"full_name"`
	BirthDate   string `json:"birth_date"`
	Gender      string `json:"gender"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	Complaint   string `json:"complaint"`   // keluhan utama saat datang
	Description string `json:"description"` // deskripsi tambahan keluhan
}

type Encounter struct {
	ID          string `json:"id"`
	PatientID   string `json:"patient_id"`
	DoctorID    string `json:"doctor_id"`
	Polyclinic  string `json:"polyclinic"`
	QueueNumber int    `json:"queue_number"`
	Status      string `json:"status"`
	VisitDate   string `json:"visit_date"`
}

// Polyclinic is a hospital outpatient department (master data).
type Polyclinic struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Doctor is a clinician assignable to encounters (master data).
type Doctor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Poli string `json:"poli"`
}

type QueueEvent struct {
	Action      string `json:"action"`
	QueueNumber int    `json:"queue_number"`
	Polyclinic  string `json:"polyclinic"`
	Status      string `json:"status"`
}
