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
		if password == "" {
			password = "password"
		}
		defaultPassword := x.DefaultPassword
		if defaultPassword == "" {
			defaultPassword = "password"
		}
		_, err := tx.Exec(
			`INSERT INTO lecturers (name,email,password,default_password,nidn,courses,is_password_changed)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			x.Name, x.Email, password, defaultPassword, x.NIDN, pq.Array(x.Courses), false,
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
		if password == "" {
			password = "password"
		}
		defaultPassword := x.DefaultPassword
		if defaultPassword == "" {
			defaultPassword = "password"
		}
		_, err := tx.Exec(
			`INSERT INTO lab_assistants (name,email,phone,student_id,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date,password,default_password,is_password_changed)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,COALESCE(NULLIF($11,''),'Aktif'),CURRENT_DATE,$12,$13,$14)`,
			x.Name, x.Email, x.Phone, x.StudentID, x.Lab, x.Supervisor, x.Semester, x.GPA, x.AssignedCourses, x.WeeklyHours, x.Status, password, defaultPassword, false,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
