package repositories

import (
	"github.com/lib/pq"
	"studi-ku-backend/internal/models"
)

func (r *Repository) ImportStudents(students []models.Student) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, s := range students {
		password := s.Password
		if password == "" {
			password = "password"
		}
		defaultPassword := s.DefaultPassword
		if defaultPassword == "" {
			defaultPassword = "password"
		}

		_, err := tx.Exec(
			`INSERT INTO students (name,email,password,default_password,student_id,program,semester,courses,status,join_date,is_password_changed)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE(NULLIF($9,''),'Aktif'),CURRENT_DATE,$10)`,
			s.Name,
			s.Email,
			password,
			defaultPassword,
			s.StudentID,
			s.Program,
			s.Semester,
			pq.Array(s.Courses),
			s.Status,
			false,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
