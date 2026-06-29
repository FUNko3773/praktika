package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// ---------- МОДЕЛИ (структуры данных) ----------

// Врач (сотрудник больницы)
type Doctor struct {
	ID            int       `json:"id" db:"id"`
	FullName      string    `json:"full_name" db:"full_name"`
	Phone         string    `json:"phone" db:"phone"`
	Address       string    `json:"address" db:"address"`
	GraduationYear int      `json:"graduation_year" db:"graduation_year"`
	Position      string    `json:"position" db:"position"` // из таблицы Квалификация
}

// Пациент
type Patient struct {
	ID               int     `json:"id" db:"id"`
	FullName         string  `json:"full_name" db:"full_name"`
	Address          string  `json:"address" db:"address"`
	PensionerDiscount *float64 `json:"pensioner_discount" db:"pensioner_discount"`
}

// Медицинская услуга
type MedicalService struct {
	ID          int    `json:"id" db:"id"`
	ServiceName string `json:"service_name" db:"service_name"`
	Cost        int    `json:"cost" db:"cost"`
}

// Обращение
type Appeal struct {
	ID          int       `json:"id" db:"id"`
	PatientID   int       `json:"patient_id" db:"patient_id"`
	Diagnosis   string    `json:"diagnosis" db:"diagnosis"`
	AppealDate  time.Time `json:"appeal_date" db:"appeal_date"`
}

// Приём
type Appointment struct {
	ID             int       `json:"id" db:"id"`
	AppealID       int       `json:"appeal_id" db:"appeal_id"`
	DoctorID       int       `json:"doctor_id" db:"doctor_id"`
	Diagnosis      string    `json:"diagnosis" db:"diagnosis"`
	AppointmentDate time.Time `json:"appointment_date" db:"appointment_date"`
}

// Связь приёма с услугами
type AppointmentService struct {
	AppointmentID int `json:"appointment_id" db:"appointment_id"`
	ServiceID     int `json:"service_id" db:"service_id"`
}

// ---------- РЕПОЗИТОРИЙ (работа с БД) ----------

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// ----- ВРАЧИ -----
func (r *Repository) GetDoctors() ([]Doctor, error) {
	query := `
		SELECT d.*, q.position 
		FROM Doctors d 
		LEFT JOIN Qualifications q ON d.id = q.doctor_id
	`
	var doctors []Doctor
	err := r.db.Select(&doctors, query)
	return doctors, err
}

