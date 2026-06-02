package repositories

import (
	"github.com/lib/pq"
	"studi-ku-backend/internal/models"
)

func (r *Repository) ImportLecturers(items []models.Lecturer) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, x := range items {
		password := x.Password
		defaultPassword := x.DefaultPassword
		if defaultPassword == "" {
			defaultPassword = generateDefaultPassword()
		}
		if password == "" {
			password = defaultPassword
		}
		hashedPassword, err := ensurePasswordHash(password)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			`INSERT INTO lecturers (name,email,password,default_password,nidn,courses,is_password_changed)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			x.Name, x.Email, hashedPassword, defaultPassword, x.NIDN, pq.Array(x.Courses), false,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) ImportAssistants(items []models.LabAssistant) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, x := range items {
		password := x.Password
		defaultPassword := x.DefaultPassword
		if defaultPassword == "" {
			defaultPassword = generateDefaultPassword()
		}
		if password == "" {
			password = defaultPassword
		}
		if x.Role == "" {
			x.Role = "aslab"
		}
		if x.StudentID == "" {
			x.StudentID = "ASLAB-" + x.Email
		}
		if x.Lab == "" {
			x.Lab = "Umum"
		}
		if x.Semester < 1 {
			x.Semester = 1
		}
		hashedPassword, err := ensurePasswordHash(password)
		if err != nil {
			return err
		}
		_, err = tx.Exec(
			`INSERT INTO lab_assistants (name,email,phone,student_id,role,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date,password,default_password,is_password_changed)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,COALESCE(NULLIF($12,''),'Aktif'),CURRENT_DATE,$13,$14,$15)`,
			x.Name, x.Email, x.Phone, x.StudentID, x.Role, x.Lab, x.Supervisor, x.Semester, x.GPA, x.AssignedCourses, x.WeeklyHours, x.Status, hashedPassword, defaultPassword, false,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