func (r *Repository) GetDoctorByID(id int) (*Doctor, error) {
	query := `
		SELECT d.*, q.position 
		FROM Doctors d 
		LEFT JOIN Qualifications q ON d.id = q.doctor_id 
		WHERE d.id = ?
	`
	var doctor Doctor
	err := r.db.Get(&doctor, query, id)
	if err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (r *Repository) CreateDoctor(doctor *Doctor) error {
	query := `
		INSERT INTO Doctors (full_name, phone, address, graduation_year) 
		VALUES (?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, doctor.FullName, doctor.Phone, doctor.Address, doctor.GraduationYear)
	if err != nil {
		return err
	}
	
	// Получаем ID созданного врача
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	doctor.ID = int(id)
	
	// Если указана должность, добавляем в таблицу квалификации
	if doctor.Position != "" {
		queryQual := "INSERT INTO Qualifications (doctor_id, position) VALUES (?, ?)"
		_, err = r.db.Exec(queryQual, doctor.ID, doctor.Position)
	}
	return err
}

func (r *Repository) UpdateDoctor(doctor *Doctor) error {
	query := `
		UPDATE Doctors 
		SET full_name = ?, phone = ?, address = ?, graduation_year = ? 
		WHERE id = ?
	`
	_, err := r.db.Exec(query, doctor.FullName, doctor.Phone, doctor.Address, doctor.GraduationYear, doctor.ID)
	if err != nil {
		return err
	}
	
	// Обновляем квалификацию
	if doctor.Position != "" {
		queryQual := `
			INSERT INTO Qualifications (doctor_id, position) 
			VALUES (?, ?) 
			ON DUPLICATE KEY UPDATE position = ?
		`
		_, err = r.db.Exec(queryQual, doctor.ID, doctor.Position, doctor.Position)
	}
	return err
}

func (r *Repository) DeleteDoctor(id int) error {
	// Сначала удаляем квалификацию (из-за ON DELETE CASCADE это не обязательно)
	_, err := r.db.Exec("DELETE FROM Qualifications WHERE doctor_id = ?", id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM Doctors WHERE id = ?", id)
	return err
}

// ----- ПАЦИЕНТЫ -----
func (r *Repository) GetPatients() ([]Patient, error) {
	var patients []Patient
	err := r.db.Select(&patients, "SELECT * FROM Patients")
	return patients, err
}

func (r *Repository) GetPatientByID(id int) (*Patient, error) {
	var patient Patient
	err := r.db.Get(&patient, "SELECT * FROM Patients WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *Repository) CreatePatient(patient *Patient) error {
	query := "INSERT INTO Patients (full_name, address, pensioner_discount) VALUES (?, ?, ?)"
	result, err := r.db.Exec(query, patient.FullName, patient.Address, patient.PensionerDiscount)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	patient.ID = int(id)
	return nil
}

func (r *Repository) UpdatePatient(patient *Patient) error {
	query := "UPDATE Patients SET full_name = ?, address = ?, pensioner_discount = ? WHERE id = ?"
	_, err := r.db.Exec(query, patient.FullName, patient.Address, patient.PensionerDiscount, patient.ID)
	return err
}

func (r *Repository) DeletePatient(id int) error {
	_, err := r.db.Exec("DELETE FROM Patients WHERE id = ?", id)
	return err
}

// ----- МЕДИЦИНСКИЕ УСЛУГИ -----
func (r *Repository) GetMedicalServices() ([]MedicalService, error) {
	var services []MedicalService
	err := r.db.Select(&services, "SELECT * FROM MedicalServices")
	return services, err
}

func (r *Repository) GetMedicalServiceByID(id int) (*MedicalService, error) {
	var service MedicalService
	err := r.db.Get(&service, "SELECT * FROM MedicalServices WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *Repository) CreateMedicalService(service *MedicalService) error {
	query := "INSERT INTO MedicalServices (service_name, cost) VALUES (?, ?)"
	result, err := r.db.Exec(query, service.ServiceName, service.Cost)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	service.ID = int(id)
	return nil
}

func (r *Repository) UpdateMedicalService(service *MedicalService) error {
	query := "UPDATE MedicalServices SET service_name = ?, cost = ? WHERE id = ?"
	_, err := r.db.Exec(query, service.ServiceName, service.Cost, service.ID)
	return err
}

func (r *Repository) DeleteMedicalService(id int) error {
	_, err := r.db.Exec("DELETE FROM MedicalServices WHERE id = ?", id)
	return err
}

// ----- ОБРАЩЕНИЯ -----
func (r *Repository) GetAppeals() ([]Appeal, error) {
	var appeals []Appeal
	err := r.db.Select(&appeals, "SELECT * FROM Appeals")
	return appeals, err
}

func (r *Repository) GetAppealByID(id int) (*Appeal, error) {
	var appeal Appeal
	err := r.db.Get(&appeal, "SELECT * FROM Appeals WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &appeal, nil
}

func (r *Repository) GetAppealsByPatient(patientID int) ([]Appeal, error) {
	var appeals []Appeal
	err := r.db.Select(&appeals, "SELECT * FROM Appeals WHERE patient_id = ?", patientID)
	return appeals, err
}

func (r *Repository) CreateAppeal(appeal *Appeal) error {
	query := "INSERT INTO Appeals (patient_id, diagnosis, appeal_date) VALUES (?, ?, ?)"
	result, err := r.db.Exec(query, appeal.PatientID, appeal.Diagnosis, appeal.AppealDate)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	appeal.ID = int(id)
	return nil
}

func (r *Repository) UpdateAppeal(appeal *Appeal) error {
	query := "UPDATE Appeals SET patient_id = ?, diagnosis = ?, appeal_date = ? WHERE id = ?"
	_, err := r.db.Exec(query, appeal.PatientID, appeal.Diagnosis, appeal.AppealDate, appeal.ID)
	return err
}

func (r *Repository) DeleteAppeal(id int) error {
	_, err := r.db.Exec("DELETE FROM Appeals WHERE id = ?", id)
	return err
}

// ----- ПРИЁМЫ -----
func (r *Repository) GetAppointments() ([]Appointment, error) {
	var apps []Appointment
	err := r.db.Select(&apps, "SELECT * FROM Appointments")
	return apps, err
}

func (r *Repository) GetAppointmentByID(id int) (*Appointment, error) {
	var app Appointment
	err := r.db.Get(&app, "SELECT * FROM Appointments WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *Repository) GetAppointmentsByAppeal(appealID int) ([]Appointment, error) {
	var apps []Appointment
	err := r.db.Select(&apps, "SELECT * FROM Appointments WHERE appeal_id = ?", appealID)
	return apps, err
}

func (r *Repository) CreateAppointment(app *Appointment) error {
	query := "INSERT INTO Appointments (appeal_id, doctor_id, diagnosis, appointment_date) VALUES (?, ?, ?, ?)"
	result, err := r.db.Exec(query, app.AppealID, app.DoctorID, app.Diagnosis, app.AppointmentDate)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	app.ID = int(id)
	return nil
}

func (r *Repository) UpdateAppointment(app *Appointment) error {
	query := "UPDATE Appointments SET appeal_id = ?, doctor_id = ?, diagnosis = ?, appointment_date = ? WHERE id = ?"
	_, err := r.db.Exec(query, app.AppealID, app.DoctorID, app.Diagnosis, app.AppointmentDate, app.ID)
	return err
}

func (r *Repository) DeleteAppointment(id int) error {
	// Сначала удаляем связи с услугами
	_, err := r.db.Exec("DELETE FROM AppointmentServices WHERE appointment_id = ?", id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM Appointments WHERE id = ?", id)
	return err
}

// ----- СВЯЗИ ПРИЁМОВ С УСЛУГАМИ -----
func (r *Repository) AddServiceToAppointment(appointmentID, serviceID int) error {
	query := "INSERT INTO AppointmentServices (appointment_id, service_id) VALUES (?, ?)"
	_, err := r.db.Exec(query, appointmentID, serviceID)
	return err
}

func (r *Repository) RemoveServiceFromAppointment(appointmentID, serviceID int) error {
	query := "DELETE FROM AppointmentServices WHERE appointment_id = ? AND service_id = ?"
	_, err := r.db.Exec(query, appointmentID, serviceID)
	return err
}

func (r *Repository) GetAppointmentServices(appointmentID int) ([]MedicalService, error) {
	query := `
		SELECT ms.* 
		FROM MedicalServices ms
		JOIN AppointmentServices aps ON ms.id = aps.service_id
		WHERE aps.appointment_id = ?
	`
	var services []MedicalService
	err := r.db.Select(&services, query, appointmentID)
	return services, err
}

// Получить все услуги для приёма с деталями
func (r *Repository) GetAppointmentWithServices(appointmentID int) (*Appointment, []MedicalService, error) {
	app, err := r.GetAppointmentByID(appointmentID)
	if err != nil {
		return nil, nil, err
	}
	services, err := r.GetAppointmentServices(appointmentID)
	if err != nil {
		return app, nil, err
	}
	return app, services, nil
}

// ---------- ОБРАБОТЧИКИ (HTTP Handlers) ----------

// ----- ВРАЧИ -----
func GetDoctors(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doctors, err := repo.GetDoctors()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
			return
		}
		if len(doctors) == 0 {
			w.Write([]byte("Врачей нет"))
			return
		}
		for _, d := range doctors {
			w.Write([]byte(fmt.Sprintf("ID: %d, ФИО: %s, Должность: %s, Телефон: %s\n",
				d.ID, d.FullName, d.Position, d.Phone)))
		}
	}
}

func GetDoctorByID(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		doctor, err := repo.GetDoctorByID(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Врач с ID %d не найден", id), http.StatusNotFound)
			return
		}
		w.Write([]byte(fmt.Sprintf("Врач: %s, Должность: %s, Телефон: %s, Адрес: %s, Год окончания: %d",
			doctor.FullName, doctor.Position, doctor.Phone, doctor.Address, doctor.GraduationYear)))
	}
}

func CreateDoctor(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var doctor Doctor
		err := json.NewDecoder(r.Body).Decode(&doctor)
		if err != nil {
			http.Error(w, fmt.Sprintf("Некорректный JSON: %v", err), http.StatusBadRequest)
			return
		}
		err = repo.CreateDoctor(&doctor)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания врача: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(doctor)
	}
}

func UpdateDoctor(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		var doctor Doctor
		err = json.NewDecoder(r.Body).Decode(&doctor)
		if err != nil {
			http.Error(w, fmt.Sprintf("Некорректный JSON: %v", err), http.StatusBadRequest)
			return
		}
		doctor.ID = id
		err = repo.UpdateDoctor(&doctor)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка обновления: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Врач обновлён!"))
	}
}

func DeleteDoctor(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		err = repo.DeleteDoctor(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка удаления: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Врач удалён!"))
	}
}

// ----- ПАЦИЕНТЫ -----
func GetPatients(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patients, err := repo.GetPatients()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
			return
		}
		if len(patients) == 0 {
			w.Write([]byte("Пациентов нет"))
			return
		}
		for _, p := range patients {
			discount := 0.0
			if p.PensionerDiscount != nil {
				discount = *p.PensionerDiscount
			}
			w.Write([]byte(fmt.Sprintf("ID: %d, ФИО: %s, Адрес: %s, Скидка: %.0f%%\n",
				p.ID, p.FullName, p.Address, discount*100)))
		}
	}
}

func GetPatientByID(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		patient, err := repo.GetPatientByID(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Пациент с ID %d не найден", id), http.StatusNotFound)
			return
		}
		discount := 0.0
		if patient.PensionerDiscount != nil {
			discount = *patient.PensionerDiscount
		}
		w.Write([]byte(fmt.Sprintf("Пациент: %s, Адрес: %s, Скидка: %.0f%%",
			patient.FullName, patient.Address, discount*100)))
	}
}

func CreatePatient(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var patient Patient
		err := json.NewDecoder(r.Body).Decode(&patient)
		if err != nil {
			http.Error(w, fmt.Sprintf("Некорректный JSON: %v", err), http.StatusBadRequest)
			return
		}
		err = repo.CreatePatient(&patient)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания пациента: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(patient)
	}
}

func DeletePatient(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		err = repo.DeletePatient(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка удаления: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Пациент удалён!"))
	}
}

// ----- МЕДИЦИНСКИЕ УСЛУГИ -----
func GetMedicalServices(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		services, err := repo.GetMedicalServices()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
			return
		}
		if len(services) == 0 {
			w.Write([]byte("Услуг нет"))
			return
		}
		for _, s := range services {
			w.Write([]byte(fmt.Sprintf("ID: %d, Услуга: %s, Стоимость: %d руб.\n", s.ID, s.ServiceName, s.Cost)))
		}
	}
}

func GetMedicalServiceByID(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		service, err := repo.GetMedicalServiceByID(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Услуга с ID %d не найдена", id), http.StatusNotFound)
			return
		}
		w.Write([]byte(fmt.Sprintf("Услуга: %s, Стоимость: %d руб.", service.ServiceName, service.Cost)))
	}
}

func CreateMedicalService(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var service MedicalService
		err := json.NewDecoder(r.Body).Decode(&service)
		if err != nil {
			http.Error(w, fmt.Sprintf("Некорректный JSON: %v", err), http.StatusBadRequest)
			return
		}
		err = repo.CreateMedicalService(&service)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания услуги: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(service)
	}
}

func DeleteMedicalService(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		err = repo.DeleteMedicalService(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка удаления: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Услуга удалена!"))
	}
}

// ----- ОБРАЩЕНИЯ -----
func GetAppeals(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appeals, err := repo.GetAppeals()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
			return
		}
		if len(appeals) == 0 {
			w.Write([]byte("Обращений нет"))
			return
		}
		for _, a := range appeals {
			patient, _ := repo.GetPatientByID(a.PatientID)
			patientName := "неизвестно"
			if patient != nil {
				patientName = patient.FullName
			}
			w.Write([]byte(fmt.Sprintf("ID: %d, Пациент: %s, Диагноз: %s, Дата: %s\n",
				a.ID, patientName, a.Diagnosis, a.AppealDate.Format("2006-01-02"))))
		}
	}
}

func GetAppealByID(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		appeal, err := repo.GetAppealByID(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Обращение с ID %d не найдено", id), http.StatusNotFound)
			return
		}
		patient, _ := repo.GetPatientByID(appeal.PatientID)
		patientName := "неизвестно"
		if patient != nil {
			patientName = patient.FullName
		}
		w.Write([]byte(fmt.Sprintf("Обращение: %d\nПациент: %s\nДиагноз: %s\nДата: %s",
			appeal.ID, patientName, appeal.Diagnosis, appeal.AppealDate.Format("2006-01-02"))))
	}
}

func CreateAppeal(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var appeal Appeal
		err := json.NewDecoder(r.Body).Decode(&appeal)
		if err != nil {
			http.Error(w, fmt.Sprintf("Некорректный JSON: %v", err), http.StatusBadRequest)
			return
		}
		err = repo.CreateAppeal(&appeal)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания обращения: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(appeal)
	}
}

func DeleteAppeal(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		err = repo.DeleteAppeal(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка удаления: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Обращение удалено!"))
	}
}

// ----- ПРИЁМЫ -----
func GetAppointments(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apps, err := repo.GetAppointments()
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
			return
		}
		if len(apps) == 0 {
			w.Write([]byte("Приёмов нет"))
			return
		}
		for _, a := range apps {
			appeal, _ := repo.GetAppealByID(a.AppealID)
			doctor, _ := repo.GetDoctorByID(a.DoctorID)
			appealInfo := "неизвестно"
			if appeal != nil {
				appealInfo = fmt.Sprintf("Обращение #%d", appeal.ID)
			}
			doctorName := "неизвестно"
			if doctor != nil {
				doctorName = doctor.FullName
			}
			services, _ := repo.GetAppointmentServices(a.ID)
			serviceNames := ""
			for i, s := range services {
				if i > 0 {
					serviceNames += ", "
				}
				serviceNames += s.ServiceName
			}
			w.Write([]byte(fmt.Sprintf("ID: %d, Дата: %s, Диагноз: %s, %s, Врач: %s, Услуги: [%s]\n",
				a.ID, a.AppointmentDate.Format("2006-01-02"), a.Diagnosis, appealInfo, doctorName, serviceNames)))
		}
	}
}

func GetAppointmentByID(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		app, services, err := repo.GetAppointmentWithServices(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Приём с ID %d не найден", id), http.StatusNotFound)
			return
		}
		appeal, _ := repo.GetAppealByID(app.AppealID)
		doctor, _ := repo.GetDoctorByID(app.DoctorID)
		
		appealInfo := "неизвестно"
		if appeal != nil {
			appealInfo = fmt.Sprintf("Обращение #%d", appeal.ID)
		}
		doctorName := "неизвестно"
		if doctor != nil {
			doctorName = doctor.FullName
		}
		
		serviceNames := ""
		for i, s := range services {
			if i > 0 {
				serviceNames += ", "
			}
			serviceNames += fmt.Sprintf("%s (%d руб.)", s.ServiceName, s.Cost)
		}
		
		w.Write([]byte(fmt.Sprintf("Приём ID: %d\nДата: %s\nДиагноз: %s\n%s\nВрач: %s\nУслуги: [%s]",
			app.ID, app.AppointmentDate.Format("2006-01-02"), app.Diagnosis, appealInfo, doctorName, serviceNames)))
	}
}

func CreateAppointment(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			AppealID        int    `json:"appeal_id"`
			DoctorID        int    `json:"doctor_id"`
			Diagnosis       string `json:"diagnosis"`
			AppointmentDate string `json:"appointment_date"`
			ServiceIDs      []int  `json:"service_ids"`
		}

		// Читаем JSON
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, fmt.Sprintf("Некорректный JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Проверяем обязательные поля
		if input.AppealID == 0 {
			http.Error(w, "appeal_id обязательно", http.StatusBadRequest)
			return
		}
		if input.DoctorID == 0 {
			http.Error(w, "doctor_id обязательно", http.StatusBadRequest)
			return
		}
		if input.Diagnosis == "" {
			http.Error(w, "diagnosis обязательно", http.StatusBadRequest)
			return
		}

		// Обрабатываем дату
		var appointmentDate time.Time
		if input.AppointmentDate == "" {
			// Если дата не передана - ставим сегодняшнюю
			appointmentDate = time.Now()
		} else {
			// Пробуем парсить дату в разных форматах
			appointmentDate, err = time.Parse("2006-01-02", input.AppointmentDate)
			if err != nil {
				appointmentDate, err = time.Parse("2006-01-02T15:04:05Z", input.AppointmentDate)
				if err != nil {
					http.Error(w, "Неверный формат даты. Используйте YYYY-MM-DD или YYYY-MM-DDTHH:MM:SSZ", http.StatusBadRequest)
					return
				}
			}
		}

		// Создаём приём
		app := &Appointment{
			AppealID:        input.AppealID,
			DoctorID:        input.DoctorID,
			Diagnosis:       input.Diagnosis,
			AppointmentDate: appointmentDate,
		}

		// Сохраняем в базу
		err = repo.CreateAppointment(app)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка создания приёма: %v", err), http.StatusInternalServerError)
			return
		}

		// Добавляем услуги, если они есть
		for _, serviceID := range input.ServiceIDs {
			err = repo.AddServiceToAppointment(app.ID, serviceID)
			if err != nil {
				log.Printf("Ошибка добавления услуги %d: %v", serviceID, err)
			}
		}

		// Возвращаем результат
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(app)
	}
}

func UpdateAppointment(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		
		var appointment Appointment
		err = json.NewDecoder(r.Body).Decode(&appointment)
		if err != nil {
			http.Error(w, fmt.Sprintf("Некорректный JSON: %v", err), http.StatusBadRequest)
			return
		}
		appointment.ID = id
		
		err = repo.UpdateAppointment(&appointment)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка обновления: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Приём обновлён!"))
	}
}

func DeleteAppointment(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		err = repo.DeleteAppointment(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка удаления: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Приём удалён!"))
	}
}

// Добавление услуги к приёму
func AddServiceToAppointmentHandler(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		appointmentID, err := strconv.Atoi(vars["appointmentId"])
		if err != nil {
			http.Error(w, "Неверный ID приёма", http.StatusBadRequest)
			return
		}
		
		var request struct {
			ServiceID int `json:"service_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, fmt.Sprintf("Некорректный JSON: %v", err), http.StatusBadRequest)
			return
		}
		
		err = repo.AddServiceToAppointment(appointmentID, request.ServiceID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка добавления услуги: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Услуга добавлена к приёму!"))
	}
}

// Удаление услуги из приёма
func RemoveServiceFromAppointmentHandler(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		appointmentID, err := strconv.Atoi(vars["appointmentId"])
		if err != nil {
			http.Error(w, "Неверный ID приёма", http.StatusBadRequest)
			return
		}
		serviceID, err := strconv.Atoi(vars["serviceId"])
		if err != nil {
			http.Error(w, "Неверный ID услуги", http.StatusBadRequest)
			return
		}
		
		err = repo.RemoveServiceFromAppointment(appointmentID, serviceID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка удаления услуги: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Услуга удалена из приёма!"))
	}
}

// Получение услуг для приёма
func GetAppointmentServicesHandler(repo *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		appointmentID, err := strconv.Atoi(vars["appointmentId"])
		if err != nil {
			http.Error(w, "Неверный ID приёма", http.StatusBadRequest)
			return
		}
		
		services, err := repo.GetAppointmentServices(appointmentID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка: %v", err), http.StatusInternalServerError)
			return
		}
		
		if len(services) == 0 {
			w.Write([]byte("Услуг для этого приёма нет"))
			return
		}
		
		for _, s := range services {
			w.Write([]byte(fmt.Sprintf("Услуга: %s, Стоимость: %d руб.\n", s.ServiceName, s.Cost)))
		}
	}
}

// ---------- ГЛАВНАЯ ФУНКЦИЯ ----------

func main() {
	// Подключение к БД (измените на свою)
	dsn := "root:root@tcp(127.0.0.1:3307)/hospital?parseTime=true"
	// Если пароль есть: "root:password@tcp(127.0.0.1:3306)/hospital?parseTime=true"

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("БД не отвечает:", err)
	}
	log.Println("✅ Подключено к MySQL!")

	repo := NewRepository(db)
	r := mux.NewRouter()

	// === ВРАЧИ ===
	r.HandleFunc("/doctors", GetDoctors(repo)).Methods("GET")
	r.HandleFunc("/doctors/{id}", GetDoctorByID(repo)).Methods("GET")
	r.HandleFunc("/doctors", CreateDoctor(repo)).Methods("POST")
	r.HandleFunc("/doctors/{id}", UpdateDoctor(repo)).Methods("PUT")
	r.HandleFunc("/doctors/{id}", DeleteDoctor(repo)).Methods("DELETE")

	// === ПАЦИЕНТЫ ===
	r.HandleFunc("/patients", GetPatients(repo)).Methods("GET")
	r.HandleFunc("/patients/{id}", GetPatientByID(repo)).Methods("GET")
	r.HandleFunc("/patients", CreatePatient(repo)).Methods("POST")
	r.HandleFunc("/patients/{id}", DeletePatient(repo)).Methods("DELETE")

	// === МЕДИЦИНСКИЕ УСЛУГИ ===
	r.HandleFunc("/services", GetMedicalServices(repo)).Methods("GET")
	r.HandleFunc("/services/{id}", GetMedicalServiceByID(repo)).Methods("GET")
	r.HandleFunc("/services", CreateMedicalService(repo)).Methods("POST")
	r.HandleFunc("/services/{id}", DeleteMedicalService(repo)).Methods("DELETE")

	// === ОБРАЩЕНИЯ ===
	r.HandleFunc("/appeals", GetAppeals(repo)).Methods("GET")
	r.HandleFunc("/appeals/{id}", GetAppealByID(repo)).Methods("GET")
	r.HandleFunc("/appeals", CreateAppeal(repo)).Methods("POST")
	r.HandleFunc("/appeals/{id}", DeleteAppeal(repo)).Methods("DELETE")

	// === ПРИЁМЫ ===
	r.HandleFunc("/appointments", GetAppointments(repo)).Methods("GET")
	r.HandleFunc("/appointments/{id}", GetAppointmentByID(repo)).Methods("GET")
	r.HandleFunc("/appointments", CreateAppointment(repo)).Methods("POST")
	r.HandleFunc("/appointments/{id}", UpdateAppointment(repo)).Methods("PUT")
	r.HandleFunc("/appointments/{id}", DeleteAppointment(repo)).Methods("DELETE")

	// === СВЯЗИ ПРИЁМОВ С УСЛУГАМИ ===
	r.HandleFunc("/appointments/{appointmentId}/services", GetAppointmentServicesHandler(repo)).Methods("GET")
	r.HandleFunc("/appointments/{appointmentId}/services", AddServiceToAppointmentHandler(repo)).Methods("POST")
	r.HandleFunc("/appointments/{appointmentId}/services/{serviceId}", RemoveServiceFromAppointmentHandler(repo)).Methods("DELETE")

	log.Println("🚀 Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
} 
