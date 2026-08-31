package repositories

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"studi-ku-backend/internal/models"
)

type Repository struct{ db *sql.DB }

func (r *Repository) DB() *sql.DB { return r.db }

func New(db *sql.DB) *Repository {
	r := &Repository{db: db}
	r.autoMigrate()
	return r
}

// autoMigrate applies safe schema additions that may not exist in older deployments.
func (r *Repository) autoMigrate() {
	stmts := []string{
		`ALTER TABLE course_sessions ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0`,
		`UPDATE course_sessions SET sort_order = session_number WHERE sort_order = 0`,
		`ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS rejection_note TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS feedback TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS file_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS file_path VARCHAR(500) NOT NULL DEFAULT ''`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS created_by INT`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMP`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS approved_by INT`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS rejected_by INT`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMP`,
		`ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS rejection_note TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE course_materials ALTER COLUMN material_type TYPE VARCHAR(255)`,
		`ALTER TABLE course_materials ALTER COLUMN size TYPE VARCHAR(255)`,
		`ALTER TABLE course_materials ALTER COLUMN upload_date TYPE VARCHAR(255)`,
		`ALTER TABLE session_assignments ADD COLUMN IF NOT EXISTS file_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE session_assignments ALTER COLUMN due_date TYPE VARCHAR(255)`,
		`UPDATE courses SET students = COALESCE((SELECT cardinality(cls.students) FROM classes cls WHERE cls.code = courses.class_code), students) WHERE class_code != ''`,
		`UPDATE attendance_sessions s SET total_students = COALESCE((SELECT cardinality(cls.students) FROM classes cls WHERE cls.code = s.course_code), s.total_students)`,
		`UPDATE session_assignments sa SET total_students = COALESCE((SELECT cardinality(cls.students) FROM courses c JOIN classes cls ON cls.code = c.class_code WHERE c.id = sa.course_id), sa.total_students)`,
		`UPDATE lab_assistants la SET assigned_courses = (SELECT count(*) FROM courses WHERE assistant = la.name)`,
		`CREATE TABLE IF NOT EXISTS assistant_session_attendance (
			id SERIAL PRIMARY KEY,
			session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
			status VARCHAR(20) NOT NULL DEFAULT 'Hadir',
			check_in_time VARCHAR(20) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(session_id)
		)`,
		`CREATE TABLE IF NOT EXISTS student_submissions (
			id SERIAL PRIMARY KEY,
			assignment_id INT NOT NULL REFERENCES session_assignments(id) ON DELETE CASCADE,
			student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
			answer_text TEXT NOT NULL DEFAULT '',
			file_url TEXT NOT NULL DEFAULT '',
			file_name VARCHAR(255) NOT NULL DEFAULT '',
			file_size VARCHAR(50) NOT NULL DEFAULT '',
			submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			score INT,
			feedback TEXT NOT NULL DEFAULT '',
			UNIQUE(assignment_id, student_id)
		)`,
		`CREATE TABLE IF NOT EXISTS session_assessments (
			id SERIAL PRIMARY KEY,
			session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
			assessment_type VARCHAR(20) NOT NULL CHECK (assessment_type IN ('pretest','posttest')),
			title VARCHAR(255) NOT NULL DEFAULT '',
			score INT,
			max_score INT NOT NULL DEFAULT 100,
			status VARCHAR(50) NOT NULL DEFAULT 'not_started',
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(session_id, assessment_type)
		)`,
		`CREATE TABLE IF NOT EXISTS session_assessment_results (
			id SERIAL PRIMARY KEY,
			session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
			student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
			assessment_type VARCHAR(20) NOT NULL CHECK (assessment_type IN ('pretest','posttest')),
			score INT,
			max_score INT NOT NULL DEFAULT 100,
			status VARCHAR(50) NOT NULL DEFAULT 'not_started',
			note TEXT NOT NULL DEFAULT '',
			submitted_at TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(session_id, student_id, assessment_type)
		)`,
		`CREATE TABLE IF NOT EXISTS session_pretest_questions (
			id SERIAL PRIMARY KEY,
			session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
			assessment_type VARCHAR(20) NOT NULL DEFAULT 'pretest',
			question TEXT NOT NULL,
			options JSONB NOT NULL DEFAULT '[]'::jsonb,
			correct_option INT NOT NULL DEFAULT 0,
			points INT NOT NULL DEFAULT 10,
			explanation TEXT NOT NULL DEFAULT '',
			sort_order INT NOT NULL DEFAULT 0,
			created_by INT REFERENCES lab_assistants(id) ON DELETE SET NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS session_pretest_submissions (
			id SERIAL PRIMARY KEY,
			session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
			student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
			assessment_type VARCHAR(20) NOT NULL DEFAULT 'pretest',
			answers JSONB NOT NULL DEFAULT '[]'::jsonb,
			score INT NOT NULL DEFAULT 0,
			max_score INT NOT NULL DEFAULT 100,
			status VARCHAR(50) NOT NULL DEFAULT 'not_started',
			submitted_at TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		// Kolom assessment_type membuat satu bank soal dipakai pretest dan post-test.
		// ALTER dijalankan terpisah supaya database lama ikut ternormalisasi.
		`ALTER TABLE session_pretest_questions ADD COLUMN IF NOT EXISTS assessment_type VARCHAR(20) NOT NULL DEFAULT 'pretest'`,
		`ALTER TABLE session_pretest_submissions ADD COLUMN IF NOT EXISTS assessment_type VARCHAR(20) NOT NULL DEFAULT 'pretest'`,
		`ALTER TABLE session_pretest_questions DROP CONSTRAINT IF EXISTS session_pretest_questions_assessment_type_check`,
		`ALTER TABLE session_pretest_questions ADD CONSTRAINT session_pretest_questions_assessment_type_check CHECK (assessment_type IN ('pretest','posttest'))`,
		`ALTER TABLE session_pretest_submissions DROP CONSTRAINT IF EXISTS session_pretest_submissions_assessment_type_check`,
		`ALTER TABLE session_pretest_submissions ADD CONSTRAINT session_pretest_submissions_assessment_type_check CHECK (assessment_type IN ('pretest','posttest'))`,
		`ALTER TABLE session_pretest_submissions DROP CONSTRAINT IF EXISTS session_pretest_submissions_session_id_student_id_key`,
		`CREATE UNIQUE INDEX IF NOT EXISTS session_pretest_submissions_unique_idx ON session_pretest_submissions (session_id, student_id, assessment_type)`,
		`CREATE INDEX IF NOT EXISTS session_pretest_questions_lookup_idx ON session_pretest_questions (session_id, assessment_type, sort_order)`,
		`CREATE TABLE IF NOT EXISTS institution_settings (
			id SERIAL PRIMARY KEY,
			university_name VARCHAR(255) NOT NULL,
			faculty_name VARCHAR(255) NOT NULL DEFAULT '',
			study_program_name VARCHAR(255) NOT NULL DEFAULT '',
			laboratory_name VARCHAR(255) NOT NULL DEFAULT '',
			campus_a_address TEXT NOT NULL DEFAULT '',
			campus_b_address TEXT NOT NULL DEFAULT '',
			website VARCHAR(255) NOT NULL DEFAULT '',
			email VARCHAR(255) NOT NULL DEFAULT '',
			phone VARCHAR(255) NOT NULL DEFAULT '',
			logo_path VARCHAR(500) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO institution_settings (id,university_name,faculty_name,study_program_name,laboratory_name,campus_a_address,campus_b_address,website,email,phone,logo_path)
		 VALUES (1,'Universitas Muhammadiyah Jakarta','Fakultas Teknik','Teknik Informatika','Laboratorium Teknik Informatika','JL. K. H. Ahmad Dahlan Cirendeu Ciputat Tangerang Selatan','Jl. Cempaka Putih Tengah XXVII, Jakarta Pusat 10510','umj.ac.id','info@umj.ac.id','+6221-7492862/7401894','')
		 ON CONFLICT (id) DO NOTHING`,
		`CREATE TABLE IF NOT EXISTS course_assessment_weights (
			id SERIAL PRIMARY KEY,
			course_id INT REFERENCES courses(id) ON DELETE CASCADE,
			attendance_weight NUMERIC(5,2) NOT NULL DEFAULT 10,
			pretest_weight NUMERIC(5,2) NOT NULL DEFAULT 15,
			assignment_weight NUMERIC(5,2) NOT NULL DEFAULT 20,
			practicum_weight NUMERIC(5,2) NOT NULL DEFAULT 20,
			posttest_weight NUMERIC(5,2) NOT NULL DEFAULT 35,
			passing_grade NUMERIC(5,2) NOT NULL DEFAULT 55,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(course_id)
		)`,
		`CREATE TABLE IF NOT EXISTS grade_scales (
			id SERIAL PRIMARY KEY,
			min_score NUMERIC(5,2) NOT NULL,
			max_score NUMERIC(5,2) NOT NULL,
			grade VARCHAR(5) NOT NULL,
			is_passed BOOLEAN NOT NULL DEFAULT TRUE
		)`,
		`INSERT INTO grade_scales (id,min_score,max_score,grade,is_passed) VALUES
			(1,85,100,'A',TRUE),(2,80,84.99,'A-',TRUE),(3,75,79.99,'B+',TRUE),(4,70,74.99,'B',TRUE),
			(5,65,69.99,'B-',TRUE),(6,60,64.99,'C+',TRUE),(7,55,59.99,'C',TRUE),(8,45,54.99,'D',FALSE),(9,0,44.99,'E',FALSE)
		 ON CONFLICT (id) DO NOTHING`,
		`CREATE TABLE IF NOT EXISTS report_signers (
			id SERIAL PRIMARY KEY,
			role VARCHAR(80) NOT NULL,
			name VARCHAR(255) NOT NULL,
			identifier_type VARCHAR(50) NOT NULL DEFAULT 'NIDN',
			identifier_number VARCHAR(80) NOT NULL DEFAULT '',
			signature_path VARCHAR(500) NOT NULL DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT TRUE
		)`,
		`CREATE TABLE IF NOT EXISTS student_activity_logs (
			id SERIAL PRIMARY KEY,
			student_id INT REFERENCES students(id) ON DELETE SET NULL,
			course_id INT REFERENCES courses(id) ON DELETE SET NULL,
			session_id INT REFERENCES course_sessions(id) ON DELETE SET NULL,
			activity_type VARCHAR(80) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			ip_address VARCHAR(80) NOT NULL DEFAULT '',
			user_agent TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO session_pretest_questions (session_id,assessment_type,question,options,correct_option,points,explanation,sort_order,created_by)
		 SELECT * FROM (
			VALUES
				((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1 LIMIT 1),'pretest','Apa tujuan pretest dalam praktikum?','["Mengukur pemahaman awal","Menilai akhir pembelajaran","Mengganti presensi","Membuat laporan akhir"]'::jsonb,0,10,'Pretest dipakai untuk mengukur pemahaman awal mahasiswa.',1,1),
				((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1 LIMIT 1),'pretest','Apa keluaran utama dari pretest?','["Nilai awal","Nilai tugas","Nilai presensi","Nilai laporan"]'::jsonb,0,10,'Pretest menghasilkan nilai awal sebelum pembelajaran dimulai.',2,1),
				((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1 LIMIT 1),'pretest','Kapan pretest dikerjakan?','["Sebelum materi","Setelah post-test","Saat upload tugas","Setelah laporan"]'::jsonb,0,10,'Pretest dikerjakan sebelum mahasiswa mempelajari materi sesi.',3,1),
				((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1 LIMIT 1),'posttest','Setelah praktikum, apa fungsi post-test?','["Mengukur pemahaman akhir","Mengukur pemahaman awal","Mengganti presensi","Menentukan jadwal"]'::jsonb,0,10,'Post-test mengukur pemahaman akhir setelah materi dipelajari.',1,1),
				((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1 LIMIT 1),'posttest','Nilai pretest dan post-test dipakai untuk menghitung apa?','["N-Gain","Presensi","Jumlah SKS","Kuota kelas"]'::jsonb,0,10,'Selisih ternormalisasi pretest dan post-test menghasilkan N-Gain.',2,1),
				((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1 LIMIT 1),'posttest','Nilai N-Gain 0,75 termasuk kategori apa?','["Tinggi","Sedang","Rendah","Tidak valid"]'::jsonb,0,10,'Menurut Hake, N-Gain >= 0,70 termasuk kategori tinggi.',3,1)
		 ) AS seed(session_id,assessment_type,question,options,correct_option,points,explanation,sort_order,created_by)
		 WHERE seed.session_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM session_pretest_questions)`,
		`INSERT INTO session_pretest_submissions (session_id,student_id,assessment_type,answers,score,max_score,status,submitted_at)
		 SELECT * FROM (
			VALUES
				((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1 LIMIT 1),1,'pretest','[{"questionId":1,"answerIndex":0},{"questionId":2,"answerIndex":0},{"questionId":3,"answerIndex":0}]'::jsonb,100,100,'completed',CURRENT_TIMESTAMP),
				((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1 LIMIT 1),2,'pretest','[{"questionId":1,"answerIndex":0},{"questionId":2,"answerIndex":1},{"questionId":3,"answerIndex":0}]'::jsonb,66,100,'completed',CURRENT_TIMESTAMP)
		 ) AS seed(session_id,student_id,assessment_type,answers,score,max_score,status,submitted_at)
		 WHERE seed.session_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM session_pretest_submissions)`,
	}
	for _, s := range stmts {
		_, _ = r.db.Exec(s)
	}
}

func (r *Repository) Login(email string, password string) (*models.LoginUser, error) {
	var user models.LoginUser
	var storedPassword string

	err := r.db.QueryRow(`
		SELECT id, name, email, role, password
		FROM (
			SELECT id, name, email, 'student' AS role, password FROM students
			UNION ALL
			SELECT id, name, email, COALESCE(NULLIF(role,''), 'aslab') AS role, password
			FROM lab_assistants
			UNION ALL
			SELECT id, name, email, 'admin' AS role, password FROM admins
		) users
		WHERE lower(email) = lower($1)
		LIMIT 1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &storedPassword)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	if !passwordMatches(storedPassword, password) {
		return nil, sql.ErrNoRows
	}
	if !strings.HasPrefix(storedPassword, "$2a$") && !strings.HasPrefix(storedPassword, "$2b$") && !strings.HasPrefix(storedPassword, "$2y$") {
		if hashed, err := hashPassword(password); err == nil {
			_ = r.setPassword(user.ID, user.Role, hashed, true)
		}
	}

	return &user, nil
}

func passwordMatches(storedPassword, password string) bool {
	if strings.HasPrefix(storedPassword, "$2a$") || strings.HasPrefix(storedPassword, "$2b$") || strings.HasPrefix(storedPassword, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) == nil
	}
	return storedPassword == password
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

func ensurePasswordHash(password string) (string, error) {
	if strings.HasPrefix(password, "$2a$") || strings.HasPrefix(password, "$2b$") || strings.HasPrefix(password, "$2y$") {
		return password, nil
	}
	return hashPassword(password)
}

func generateDefaultPassword() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	out := make([]byte, 10)
	for i := range out {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "StudiKu" + time.Now().Format("150405000")
		}
		out[i] = chars[n.Int64()]
	}
	return string(out)
}

func (r *Repository) UserExists(id int, email, role string) bool {
	table := tableForRole(role)
	if table == "" {
		return false
	}
	var exists bool
	if err := r.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM `+table+` WHERE id=$1 AND lower(email)=lower($2))`, id, email).Scan(&exists); err != nil {
		return false
	}
	return exists
}

func tableForRole(role string) string {
	switch role {
	case "student":
		return "students"
	case "admin":
		return "admins"
	case "aslab", "laboran", "kalab":
		return "lab_assistants"
	}
	return ""
}

func (r *Repository) setPassword(id int, role string, password string, changed bool) error {
	table := tableForRole(role)
	if table == "" {
		return sql.ErrNoRows
	}
	_, err := r.db.Exec(`UPDATE `+table+` SET password=$1,is_password_changed=$2,updated_at=now() WHERE id=$3`, password, changed, id)
	return err
}

func (r *Repository) ChangePassword(id int, role string, currentPassword string, newPassword string) error {
	table := tableForRole(role)
	if table == "" {
		return sql.ErrNoRows
	}
	var storedPassword string
	if err := r.db.QueryRow(`SELECT password FROM `+table+` WHERE id=$1`, id).Scan(&storedPassword); err != nil {
		return err
	}
	if !passwordMatches(storedPassword, currentPassword) {
		return errors.New("password lama tidak sesuai")
	}
	hashed, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	return r.setPassword(id, role, hashed, true)
}

func (r *Repository) Dashboard(studentID int) (map[string]interface{}, error) {
	courses, err := r.StudentCourses(studentID)
	if err != nil {
		return nil, err
	}
	assignments, err := r.Assignments(studentID)
	if err != nil {
		return nil, err
	}

	today := make([]models.DashboardCourse, 0)
	currentDay := indonesianWeekday(time.Now().Weekday())
	for _, c := range courses {
		if len(today) == 2 {
			break
		}
		if strings.Contains(strings.ToLower(c.Schedule.Day), strings.ToLower(currentDay)) {
			today = append(today, models.DashboardCourse{ID: c.ID, Name: c.Name, Time: c.Schedule.StartTime + " - " + c.Schedule.EndTime, Room: c.Room, Instructor: c.Lecturer, Color: c.Color})
		}
	}
	deadlines := make([]models.DashboardDeadline, 0)
	for _, a := range assignments {
		if len(deadlines) == 3 {
			break
		}
		if a.Status == "pending" {
			deadlines = append(deadlines, models.DashboardDeadline{ID: a.ID, Title: a.Title, Course: a.Course, DueDate: a.DueDate, Urgent: false})
		}
	}
	// Jumlah tugas belum dikumpulkan milik mahasiswa ini.
	pendingTaskCount := 0
	for _, a := range assignments {
		if a.Status == "pending" {
			pendingTaskCount++
		}
	}
	return map[string]interface{}{"todayCourses": today, "upcomingDeadlines": deadlines, "pendingTaskCount": pendingTaskCount}, nil
}

func indonesianWeekday(day time.Weekday) string {
	days := map[time.Weekday]string{
		time.Sunday:    "Minggu",
		time.Monday:    "Senin",
		time.Tuesday:   "Selasa",
		time.Wednesday: "Rabu",
		time.Thursday:  "Kamis",
		time.Friday:    "Jumat",
		time.Saturday:  "Sabtu",
	}
	return days[day]
}

func (r *Repository) StudentCourses(studentID int) ([]models.StudentCourse, error) {
	// Hanya kursus dari kelas yang diikuti mahasiswa. studentID 0 (mis. pratinjau)
	// mengembalikan seluruh kursus supaya perilaku lama tidak berubah drastis.
	rows, err := r.db.Query(`
		SELECT c.id, c.name, c.instructor, c.assistant, c.day, c.start_time, c.end_time, c.room, c.attendance_present, c.attendance_total, c.color
		FROM courses c
		WHERE $1 = 0 OR EXISTS (
			SELECT 1 FROM classes cls WHERE cls.code = c.class_code AND $1 = ANY(cls.students)
		)
		ORDER BY c.id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.StudentCourse{}
	for rows.Next() {
		var c models.StudentCourse
		if err := rows.Scan(&c.ID, &c.Name, &c.Lecturer, &c.Tutor, &c.Schedule.Day, &c.Schedule.StartTime, &c.Schedule.EndTime, &c.Room, &c.Attendance.Present, &c.Attendance.Total, &c.Color); err != nil {
			return nil, err
		}
		if c.Attendance.Total > 0 {
			c.Attendance.Percentage = c.Attendance.Present * 100 / c.Attendance.Total
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) Assignments(studentID int) ([]models.Assignment, error) {
	query := `
		SELECT
			sa.id,
			sa.title,
			c.name AS course,
			c.assistant,
			sa.due_date,
			'23:59' AS due_time,
			CASE
				WHEN sub.id IS NULL THEN 'pending'
				WHEN sub.score IS NULL THEN 'submitted'
				ELSE 'graded'
			END AS status,
			COALESCE(c.color, 'bg-blue-500') AS course_color,
			CASE WHEN sub.submitted_at IS NULL THEN NULL ELSE to_char(sub.submitted_at, 'DD Mon YYYY') END AS submitted_date,
			sub.score,
			COALESCE(sub.answer_text, '') AS answer_text,
			COALESCE(sub.file_url, '') AS file_url,
			COALESCE(sub.file_name, '') AS file_name,
			COALESCE(sub.file_size, '') AS file_size
		FROM session_assignments sa
		JOIN courses c ON c.id = sa.course_id
		LEFT JOIN student_submissions sub
			ON sub.assignment_id = sa.id AND ($1 = 0 OR sub.student_id = $1)
		WHERE (
			$1 = 0
			OR EXISTS (
				SELECT 1
				FROM classes cls
				WHERE cls.code = c.class_code
					AND $1 = ANY(cls.students)
			)
		)
		ORDER BY sa.id
	`
	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Assignment{}
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(&a.ID, &a.Title, &a.Course, &a.Assistant, &a.DueDate, &a.DueTime, &a.Status, &a.CourseColor, &a.SubmittedDate, &a.Score, &a.AnswerText, &a.FileURL, &a.FileName, &a.FileSize); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *Repository) Grades() ([]models.Grade, error) {
	rows, err := r.db.Query(`SELECT id,course_name,code,semester,credits,grade,score,color FROM grades ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Grade{}
	for rows.Next() {
		var g models.Grade
		if err := rows.Scan(&g.ID, &g.CourseName, &g.Code, &g.Semester, &g.Credits, &g.Grade, &g.Score, &g.Color); err != nil {
			return nil, err
		}
		items = append(items, g)
	}
	return items, rows.Err()
}

func (r *Repository) SubmitAssignment(id int, submission *models.AssignmentSubmission) error {
	_, err := r.db.Exec(`UPDATE assignments SET status='submitted',submitted_date=to_char(CURRENT_DATE,'DD Mon YYYY'),updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *Repository) AdminCourses() ([]models.AdminCourse, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id,c.name,c.instructor,c.assistant,c.study_program,c.academic_year,c.class_code,c.status,
			c.day,c.start_time,c.end_time,c.room,c.sessions,c.credits,
			COALESCE(cardinality(cls.students), c.students, 0) AS students
		FROM courses c
		LEFT JOIN classes cls ON cls.code = c.class_code
		ORDER BY c.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.AdminCourse{}
	for rows.Next() {
		var c models.AdminCourse
		if err := rows.Scan(&c.ID, &c.Name, &c.Instructor, &c.Assistant, &c.StudyProgram, &c.AcademicYear, &c.ClassCode, &c.Status, &c.Day, &c.StartTime, &c.EndTime, &c.Room, &c.Sessions, &c.Credits, &c.Students); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) CreateCourse(c *models.AdminCourse) error {
	c.Students = r.classStudentCount(c.ClassCode)
	if err := r.db.QueryRow(`INSERT INTO courses (name,instructor,assistant,study_program,academic_year,class_code,status,day,start_time,end_time,room,sessions,credits,students) VALUES ($1,COALESCE(NULLIF($2,''),'-'),$3,$4,$5,$6,COALESCE(NULLIF($7,''),'Aktif'),$8,$9,$10,$11,$12,$13,$14) RETURNING id`, c.Name, c.Instructor, c.Assistant, c.StudyProgram, c.AcademicYear, c.ClassCode, c.Status, c.Day, c.StartTime, c.EndTime, c.Room, c.Sessions, c.Credits, c.Students).Scan(&c.ID); err != nil {
		return err
	}
	if err := r.syncCourseSessions(c.ID); err != nil {
		return err
	}
	return r.syncAssistantCourseCount(c.Assistant)
}
func (r *Repository) UpdateCourse(id int, c *models.AdminCourse) error {
	c.Students = r.classStudentCount(c.ClassCode)
	var oldAssistant string
	_ = r.db.QueryRow(`SELECT assistant FROM courses WHERE id=$1`, id).Scan(&oldAssistant)
	if _, err := r.db.Exec(`UPDATE courses SET name=$1,instructor=COALESCE(NULLIF($2,''),'-'),assistant=$3,study_program=$4,academic_year=$5,class_code=$6,status=$7,day=$8,start_time=$9,end_time=$10,room=$11,sessions=$12,credits=$13,students=$14,updated_at=now() WHERE id=$15`, c.Name, c.Instructor, c.Assistant, c.StudyProgram, c.AcademicYear, c.ClassCode, c.Status, c.Day, c.StartTime, c.EndTime, c.Room, c.Sessions, c.Credits, c.Students, id); err != nil {
		return err
	}
	if err := r.syncCourseSessions(id); err != nil {
		return err
	}
	if oldAssistant != "" && oldAssistant != c.Assistant {
		if err := r.syncAssistantCourseCount(oldAssistant); err != nil {
			return err
		}
	}
	return r.syncAssistantCourseCount(c.Assistant)
}

func (r *Repository) classStudentCount(classCode string) int {
	var count int
	if err := r.db.QueryRow(`SELECT COALESCE(cardinality(students),0) FROM classes WHERE code=$1`, classCode).Scan(&count); err != nil {
		return 0
	}
	return count
}

func (r *Repository) syncAllCourseSessions() error {
	rows, err := r.db.Query(`SELECT id FROM courses`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := r.syncCourseSessions(id); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *Repository) syncCourseSessions(courseID int) error {
	_, err := r.db.Exec(`
		INSERT INTO course_sessions (course_id,session_number,title,topic,session_date,session_time,session_type,conference_link,room,description,sort_order)
		SELECT c.id,n,'Sesi ' || n,c.name,
			to_char(CURRENT_DATE + ((n - 1) * INTERVAL '7 days'), 'YYYY-MM-DD'),
			CASE WHEN c.start_time = '' AND c.end_time = '' THEN c.day ELSE c.start_time || ' - ' || c.end_time END,
			'offline',NULL,c.room,'',n
		FROM courses c
		CROSS JOIN generate_series(1,c.sessions) n
		WHERE c.id=$1
		ON CONFLICT (course_id,session_number) DO UPDATE SET
			session_time=EXCLUDED.session_time,
			room=EXCLUDED.room
	`, courseID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		DELETE FROM course_sessions cs
		USING courses c
		WHERE cs.course_id=c.id
			AND c.id=$1
			AND cs.session_number > c.sessions
	`, courseID)
	return err
}

func (r *Repository) syncCourseStudentCount(classCode string) error {
	_, err := r.db.Exec(`
		UPDATE courses
		SET students = COALESCE((SELECT cardinality(students) FROM classes WHERE code=$1), 0),
			updated_at = now()
		WHERE class_code = $1
	`, classCode)
	return err
}

func (r *Repository) syncAssistantCourseCount(name string) error {
	if strings.TrimSpace(name) == "" {
		return nil
	}
	_, err := r.db.Exec(`
		UPDATE lab_assistants
		SET assigned_courses = (
				SELECT count(*)
				FROM courses
				WHERE assistant = $1
			),
			updated_at = now()
		WHERE name = $1
	`, name)
	return err
}

func (r *Repository) assistantCourseCount(name string) int {
	if strings.TrimSpace(name) == "" {
		return 0
	}
	var count int
	if err := r.db.QueryRow(`SELECT count(*) FROM courses WHERE assistant=$1`, name).Scan(&count); err != nil {
		return 0
	}
	return count
}

func (r *Repository) DeleteCourse(id int) error {
	var oldAssistant string
	_ = r.db.QueryRow(`SELECT assistant FROM courses WHERE id=$1`, id).Scan(&oldAssistant)
	if _, err := r.db.Exec(`DELETE FROM courses WHERE id=$1`, id); err != nil {
		return err
	}
	return r.syncAssistantCourseCount(oldAssistant)
}

func (r *Repository) AcademicYears() ([]models.AcademicYear, error) {
	rows, err := r.db.Query(`
		SELECT
			ay.id,ay.name,ay.start_date::text,ay.end_date::text,ay.semester,ay.status,
			(SELECT count(*) FROM courses c WHERE c.academic_year=ay.name) AS total_courses,
			COALESCE((SELECT sum(cardinality(cls.students))::int FROM classes cls WHERE cls.academic_year=ay.name), 0) AS total_students
		FROM academic_years ay
		ORDER BY ay.start_date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.AcademicYear{}
	for rows.Next() {
		var y models.AcademicYear
		if err := rows.Scan(&y.ID, &y.Name, &y.StartDate, &y.EndDate, &y.Semester, &y.Status, &y.TotalCourses, &y.TotalStudents); err != nil {
			return nil, err
		}
		items = append(items, y)
	}
	return items, rows.Err()
}
func (r *Repository) CreateAcademicYear(y *models.AcademicYear) error {
	return r.db.QueryRow(`INSERT INTO academic_years (name,start_date,end_date,semester,status,total_courses,total_students) VALUES ($1,$2,$3,$4,COALESCE(NULLIF($5,''),'Mendatang'),$6,$7) RETURNING id`, y.Name, y.StartDate, y.EndDate, y.Semester, y.Status, y.TotalCourses, y.TotalStudents).Scan(&y.ID)
}
func (r *Repository) UpdateAcademicYear(id int, y *models.AcademicYear) error {
	_, err := r.db.Exec(`UPDATE academic_years SET name=$1,start_date=$2,end_date=$3,semester=$4,status=$5,total_courses=$6,total_students=$7,updated_at=now() WHERE id=$8`, y.Name, y.StartDate, y.EndDate, y.Semester, y.Status, y.TotalCourses, y.TotalStudents, id)
	return err
}
func (r *Repository) DeleteAcademicYear(id int) error {
	_, err := r.db.Exec(`DELETE FROM academic_years WHERE id=$1`, id)
	return err
}

func (r *Repository) Students() ([]models.Student, error) {
	rows, err := r.db.Query(`SELECT id,name,email,password,default_password,student_id,program,semester,courses,status,join_date::text,is_password_changed FROM students ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Student{}
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.Name, &s.Email, &s.Password, &s.DefaultPassword, &s.StudentID, &s.Program, &s.Semester, pq.Array(&s.Courses), &s.Status, &s.JoinDate, &s.IsPasswordChanged); err != nil {
			return nil, err
		}
		s.Password = ""
		s.DefaultPassword = ""
		items = append(items, s)
	}
	return items, rows.Err()
}
func (r *Repository) CreateStudent(s *models.Student) error {
	if s.DefaultPassword == "" {
		s.DefaultPassword = generateDefaultPassword()
	}
	if s.Password == "" {
		s.Password = s.DefaultPassword
	}
	hashed, err := hashPassword(s.Password)
	if err != nil {
		return err
	}
	s.Password = hashed
	return r.db.QueryRow(`INSERT INTO students (name,email,password,default_password,student_id,program,semester,courses,status,join_date,is_password_changed) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE(NULLIF($9,''),'Aktif'),COALESCE(NULLIF($10,'')::date,CURRENT_DATE),$11) RETURNING id`, s.Name, s.Email, s.Password, s.DefaultPassword, s.StudentID, s.Program, s.Semester, pq.Array(s.Courses), s.Status, s.JoinDate, s.IsPasswordChanged).Scan(&s.ID)
}
func (r *Repository) UpdateStudent(id int, s *models.Student) error {
	if s.Password == "" {
		if err := r.db.QueryRow(`SELECT password FROM students WHERE id=$1`, id).Scan(&s.Password); err != nil {
			return err
		}
	} else {
		hashed, err := ensurePasswordHash(s.Password)
		if err != nil {
			return err
		}
		s.Password = hashed
	}
	_, err := r.db.Exec(`UPDATE students SET name=$1,email=$2,password=$3,default_password=$4,student_id=$5,program=$6,semester=$7,courses=$8,status=$9,join_date=$10,is_password_changed=$11,updated_at=now() WHERE id=$12`, s.Name, s.Email, s.Password, s.DefaultPassword, s.StudentID, s.Program, s.Semester, pq.Array(s.Courses), s.Status, s.JoinDate, s.IsPasswordChanged, id)
	return err
}
func (r *Repository) DeleteStudent(id int) error {
	_, err := r.db.Exec(`DELETE FROM students WHERE id=$1`, id)
	return err
}
func (r *Repository) ResetStudentPassword(id int) (*models.Student, error) {
	var defaultPassword string
	if err := r.db.QueryRow(`SELECT default_password FROM students WHERE id=$1`, id).Scan(&defaultPassword); err != nil {
		return nil, err
	}
	hashed, err := hashPassword(defaultPassword)
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(`UPDATE students SET password=$1,is_password_changed=false,updated_at=now() WHERE id=$2`, hashed, id)
	if err != nil {
		return nil, err
	}
	items, err := r.Students()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *Repository) Lecturers() ([]models.Lecturer, error) {
	rows, err := r.db.Query(`SELECT id,name,email,password,default_password,nidn,courses,is_password_changed FROM lecturers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Lecturer{}
	for rows.Next() {
		var x models.Lecturer
		if err := rows.Scan(&x.ID, &x.Name, &x.Email, &x.Password, &x.DefaultPassword, &x.NIDN, pq.Array(&x.Courses), &x.IsPasswordChanged); err != nil {
			return nil, err
		}
		x.Password = ""
		x.DefaultPassword = ""
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) CreateLecturer(x *models.Lecturer) error {
	if x.DefaultPassword == "" {
		x.DefaultPassword = generateDefaultPassword()
	}
	if x.Password == "" {
		x.Password = x.DefaultPassword
	}
	hashed, err := hashPassword(x.Password)
	if err != nil {
		return err
	}
	x.Password = hashed
	return r.db.QueryRow(`INSERT INTO lecturers (name,email,password,default_password,nidn,courses,is_password_changed) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, x.Name, x.Email, x.Password, x.DefaultPassword, x.NIDN, pq.Array(x.Courses), x.IsPasswordChanged).Scan(&x.ID)
}
func (r *Repository) UpdateLecturer(id int, x *models.Lecturer) error {
	if x.DefaultPassword == "" {
		x.DefaultPassword = generateDefaultPassword()
	}
	if x.Password == "" {
		if err := r.db.QueryRow(`SELECT password FROM lecturers WHERE id=$1`, id).Scan(&x.Password); err != nil {
			return err
		}
	} else {
		hashed, err := ensurePasswordHash(x.Password)
		if err != nil {
			return err
		}
		x.Password = hashed
	}
	_, err := r.db.Exec(`UPDATE lecturers SET name=$1,email=$2,password=$3,default_password=$4,nidn=$5,courses=$6,is_password_changed=$7,updated_at=now() WHERE id=$8`, x.Name, x.Email, x.Password, x.DefaultPassword, x.NIDN, pq.Array(x.Courses), x.IsPasswordChanged, id)
	return err
}
func (r *Repository) DeleteLecturer(id int) error {
	_, err := r.db.Exec(`DELETE FROM lecturers WHERE id=$1`, id)
	return err
}
func (r *Repository) ResetLecturerPassword(id int) (*models.Lecturer, error) {
	var defaultPassword string
	if err := r.db.QueryRow(`SELECT default_password FROM lecturers WHERE id=$1`, id).Scan(&defaultPassword); err != nil {
		return nil, err
	}
	hashed, err := hashPassword(defaultPassword)
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(`UPDATE lecturers SET password=$1,is_password_changed=false,updated_at=now() WHERE id=$2`, hashed, id)
	if err != nil {
		return nil, err
	}
	items, err := r.Lecturers()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *Repository) Assistants() ([]models.LabAssistant, error) {
	rows, err := r.db.Query(`
		SELECT
			la.id,la.name,la.email,la.phone,la.student_id,COALESCE(NULLIF(la.role,''),'aslab'),la.lab,la.supervisor,la.semester,la.gpa,
			(SELECT count(*) FROM courses c WHERE c.assistant = la.name) AS assigned_courses,
			la.weekly_hours,la.status,la.join_date::text,la.password,la.default_password,la.is_password_changed
		FROM lab_assistants la
		ORDER BY la.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.LabAssistant{}
	for rows.Next() {
		var x models.LabAssistant
		if err := rows.Scan(&x.ID, &x.Name, &x.Email, &x.Phone, &x.StudentID, &x.Role, &x.Lab, &x.Supervisor, &x.Semester, &x.GPA, &x.AssignedCourses, &x.WeeklyHours, &x.Status, &x.JoinDate, &x.Password, &x.DefaultPassword, &x.IsPasswordChanged); err != nil {
			return nil, err
		}
		x.Password = ""
		x.DefaultPassword = ""
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) CreateAssistant(x *models.LabAssistant) error {
	if x.DefaultPassword == "" {
		x.DefaultPassword = generateDefaultPassword()
	}
	if x.Password == "" {
		x.Password = x.DefaultPassword
	}
	if x.Role == "" {
		x.Role = "aslab"
	}
	if strings.TrimSpace(x.StudentID) == "" {
		x.StudentID = strings.ToUpper(x.Role) + "-" + strings.ReplaceAll(strings.ToLower(x.Email), "@", "-")
	}
	if strings.TrimSpace(x.Lab) == "" {
		x.Lab = "Umum"
	}
	if x.Semester < 1 {
		x.Semester = 1
	}
	hashed, err := ensurePasswordHash(x.Password)
	if err != nil {
		return err
	}
	x.Password = hashed
	x.AssignedCourses = r.assistantCourseCount(x.Name)
	return r.db.QueryRow(`INSERT INTO lab_assistants (name,email,phone,student_id,role,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date,password,default_password,is_password_changed) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,COALESCE(NULLIF($12,''),'Aktif'),COALESCE(NULLIF($13,'')::date,CURRENT_DATE),$14,$15,$16) RETURNING id`, x.Name, x.Email, x.Phone, x.StudentID, x.Role, x.Lab, x.Supervisor, x.Semester, x.GPA, x.AssignedCourses, x.WeeklyHours, x.Status, x.JoinDate, x.Password, x.DefaultPassword, x.IsPasswordChanged).Scan(&x.ID)
}
func (r *Repository) UpdateAssistant(id int, x *models.LabAssistant) error {
	if x.DefaultPassword == "" {
		x.DefaultPassword = generateDefaultPassword()
	}
	if x.Password == "" {
		if err := r.db.QueryRow(`SELECT password FROM lab_assistants WHERE id=$1`, id).Scan(&x.Password); err != nil {
			return err
		}
	} else {
		hashed, err := ensurePasswordHash(x.Password)
		if err != nil {
			return err
		}
		x.Password = hashed
	}
	if x.Role == "" {
		x.Role = "aslab"
	}
	if strings.TrimSpace(x.StudentID) == "" {
		x.StudentID = strings.ToUpper(x.Role) + "-" + strings.ReplaceAll(strings.ToLower(x.Email), "@", "-")
	}
	if strings.TrimSpace(x.Lab) == "" {
		x.Lab = "Umum"
	}
	if x.Semester < 1 {
		x.Semester = 1
	}
	var oldName string
	_ = r.db.QueryRow(`SELECT name FROM lab_assistants WHERE id=$1`, id).Scan(&oldName)
	x.AssignedCourses = r.assistantCourseCount(x.Name)
	if _, err := r.db.Exec(`UPDATE lab_assistants SET name=$1,email=$2,phone=$3,student_id=$4,role=$5,lab=$6,supervisor=$7,semester=$8,gpa=$9,assigned_courses=$10,weekly_hours=$11,status=$12,join_date=$13,password=$14,default_password=$15,is_password_changed=$16,updated_at=now() WHERE id=$17`, x.Name, x.Email, x.Phone, x.StudentID, x.Role, x.Lab, x.Supervisor, x.Semester, x.GPA, x.AssignedCourses, x.WeeklyHours, x.Status, x.JoinDate, x.Password, x.DefaultPassword, x.IsPasswordChanged, id); err != nil {
		return err
	}
	if oldName != "" && oldName != x.Name {
		if _, err := r.db.Exec(`UPDATE courses SET assistant=$1,updated_at=now() WHERE assistant=$2`, x.Name, oldName); err != nil {
			return err
		}
		if err := r.syncAssistantCourseCount(oldName); err != nil {
			return err
		}
	}
	return r.syncAssistantCourseCount(x.Name)
}
func (r *Repository) DeleteAssistant(id int) error {
	var oldName string
	_ = r.db.QueryRow(`SELECT name FROM lab_assistants WHERE id=$1`, id).Scan(&oldName)
	if _, err := r.db.Exec(`DELETE FROM lab_assistants WHERE id=$1`, id); err != nil {
		return err
	}
	_, err := r.db.Exec(`UPDATE courses SET assistant='',updated_at=now() WHERE assistant=$1`, oldName)
	return err
}
func (r *Repository) ResetAssistantPassword(id int) (*models.LabAssistant, error) {
	var defaultPassword string
	if err := r.db.QueryRow(`SELECT default_password FROM lab_assistants WHERE id=$1`, id).Scan(&defaultPassword); err != nil {
		return nil, err
	}
	hashed, err := hashPassword(defaultPassword)
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(`UPDATE lab_assistants SET password=$1,is_password_changed=false,updated_at=now() WHERE id=$2`, hashed, id)
	if err != nil {
		return nil, err
	}
	items, err := r.Assistants()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *Repository) Classes() ([]models.ClassData, error) {
	rows, err := r.db.Query(`SELECT id,code,name,academic_year,assistant,schedule,room,total_students,capacity,students FROM classes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.ClassData{}
	for rows.Next() {
		var c models.ClassData
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.AcademicYear, &c.Assistant, &c.Schedule, &c.Room, &c.TotalStudents, &c.Capacity, pq.Array(&c.Students)); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}
func (r *Repository) CreateClass(c *models.ClassData) error {
	c.TotalStudents = len(c.Students)
	if err := r.db.QueryRow(`INSERT INTO classes (code,name,academic_year,assistant,schedule,room,total_students,capacity,students) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, c.Code, c.Name, c.AcademicYear, c.Assistant, c.Schedule, c.Room, c.TotalStudents, c.Capacity, pq.Array(c.Students)).Scan(&c.ID); err != nil {
		return err
	}
	return r.syncCourseStudentCount(c.Code)
}
func (r *Repository) UpdateClass(id int, c *models.ClassData) error {
	c.TotalStudents = len(c.Students)
	var oldCode string
	_ = r.db.QueryRow(`SELECT code FROM classes WHERE id=$1`, id).Scan(&oldCode)
	if _, err := r.db.Exec(`UPDATE classes SET code=$1,name=$2,academic_year=$3,assistant=$4,schedule=$5,room=$6,total_students=$7,capacity=$8,students=$9,updated_at=now() WHERE id=$10`, c.Code, c.Name, c.AcademicYear, c.Assistant, c.Schedule, c.Room, c.TotalStudents, c.Capacity, pq.Array(c.Students), id); err != nil {
		return err
	}
	if oldCode != "" && oldCode != c.Code {
		if err := r.syncCourseStudentCount(oldCode); err != nil {
			return err
		}
	}
	return r.syncCourseStudentCount(c.Code)
}
func (r *Repository) DeleteClass(id int) error {
	var oldCode string
	_ = r.db.QueryRow(`SELECT code FROM classes WHERE id=$1`, id).Scan(&oldCode)
	if _, err := r.db.Exec(`DELETE FROM classes WHERE id=$1`, id); err != nil {
		return err
	}
	if oldCode == "" {
		return nil
	}
	return r.syncCourseStudentCount(oldCode)
}

func (r *Repository) CreateLecturerCourse(c *models.LecturerCourse) error {
	if c.SKS == 0 {
		c.SKS = 3
	}
	c.Students = r.classStudentCount(c.Code)
	return r.db.QueryRow(`INSERT INTO courses (name,instructor,assistant,study_program,academic_year,class_code,status,day,start_time,end_time,room,sessions,credits,students,description) VALUES ($1,'','', 'Teknik Informatika',$2,$3,'Aktif',$4,'','',$5,14,$6,$7,$8) RETURNING id`, c.Name, c.AcademicYear, c.Code, c.Schedule, c.Room, c.SKS, c.Students, c.Description).Scan(&c.ID)
}

func (r *Repository) UpdateLecturerCourse(id int, c *models.LecturerCourse) error {
	if c.SKS == 0 {
		c.SKS = 3
	}
	c.Students = r.classStudentCount(c.Code)
	_, err := r.db.Exec(`UPDATE courses SET name=$1,academic_year=$2,class_code=$3,day=$4,room=$5,credits=$6,students=$7,description=$8,updated_at=now() WHERE id=$9`, c.Name, c.AcademicYear, c.Code, c.Schedule, c.Room, c.SKS, c.Students, c.Description, id)
	return err
}

func (r *Repository) DeleteLecturerCourse(id int) error {
	_, err := r.db.Exec(`DELETE FROM courses WHERE id=$1`, id)
	return err
}

func (r *Repository) CreateAdminAssignment(a *models.AdminAssignment) error {
	if a.Type == "" {
		a.Type = "Tugas"
	}
	if a.Status == "" {
		a.Status = "Aktif"
	}
	status := "pending"
	if a.Status == "Selesai" {
		status = "graded"
	}
	return r.db.QueryRow(`INSERT INTO assignments (title,course,assistant,due_date,due_time,status,course_color,instructor,total_students,submitted,graded,pending,assignment_type) VALUES ($1,$2,$3,$4,'23:59',$5,'bg-blue-500',$6,$7,$8,$9,$10,$11) RETURNING id`, a.Title, a.Course, a.Instructor, a.DueDate, status, a.Instructor, a.TotalStudents, a.Submitted, a.Graded, a.Pending, a.Type).Scan(&a.ID)
}

func (r *Repository) UpdateAdminAssignment(id int, a *models.AdminAssignment) error {
	if a.Type == "" {
		a.Type = "Tugas"
	}
	if a.Status == "" {
		a.Status = "Aktif"
	}
	status := "pending"
	if a.Status == "Selesai" {
		status = "graded"
	}
	_, err := r.db.Exec(`UPDATE assignments SET title=$1,course=$2,assistant=$3,due_date=$4,status=$5,instructor=$6,total_students=$7,submitted=$8,graded=$9,pending=$10,assignment_type=$11,updated_at=now() WHERE id=$12`, a.Title, a.Course, a.Instructor, a.DueDate, status, a.Instructor, a.TotalStudents, a.Submitted, a.Graded, a.Pending, a.Type, id)
	return err
}

func (r *Repository) DeleteAdminAssignment(id int) error {
	_, err := r.db.Exec(`DELETE FROM assignments WHERE id=$1`, id)
	return err
}

func (r *Repository) DeleteMaterial(id int) error {
	_, err := r.db.Exec(`UPDATE course_materials SET status='deleted',deleted_at=now() WHERE id=$1`, id)
	return err
}

func (r *Repository) MaterialsForRole(role string, userID int) ([]models.MaterialItem, error) {
	where := `m.deleted_at IS NULL`
	args := []interface{}{}
	if role != "aslab" {
		where += ` AND m.status IN ('submitted','available','approved','rejected')`
	}
	rows, err := r.db.Query(`
		SELECT m.id,m.course_id,m.session_id,m.title,m.description,m.material_type,m.size,m.upload_date,m.status,
			m.file_path,c.name,m.created_by,COALESCE(m.rejection_note,'')
		FROM course_materials m
		JOIN courses c ON c.id=m.course_id
		WHERE `+where+`
		ORDER BY m.id DESC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.MaterialItem{}
	for rows.Next() {
		var item models.MaterialItem
		if err := rows.Scan(&item.ID, &item.CourseID, &item.SessionID, &item.Title, &item.Description, &item.Type, &item.Size, &item.UploadDate, &item.Status, &item.FileURL, &item.CourseName, &item.CreatedBy, &item.RejectionNote); err != nil {
			return nil, err
		}
		item.FileURL = "/api/materials/" + fmt.Sprint(item.ID) + "/file"
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateMaterial(courseID int, sessionID *int, title string, description string, filePath string, size string, createdBy int) (*models.MaterialItem, error) {
	var item models.MaterialItem
	err := r.db.QueryRow(`
		INSERT INTO course_materials (course_id,session_id,title,description,material_type,size,upload_date,status,file_path,created_by)
		VALUES ($1,$2,$3,$4,'PDF',$5,to_char(CURRENT_DATE,'YYYY-MM-DD'),'draft',$6,$7)
		RETURNING id
	`, courseID, sessionID, title, description, size, filePath, createdBy).Scan(&item.ID)
	if err != nil {
		return nil, err
	}
	items, err := r.MaterialsForRole("aslab", createdBy)
	if err != nil {
		return nil, err
	}
	for _, x := range items {
		if x.ID == item.ID {
			return &x, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *Repository) SubmitMaterial(id int, userID int) error {
	result, err := r.db.Exec(`UPDATE course_materials SET status='submitted',submitted_at=now(),rejection_note='',rejected_by=NULL,rejected_at=NULL WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) ApproveMaterial(id int, userID int) error {
	result, err := r.db.Exec(`
		UPDATE course_materials
		SET status='approved',approved_by=$1,approved_at=now(),rejection_note='',rejected_by=NULL,rejected_at=NULL
		WHERE id=$2 AND deleted_at IS NULL AND status IN ('submitted','available','rejected')
	`, userID, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) RejectMaterial(id int, userID int, note string) error {
	if strings.TrimSpace(note) == "" {
		return errors.New("catatan penolakan wajib diisi")
	}
	result, err := r.db.Exec(`
		UPDATE course_materials
		SET status='rejected',rejected_by=$1,rejected_at=now(),rejection_note=$2
		WHERE id=$3 AND deleted_at IS NULL AND status IN ('submitted','available','approved')
	`, userID, note, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) MaterialFilePath(id int, role string) (string, error) {
	query := `SELECT file_path FROM course_materials WHERE id=$1 AND deleted_at IS NULL`
	if role != "aslab" {
		query += ` AND status IN ('submitted','available','approved','rejected')`
	}
	var path string
	if err := r.db.QueryRow(query, id).Scan(&path); err != nil {
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", sql.ErrNoRows
	}
	return path, nil
}

func (r *Repository) syncAssistantAttendanceSessions() error {
	if _, err := r.db.Exec(`
		INSERT INTO attendance_sessions (
			role_scope,course_code,course_name,class_name,session_number,session_date,session_time,room,lab,topic,
			total_students,present,absent,sick,permit,excused,status,assistant_status,assistant_check_in_time
		)
		SELECT
			'assistant',
			c.class_code,
			c.name,
			c.class_code,
			cs.session_number,
			cs.session_date,
			cs.session_time,
			'',
			c.room,
			cs.topic,
			COALESCE(cardinality(cls.students), c.students, 0),
			0,
			COALESCE(cardinality(cls.students), c.students, 0),
			0,
			0,
			0,
			'Belum Presensi',
			'',
			''
		FROM courses c
		JOIN course_sessions cs ON cs.course_id = c.id
		LEFT JOIN classes cls ON cls.code = c.class_code
		WHERE EXISTS (SELECT 1 FROM lab_assistants la WHERE la.name = c.assistant)
			AND NOT EXISTS (
				SELECT 1
				FROM attendance_sessions s
				WHERE s.role_scope = 'assistant'
					AND s.course_code = c.class_code
					AND s.session_number = cs.session_number
			)
	`); err != nil {
		return err
	}
	_, err := r.db.Exec(`
		INSERT INTO attendance_records (session_id,nim,name,status,attendance_time,check_in_time)
		SELECT s.id, st.student_id, st.name, 'Alpa', '', ''
		FROM attendance_sessions s
		JOIN classes cls ON cls.code = s.course_code
		JOIN students st ON st.id = ANY(cls.students)
		WHERE s.role_scope = 'assistant'
			AND NOT EXISTS (
				SELECT 1
				FROM attendance_records r
				WHERE r.session_id = s.id
			)
		ORDER BY s.id, st.id
	`)
	return err
}

func (r *Repository) UpdateAssistantAttendanceSession(id int, payload *models.AssistantAttendanceUpdate) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM attendance_sessions WHERE id=$1 AND role_scope='assistant')`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	if _, err := tx.Exec(`DELETE FROM attendance_records WHERE session_id=$1`, id); err != nil {
		return err
	}

	present, absent, sick, permit := 0, 0, 0, 0
	for _, record := range payload.Records {
		status := normalizeAttendanceStatus(record.Status)
		switch status {
		case "Hadir":
			present++
		case "Sakit":
			sick++
		case "Izin":
			permit++
		default:
			absent++
		}
		timeValue := record.Time
		if status != "Hadir" {
			timeValue = ""
		}
		if _, err := tx.Exec(`INSERT INTO attendance_records (session_id,nim,name,status,attendance_time,check_in_time) VALUES ($1,$2,$3,$4,$5,$5)`, id, record.NIM, record.Name, status, timeValue); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`
		UPDATE attendance_sessions
		SET total_students=$1,present=$2,absent=$3,sick=$4,permit=$5,status='Selesai'
		WHERE id=$6
	`, len(payload.Records), present, absent, sick, permit, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) UpdateCourseSessionAttendance(courseSessionID int, payload *models.AssistantAttendanceUpdate) error {
	attendanceSessionID, err := r.ensureAttendanceSessionForCourseSession(courseSessionID)
	if err != nil {
		return err
	}
	return r.UpdateAssistantAttendanceSession(attendanceSessionID, payload)
}

func (r *Repository) ensureAttendanceSessionForCourseSession(courseSessionID int) (int, error) {
	var attendanceSessionID int
	err := r.db.QueryRow(`
		WITH src AS (
			SELECT
				cs.id,
				cs.session_number,
				cs.session_date,
				cs.session_time,
				cs.topic,
				c.class_code,
				c.name course_name,
				c.room,
				COALESCE(cardinality(cls.students), c.students, 0) total_students
			FROM course_sessions cs
			JOIN courses c ON c.id = cs.course_id
			LEFT JOIN classes cls ON cls.code = c.class_code
			WHERE cs.id = $1
		), existing AS (
			SELECT s.id
			FROM attendance_sessions s
			JOIN src ON src.class_code = s.course_code AND src.session_number = s.session_number
			WHERE s.role_scope='assistant'
			LIMIT 1
		), inserted AS (
			INSERT INTO attendance_sessions (
				role_scope,course_code,course_name,class_name,session_number,session_date,session_time,room,lab,topic,
				total_students,present,absent,sick,permit,excused,status,assistant_status,assistant_check_in_time
			)
			SELECT
				'assistant', class_code, course_name, class_code, session_number, session_date, session_time, '', room, topic,
				total_students, 0, total_students, 0, 0, 0, 'Belum Presensi', '', ''
			FROM src
			WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING id
		)
		SELECT id FROM inserted
		UNION ALL
		SELECT id FROM existing
		LIMIT 1
	`, courseSessionID).Scan(&attendanceSessionID)
	if err != nil {
		return 0, err
	}

	_, err = r.db.Exec(`
		INSERT INTO attendance_records (session_id,nim,name,status,attendance_time,check_in_time)
		SELECT s.id, st.student_id, st.name, 'Alpa', '', ''
		FROM attendance_sessions s
		JOIN classes cls ON cls.code = s.course_code
		JOIN students st ON st.id = ANY(cls.students)
		WHERE s.id=$1
			AND NOT EXISTS (SELECT 1 FROM attendance_records r WHERE r.session_id=s.id)
		ORDER BY st.id
	`, attendanceSessionID)
	return attendanceSessionID, err
}

func (r *Repository) CreateSessionAssignment(a *models.SessionAssignment) error {
	if strings.TrimSpace(a.Title) == "" {
		return errors.New("judul tugas wajib diisi")
	}
	if strings.TrimSpace(a.Deadline) == "" {
		return errors.New("deadline tugas wajib diisi")
	}
	err := r.db.QueryRow(`
		INSERT INTO session_assignments (course_id,session_id,title,description,due_date,status,total_students)
		SELECT cs.course_id, cs.id, $1, COALESCE($2,''), $3, 'pending', COALESCE(cardinality(cls.students), c.students, 0)
		FROM course_sessions cs
		JOIN courses c ON c.id=cs.course_id
		LEFT JOIN classes cls ON cls.code=c.class_code
		WHERE cs.id=$4
		RETURNING id, course_id
	`, a.Title, a.Description, a.Deadline, a.SessionID).Scan(&a.ID, &a.CourseID)
	if err != nil {
		return err
	}
	return r.loadSessionAssignment(a.ID, a)
}

func (r *Repository) UpdateSessionAssignment(id int, a *models.SessionAssignment) error {
	if strings.TrimSpace(a.Title) == "" {
		return errors.New("judul tugas wajib diisi")
	}
	if strings.TrimSpace(a.Deadline) == "" {
		return errors.New("deadline tugas wajib diisi")
	}
	result, err := r.db.Exec(`
		UPDATE session_assignments
		SET title=$1, description=COALESCE($2,''), due_date=$3
		WHERE id=$4
	`, a.Title, a.Description, a.Deadline, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return r.loadSessionAssignment(id, a)
}

func (r *Repository) DeleteSessionAssignment(id int) error {
	result, err := r.db.Exec(`DELETE FROM session_assignments WHERE id=$1`, id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) loadSessionAssignment(id int, a *models.SessionAssignment) error {
	return r.db.QueryRow(`
		SELECT sa.id, sa.course_id, COALESCE(sa.session_id,0), sa.title, sa.description, sa.due_date,
			sa.status, sa.submitted_count, sa.total_students, COALESCE(cs.session_number,0)
		FROM session_assignments sa
		LEFT JOIN course_sessions cs ON cs.id=sa.session_id
		WHERE sa.id=$1
	`, id).Scan(&a.ID, &a.CourseID, &a.SessionID, &a.Title, &a.Description, &a.Deadline, &a.Status, &a.SubmittedCount, &a.TotalStudents, &a.SessionNumber)
}

func normalizeAttendanceStatus(status string) string {
	switch status {
	case "Hadir", "Sakit", "Izin":
		return status
	default:
		return "Alpa"
	}
}

func (r *Repository) UpdateStudentGrade(id int, g *models.StudentGradeUpdate) error {
	_, err := r.db.Exec(`UPDATE lecturer_student_grades SET tugas1=$1,tugas2=$2,tugas3=$3,ujian_akhir=$4,nilai_akhir=$5,grade=$6 WHERE id=$7`, g.Tugas1, g.Tugas2, g.Tugas3, g.UjianAkhir, g.NilaiAkhir, g.Grade, id)
	return err
}

func (r *Repository) ReviewAssistantReport(id int, review *models.ReportReview) error {
	if review.Status == "" {
		review.Status = "Disetujui"
	}
	_, err := r.db.Exec(`UPDATE assistant_reports SET score=$1,status=$2 WHERE id=$3`, review.Score, review.Status, id)
	return err
}

func (r *Repository) AssistantReportsForRole(role string, userID int) ([]models.ReportItem, error) {
	where := `1=1`
	switch role {
	case "aslab":
		where = `1=1`
	case "laboran":
		where = `status IN ('submitted_to_laboran','rejected_by_kalab','Menunggu Review')`
	case "kalab":
		where = `status IN ('submitted_to_kalab','approved_by_laboran','Disetujui')`
	case "admin":
		where = `1=1`
	}
	rows, err := r.db.Query(`
		SELECT id,course_id,nim,name,course_code,course_name,class_name,week,topic,submitted_at,status,score,
			file_name,file_size,file_path,rejection_note,returned_to_role
		FROM assistant_reports
		WHERE ` + where + `
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.ReportItem{}
	for rows.Next() {
		var item models.ReportItem
		if err := rows.Scan(&item.ID, &item.CourseID, &item.NIM, &item.Name, &item.CourseCode, &item.CourseName, &item.Class, &item.Week, &item.Topic, &item.SubmittedAt, &item.Status, &item.Score, &item.FileName, &item.FileSize, &item.FileURL, &item.RejectionNote, &item.ReturnedToRole); err != nil {
			return nil, err
		}
		item.FileURL = "/api/reports/" + fmt.Sprint(item.ID) + "/file"
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ReportFilePath(id int, role string) (string, error) {
	var path string
	if err := r.db.QueryRow(`SELECT file_path FROM assistant_reports WHERE id=$1`, id).Scan(&path); err != nil {
		return "", err
	}
	if strings.TrimSpace(path) == "" {
		return "", sql.ErrNoRows
	}
	return path, nil
}

func (r *Repository) ReportDocument(id int, role string, userID int) (*models.ReportDocument, error) {
	doc := &models.ReportDocument{}
	var score sql.NullInt64
	err := r.db.QueryRow(`
		SELECT ar.id, ar.course_id, ar.nim, ar.name, ar.course_code, ar.course_name, ar.class_name, ar.week, ar.topic,
			ar.submitted_at, ar.status, ar.score, ar.file_name, ar.file_size, ar.rejection_note, ar.returned_to_role,
			COALESCE(c.study_program,''), COALESCE(c.academic_year,''), COALESCE(ay.semester,''), COALESCE(c.instructor,''),
			COALESCE(c.assistant, ar.name), COALESCE(c.credits,1), COALESCE(NULLIF(c.sessions,0),0), COALESCE(NULLIF(c.students,0), cardinality(cls.students), 0)
		FROM assistant_reports ar
		LEFT JOIN courses c ON c.id=ar.course_id OR (ar.course_id IS NULL AND c.class_code=ar.course_code)
		LEFT JOIN classes cls ON cls.code=COALESCE(NULLIF(c.class_code,''), ar.class_name, ar.course_code)
		LEFT JOIN academic_years ay ON ay.name=c.academic_year
		WHERE ar.id=$1
		ORDER BY c.id NULLS LAST
		LIMIT 1
	`, id).Scan(
		&doc.Report.ID, &doc.Report.CourseID, &doc.Report.NIM, &doc.Report.Name, &doc.Report.CourseCode, &doc.Report.CourseName,
		&doc.Report.Class, &doc.Report.Week, &doc.Report.Topic, &doc.Report.SubmittedAt, &doc.Report.Status, &score,
		&doc.Report.FileName, &doc.Report.FileSize, &doc.Report.RejectionNote, &doc.Report.ReturnedToRole,
		&doc.Program, &doc.AcademicYear, &doc.Semester, &doc.Instructor, &doc.Assistant, &doc.Credits, &doc.TotalSessions, &doc.TotalStudents,
	)
	if err != nil {
		return nil, err
	}
	if score.Valid {
		value := int(score.Int64)
		doc.Report.Score = &value
	}
	doc.Report.FileURL = "/api/reports/" + fmt.Sprint(doc.Report.ID) + "/file"
	if err := r.loadReportInstitution(doc); err != nil {
		return nil, err
	}
	weights, passingGrade, err := r.reportWeights(doc.Report.CourseID)
	if err != nil {
		return nil, err
	}
	doc.PassingGrade = passingGrade

	rows, err := r.db.Query(`
		WITH report AS (
			SELECT ar.id, ar.course_id, ar.course_code, ar.class_name, ar.week
			FROM assistant_reports ar
			WHERE ar.id=$1
		), course_ref AS (
			SELECT COALESCE(c.id, report.course_id) course_id,
				COALESCE(NULLIF(c.class_code,''), report.class_name, report.course_code) class_code,
				report.week
			FROM report
			LEFT JOIN courses c ON c.id=report.course_id OR (report.course_id IS NULL AND c.class_code=report.course_code)
			ORDER BY c.id NULLS LAST
			LIMIT 1
		), student_list AS (
			SELECT st.id student_db_id, st.student_id nim, st.name
			FROM course_ref cr
			JOIN classes cls ON cls.code=cr.class_code
			JOIN students st ON st.id = ANY(cls.students)
			UNION
			SELECT 0, ar.nim, ar.name
			FROM assistant_reports ar
			WHERE ar.id=$1
				AND NOT EXISTS (
					SELECT 1
					FROM course_ref cr
					JOIN classes cls ON cls.code=cr.class_code
					JOIN students st ON st.id = ANY(cls.students)
				)
		), attendance AS (
			SELECT sl.nim,
				COUNT(ar.id)::int meetings,
				COUNT(*) FILTER (WHERE ar.status='Hadir')::int present,
				COUNT(*) FILTER (WHERE ar.status='Alpa')::int absent,
				COUNT(*) FILTER (WHERE ar.status='Izin')::int permit,
				COUNT(*) FILTER (WHERE ar.status='Sakit')::int sick
			FROM student_list sl
			CROSS JOIN course_ref cr
			LEFT JOIN attendance_sessions ats ON ats.role_scope='assistant' AND ats.course_code=cr.class_code
			LEFT JOIN attendance_records ar ON ar.session_id=ats.id AND ar.nim=sl.nim
			GROUP BY sl.nim
		)
		SELECT row_number() OVER (ORDER BY sl.nim)::int no, sl.nim, sl.name,
			CASE WHEN COALESCE(att.meetings,0) > 0 THEN round((COALESCE(att.present,0)::numeric / att.meetings::numeric) * 100, 2) ELSE 0 END::float8 attendance_score,
			COALESCE((SELECT round(avg(sar.score)::numeric,2) FROM session_assessment_results sar JOIN course_ref cr ON true JOIN course_sessions cs ON cs.course_id=cr.course_id AND cs.id=sar.session_id WHERE sar.student_id=sl.student_db_id AND sar.assessment_type='pretest' AND sar.status='completed'), 0)::float8 pretest,
			COALESCE((SELECT round(avg(ss.score)::numeric,2) FROM student_submissions ss JOIN session_assignments sa ON sa.id=ss.assignment_id JOIN course_ref cr ON cr.course_id=sa.course_id WHERE ss.student_id=sl.student_db_id AND ss.score IS NOT NULL), 0)::float8 assignment_score,
			CASE WHEN COALESCE(att.meetings,0) > 0 THEN round((COALESCE(att.present,0)::numeric / att.meetings::numeric) * 100, 2) ELSE 0 END::float8 praktikum,
			COALESCE((SELECT round(avg(sar.score)::numeric,2) FROM session_assessment_results sar JOIN course_ref cr ON true JOIN course_sessions cs ON cs.course_id=cr.course_id AND cs.id=sar.session_id WHERE sar.student_id=sl.student_db_id AND sar.assessment_type='posttest' AND sar.status='completed'), 0)::float8 posttest,
			(
				(CASE WHEN COALESCE(att.present,0) > 0 THEN 25 ELSE 0 END) +
				(CASE WHEN EXISTS (SELECT 1 FROM session_assessment_results sar JOIN course_ref cr ON true JOIN course_sessions cs ON cs.course_id=cr.course_id AND cs.id=sar.session_id WHERE sar.student_id=sl.student_db_id AND sar.assessment_type='pretest' AND sar.status='completed') THEN 25 ELSE 0 END) +
				(CASE WHEN EXISTS (SELECT 1 FROM course_materials cm JOIN course_ref cr ON cr.course_id=cm.course_id WHERE cm.deleted_at IS NULL AND cm.status IN ('submitted','available','approved')) THEN 25 ELSE 0 END) +
				(CASE WHEN EXISTS (SELECT 1 FROM session_assessment_results sar JOIN course_ref cr ON true JOIN course_sessions cs ON cs.course_id=cr.course_id AND cs.id=sar.session_id WHERE sar.student_id=sl.student_db_id AND sar.assessment_type='posttest' AND sar.status='completed') THEN 25 ELSE 0 END)
			)::float8 progress,
			COALESCE(att.meetings,0), COALESCE(att.present,0), COALESCE(att.absent,0), COALESCE(att.permit,0), COALESCE(att.sick,0)
		FROM student_list sl
		LEFT JOIN attendance att ON att.nim=sl.nim
		ORDER BY sl.nim
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doc.Students = []models.ReportStudent{}
	for rows.Next() {
		var student models.ReportStudent
		if err := rows.Scan(&student.No, &student.NIM, &student.Name, &student.AttendanceScore, &student.Pretest, &student.AssignmentScore, &student.Praktikum, &student.Posttest, &student.Progress, &student.Meetings, &student.Present, &student.Absent, &student.Permit, &student.Sick); err != nil {
			return nil, err
		}
		student.AttendancePercent = student.Praktikum
		student.FinalScore = roundScore(
			student.AttendanceScore*weights.attendance/100 +
				student.Pretest*weights.pretest/100 +
				student.AssignmentScore*weights.assignment/100 +
				student.Praktikum*weights.practicum/100 +
				student.Posttest*weights.posttest/100,
		)
		student.Grade, student.Passed = r.gradeForScore(student.FinalScore, passingGrade)
		doc.Students = append(doc.Students, student)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if doc.TotalStudents == 0 {
		doc.TotalStudents = len(doc.Students)
	}
	if err := r.loadReportPersonnel(doc); err != nil {
		return nil, err
	}
	if err := r.loadReportActivities(doc); err != nil {
		return nil, err
	}
	if err := r.loadReportSigners(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

type reportWeights struct {
	attendance float64
	pretest    float64
	assignment float64
	practicum  float64
	posttest   float64
}

func (r *Repository) loadReportInstitution(doc *models.ReportDocument) error {
	return r.db.QueryRow(`
		SELECT university_name,faculty_name,study_program_name,laboratory_name,campus_a_address,campus_b_address,website,email,phone,logo_path
		FROM institution_settings
		ORDER BY id
		LIMIT 1
	`).Scan(
		&doc.Institution.UniversityName,
		&doc.Institution.FacultyName,
		&doc.Institution.StudyProgramName,
		&doc.Institution.LaboratoryName,
		&doc.Institution.CampusAAddress,
		&doc.Institution.CampusBAddress,
		&doc.Institution.Website,
		&doc.Institution.Email,
		&doc.Institution.Phone,
		&doc.Institution.LogoPath,
	)
}

func (r *Repository) reportWeights(courseID *int) (reportWeights, float64, error) {
	weights := reportWeights{attendance: 10, pretest: 15, assignment: 20, practicum: 20, posttest: 35}
	passingGrade := 55.0
	if courseID == nil {
		return weights, passingGrade, nil
	}
	err := r.db.QueryRow(`
		SELECT attendance_weight::float8,pretest_weight::float8,assignment_weight::float8,practicum_weight::float8,posttest_weight::float8,passing_grade::float8
		FROM course_assessment_weights
		WHERE course_id=$1
	`, *courseID).Scan(&weights.attendance, &weights.pretest, &weights.assignment, &weights.practicum, &weights.posttest, &passingGrade)
	if errors.Is(err, sql.ErrNoRows) {
		return weights, passingGrade, nil
	}
	return weights, passingGrade, err
}

func (r *Repository) gradeForScore(score float64, passingGrade float64) (string, bool) {
	var grade string
	var isPassed bool
	err := r.db.QueryRow(`
		SELECT grade,is_passed
		FROM grade_scales
		WHERE $1 BETWEEN min_score AND max_score
		ORDER BY min_score DESC
		LIMIT 1
	`, score).Scan(&grade, &isPassed)
	if err == nil {
		return grade, isPassed && score >= passingGrade
	}
	return gradeFromScore(score), score >= passingGrade
}

func (r *Repository) loadReportPersonnel(doc *models.ReportDocument) error {
	var courseID interface{}
	if doc.Report.CourseID != nil {
		courseID = *doc.Report.CourseID
	}
	rows, err := r.db.Query(`
		SELECT name, role, identifier, note
		FROM (
			SELECT COALESCE(NULLIF(c.instructor,''), '-') name, 'Pengajar' role, '' identifier, 'Penanggung jawab kursus' note, 1 sort_order
			FROM courses c
			WHERE ($1::int IS NOT NULL AND c.id=$1)
			UNION ALL
			SELECT COALESCE(NULLIF(c.assistant,''), '-') name, 'Aslab' role, COALESCE(la.student_id,''), 'Asisten Laboratorium' note, 2 sort_order
			FROM courses c
			LEFT JOIN lab_assistants la ON la.name=c.assistant
			WHERE ($1::int IS NOT NULL AND c.id=$1)
			UNION ALL
			SELECT name,
				CASE WHEN role='kalab' THEN 'Kepala Lab' WHEN role='laboran' THEN 'Laboran' ELSE role END,
				student_id,
				CASE WHEN role='kalab' THEN 'Ka. Laboratorium' WHEN role='laboran' THEN 'Laboran' ELSE 'Personel' END,
				CASE WHEN role='kalab' THEN 3 ELSE 4 END
			FROM lab_assistants
			WHERE role IN ('kalab','laboran')
		) personnel
		WHERE name <> '-'
		ORDER BY sort_order, name
	`, courseID)
	if err != nil {
		return err
	}
	defer rows.Close()
	doc.Personnel = []models.ReportPerson{}
	for rows.Next() {
		var person models.ReportPerson
		if err := rows.Scan(&person.Name, &person.Role, &person.Identifier, &person.Note); err != nil {
			return err
		}
		doc.Personnel = append(doc.Personnel, person)
	}
	return rows.Err()
}

func (r *Repository) loadReportActivities(doc *models.ReportDocument) error {
	var courseID interface{}
	if doc.Report.CourseID != nil {
		courseID = *doc.Report.CourseID
	}
	rows, err := r.db.Query(`
		SELECT row_number() OVER (ORDER BY sal.created_at)::int,
			to_char(sal.created_at,'YYYY-MM-DD HH24:MI'),
			COALESCE(st.name,'-'),
			sal.activity_type,
			COALESCE(c.name,'-'),
			COALESCE('Sesi ' || cs.session_number::text,'-'),
			sal.description
		FROM student_activity_logs sal
		LEFT JOIN students st ON st.id=sal.student_id
		LEFT JOIN courses c ON c.id=sal.course_id
		LEFT JOIN course_sessions cs ON cs.id=sal.session_id
		WHERE ($1::int IS NULL OR sal.course_id=$1)
		ORDER BY sal.created_at
		LIMIT 100
	`, courseID)
	if err != nil {
		return err
	}
	defer rows.Close()
	doc.Activities = []models.ReportActivity{}
	for rows.Next() {
		var activity models.ReportActivity
		if err := rows.Scan(&activity.No, &activity.Time, &activity.StudentName, &activity.Activity, &activity.CourseName, &activity.SessionName, &activity.Description); err != nil {
			return err
		}
		doc.Activities = append(doc.Activities, activity)
	}
	return rows.Err()
}

func (r *Repository) loadReportSigners(doc *models.ReportDocument) error {
	rows, err := r.db.Query(`
		SELECT role,name,identifier_type,identifier_number,signature_path
		FROM report_signers
		WHERE is_active=TRUE
			AND lower(name) NOT IN ('popy meilina, m.kom','poppy melina, m.kom')
			AND role <> 'head_of_laboratory'
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	doc.Signers = []models.ReportSigner{}
	for rows.Next() {
		var signer models.ReportSigner
		if err := rows.Scan(&signer.Role, &signer.Name, &signer.IdentifierType, &signer.IdentifierNumber, &signer.SignaturePath); err != nil {
			return err
		}
		doc.Signers = append(doc.Signers, signer)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	var kalab models.ReportSigner
	err = r.db.QueryRow(`
		SELECT 'head_of_laboratory',
			name,
			CASE WHEN student_id <> '' THEN 'ID' ELSE '' END,
			COALESCE(student_id,''),
			''
		FROM lab_assistants
		WHERE role='kalab' AND status='Aktif'
		ORDER BY id
		LIMIT 1
	`).Scan(&kalab.Role, &kalab.Name, &kalab.IdentifierType, &kalab.IdentifierNumber, &kalab.SignaturePath)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	doc.Signers = append(doc.Signers, kalab)
	return nil
}

func roundScore(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func gradeFromScore(value float64) string {
	switch {
	case value >= 85:
		return "A"
	case value >= 80:
		return "A-"
	case value >= 75:
		return "B+"
	case value >= 70:
		return "B"
	case value >= 65:
		return "B-"
	case value >= 60:
		return "C+"
	case value >= 55:
		return "C"
	case value >= 45:
		return "D"
	default:
		return "E"
	}
}

func (r *Repository) ApproveReportByRole(id int, role string, userID int) error {
	switch role {
	case "laboran", "admin":
		_, err := r.db.Exec(`
			UPDATE assistant_reports
			SET status='submitted_to_kalab',
				laboran_approved_by=$1,
				laboran_approved_at=now(),
				submitted_to_kalab_at=now(),
				rejection_note='',
				returned_to_role=''
			WHERE id=$2
		`, userID, id)
		return err
	case "kalab":
		_, err := r.db.Exec(`
			UPDATE assistant_reports
			SET status='completed',
				kalab_approved_by=$1,
				kalab_approved_at=now(),
				rejection_note='',
				returned_to_role=''
			WHERE id=$2
		`, userID, id)
		return err
	default:
		return errors.New("role tidak berhak approve laporan")
	}
}

func (r *Repository) RejectReportByRole(id int, role string, userID int, note string) error {
	if strings.TrimSpace(note) == "" {
		return errors.New("catatan penolakan wajib diisi")
	}
	switch role {
	case "laboran", "admin":
		_, err := r.db.Exec(`
			UPDATE assistant_reports
			SET status='rejected_by_laboran',
				rejected_by=$1,
				rejected_at=now(),
				rejection_note=$2,
				returned_to_role='aslab'
			WHERE id=$3
		`, userID, note, id)
		return err
	case "kalab":
		_, err := r.db.Exec(`
			UPDATE assistant_reports
			SET status='rejected_by_kalab',
				rejected_by=$1,
				rejected_at=now(),
				rejection_note=$2,
				returned_to_role='laboran'
			WHERE id=$3
		`, userID, note, id)
		return err
	default:
		return errors.New("role tidak berhak tolak laporan")
	}
}

func (r *Repository) SubmitAssistantSessionReport(sessionID int) (map[string]interface{}, error) {
	var reportID int
	err := r.db.QueryRow(`
		WITH src AS (
			SELECT cs.session_number, cs.topic, c.class_code, c.name course_name, c.class_code class_name,
				COALESCE(NULLIF(la.student_id,''), '-') nim,
				COALESCE(NULLIF(la.name,''), c.assistant) assistant_name
			FROM course_sessions cs
			JOIN courses c ON c.id = cs.course_id
			LEFT JOIN lab_assistants la ON la.name = c.assistant
			WHERE cs.id = $1
		), upsert AS (
			UPDATE assistant_reports ar
			SET status='submitted_to_laboran',
				submitted_at=to_char(now(),'YYYY-MM-DD HH24:MI'),
				submitted_to_laboran_at=now(),
				topic=src.topic,
				file_name='Laporan Sesi ' || src.session_number || ' - ' || src.course_name || '.pdf',
				file_size='-'
			FROM src
			WHERE ar.course_code=src.class_code
				AND ar.week=src.session_number
				AND ar.name=src.assistant_name
			RETURNING ar.id
		)
		INSERT INTO assistant_reports (course_id,nim,name,course_code,course_name,class_name,week,topic,submitted_at,status,score,file_name,file_size,submitted_to_laboran_at)
		SELECT (SELECT course_id FROM course_sessions WHERE id=$1),nim,assistant_name,class_code,course_name,class_name,session_number,topic,to_char(now(),'YYYY-MM-DD HH24:MI'),'submitted_to_laboran',NULL,'Laporan Sesi ' || session_number || ' - ' || course_name || '.pdf','-',now()
		FROM src
		WHERE NOT EXISTS (SELECT 1 FROM upsert)
		RETURNING id
	`, sessionID).Scan(&reportID)
	if err == sql.ErrNoRows {
		err = r.db.QueryRow(`
			SELECT ar.id
			FROM assistant_reports ar
			JOIN course_sessions cs ON cs.session_number=ar.week
			JOIN courses c ON c.id=cs.course_id AND c.class_code=ar.course_code
			WHERE cs.id=$1
			ORDER BY ar.id DESC
			LIMIT 1
		`, sessionID).Scan(&reportID)
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": reportID, "status": "submitted_to_laboran"}, nil
}

func (r *Repository) SetAssistantReportStatus(id int, status string) error {
	_, err := r.db.Exec(`UPDATE assistant_reports SET status=$1 WHERE id=$2`, status, id)
	return err
}

// GetAssistantReportWithCourse returns report with course info for authorization check
func (r *Repository) GetAssistantReportWithCourse(reportID int) (map[string]interface{}, error) {
	var courseCode, courseName string
	err := r.db.QueryRow(`
		SELECT ar.course_code, ar.course_name
		FROM assistant_reports ar
		WHERE ar.id=$1
	`, reportID).Scan(&courseCode, &courseName)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"courseCode": courseCode,
		"courseName": courseName,
	}, nil
}

// GetCourseLecturer returns lecturer name for a course
func (r *Repository) GetCourseLecturer(courseCode string) (string, error) {
	var lecturer string
	err := r.db.QueryRow(`
		SELECT lecturer FROM courses WHERE class_code=$1 LIMIT 1
	`, courseCode).Scan(&lecturer)
	return lecturer, err
}

func (r *Repository) ReportWorkflow() ([]models.ReportWorkflowItem, error) {
	rows, err := r.db.Query(`SELECT course_id, status, updated_at::text FROM report_workflows ORDER BY course_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.ReportWorkflowItem{}
	for rows.Next() {
		var x models.ReportWorkflowItem
		if err := rows.Scan(&x.CourseID, &x.Status, &x.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

func (r *Repository) UpsertReportWorkflow(courseID int, status string) error {
	_, err := r.db.Exec(`
		INSERT INTO report_workflows (course_id, status, updated_at)
		VALUES ($1, $2, now())
		ON CONFLICT (course_id) DO UPDATE SET status=EXCLUDED.status, updated_at=now()
	`, courseID, status)
	return err
}

func SQLPlaceholders(n int) string {
	p := make([]string, n)
	for i := range p {
		p[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(p, ",")
}

// UpsertStudentAttendanceByCourseSession saves student attendance using course_session_id.
// It upserts into attendance_sessions (role_scope='assistant') then replaces attendance_records.
func (r *Repository) UpsertStudentAttendanceByCourseSession(courseSessionID int, payload *models.AssistantAttendanceUpdate) error {
	// Get course info from course_sessions
	var courseCode, courseName, sessionTime, room string
	var sessionNumber int
	var sessionDate string
	err := r.db.QueryRow(`
		SELECT cs.session_number, cs.session_date, cs.session_time, cs.room,
			c.class_code, c.name
		FROM course_sessions cs
		JOIN courses c ON c.id = cs.course_id
		WHERE cs.id = $1
	`, courseSessionID).Scan(&sessionNumber, &sessionDate, &sessionTime, &room, &courseCode, &courseName)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert attendance_session
	var sessionID int
	err = tx.QueryRow(`SELECT id FROM attendance_sessions WHERE role_scope='assistant' AND course_code=$1 AND session_number=$2`, courseCode, sessionNumber).Scan(&sessionID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow(`
			INSERT INTO attendance_sessions
				(role_scope,course_code,course_name,class_name,session_number,session_date,session_time,room,lab,topic,total_students,present,absent,sick,permit,excused,status)
			VALUES ('assistant',$1,$2,$1,$3,$4,$5,$6,'','',0,0,0,0,0,0,'Selesai')
			RETURNING id
		`, courseCode, courseName, sessionNumber, sessionDate, sessionTime, room).Scan(&sessionID)
	} else if err == nil {
		_, err = tx.Exec(`UPDATE attendance_sessions SET session_date=$1,session_time=$2,room=$3,status='Selesai' WHERE id=$4`, sessionDate, sessionTime, room, sessionID)
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM attendance_records WHERE session_id=$1`, sessionID); err != nil {
		return err
	}

	present, absent, sick, permit := 0, 0, 0, 0
	for _, record := range payload.Records {
		status := normalizeAttendanceStatus(record.Status)
		switch status {
		case "Hadir":
			present++
		case "Sakit":
			sick++
		case "Izin":
			permit++
		default:
			absent++
		}
		timeValue := record.Time
		if status != "Hadir" {
			timeValue = ""
		}
		if _, err := tx.Exec(`INSERT INTO attendance_records (session_id,nim,name,status,attendance_time,check_in_time) VALUES ($1,$2,$3,$4,$5,$5)`,
			sessionID, record.NIM, record.Name, status, timeValue); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE attendance_sessions SET total_students=$1,present=$2,absent=$3,sick=$4,permit=$5 WHERE id=$6`,
		len(payload.Records), present, absent, sick, permit, sessionID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) UpsertSessionAssessment(sessionID int, payload *models.SessionAssessmentInput) (map[string]interface{}, error) {
	assessmentType := strings.ToLower(strings.TrimSpace(payload.Type))
	if assessmentType != "pretest" && assessmentType != "posttest" {
		return nil, errors.New("type harus pretest atau posttest")
	}
	maxScore := payload.MaxScore
	if maxScore <= 0 {
		maxScore = 100
	}
	status := strings.TrimSpace(payload.Status)
	if status == "" {
		if payload.Score != nil {
			status = "completed"
		} else {
			status = "not_started"
		}
	}
	title := strings.TrimSpace(payload.Title)
	if title == "" {
		if assessmentType == "pretest" {
			title = "Pretest"
		} else {
			title = "Post-test"
		}
	}
	var id int
	err := r.db.QueryRow(`
		INSERT INTO session_assessments (session_id,assessment_type,title,score,max_score,status,note,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,now())
		ON CONFLICT (session_id,assessment_type) DO UPDATE SET
			title=EXCLUDED.title,
			score=EXCLUDED.score,
			max_score=EXCLUDED.max_score,
			status=EXCLUDED.status,
			note=EXCLUDED.note,
			updated_at=now()
		RETURNING id
	`, sessionID, assessmentType, title, payload.Score, maxScore, status, payload.Note).Scan(&id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": id, "sessionId": sessionID, "type": assessmentType, "title": title,
		"score": payload.Score, "maxScore": maxScore, "status": status, "note": payload.Note,
	}, nil
}

func (r *Repository) SessionAssessments(sessionID int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT id,session_id,assessment_type,title,score,max_score,status,note
		FROM session_assessments
		WHERE session_id=$1
		ORDER BY assessment_type
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]interface{}{}
	for rows.Next() {
		var id, sid, maxScore int
		var assessmentType, title, status, note string
		var score *int
		if err := rows.Scan(&id, &sid, &assessmentType, &title, &score, &maxScore, &status, &note); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id": id, "sessionId": sid, "type": assessmentType, "title": title,
			"score": score, "maxScore": maxScore, "status": status, "note": note,
		})
	}
	return items, rows.Err()
}

func (r *Repository) UpsertStudentSessionAssessment(sessionID int, studentID int, payload *models.StudentSessionAssessmentInput) (map[string]interface{}, error) {
	assessmentType := strings.ToLower(strings.TrimSpace(payload.Type))
	if assessmentType != "pretest" && assessmentType != "posttest" {
		return nil, errors.New("type harus pretest atau posttest")
	}
	if studentID <= 0 {
		return nil, errors.New("student wajib dipilih")
	}
	maxScore := payload.MaxScore
	if maxScore <= 0 {
		maxScore = 100
	}
	status := strings.TrimSpace(payload.Status)
	if status == "" {
		if payload.Score != nil {
			status = "completed"
		} else {
			status = "not_started"
		}
	}
	var id int
	err := r.db.QueryRow(`
		INSERT INTO session_assessment_results (session_id,student_id,assessment_type,score,max_score,status,note,submitted_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,CASE WHEN $6='completed' THEN now() ELSE NULL END,now())
		ON CONFLICT (session_id,student_id,assessment_type) DO UPDATE SET
			score=EXCLUDED.score,
			max_score=EXCLUDED.max_score,
			status=EXCLUDED.status,
			note=EXCLUDED.note,
			submitted_at=CASE WHEN EXCLUDED.status='completed' THEN COALESCE(session_assessment_results.submitted_at, now()) ELSE NULL END,
			updated_at=now()
		RETURNING id
	`, sessionID, studentID, assessmentType, payload.Score, maxScore, status, payload.Note).Scan(&id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": id, "sessionId": sessionID, "studentId": studentID, "type": assessmentType,
		"score": payload.Score, "maxScore": maxScore, "status": status, "note": payload.Note,
	}, nil
}

func (r *Repository) SessionAssessmentsForStudent(sessionID int, studentID int) (map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT sar.id,sar.assessment_type,sar.score,sar.max_score,sar.status,sar.note,
			COALESCE(sa.title, CASE WHEN sar.assessment_type='pretest' THEN 'Pretest' ELSE 'Post-test' END) title
		FROM session_assessment_results sar
		LEFT JOIN session_assessments sa ON sa.session_id=sar.session_id AND sa.assessment_type=sar.assessment_type
		WHERE sar.session_id=$1 AND sar.student_id=$2
		ORDER BY sar.assessment_type
	`, sessionID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]interface{}{
		"pretest":  map[string]interface{}{"title": "Pretest", "score": nil, "maxScore": 100, "status": "not_started"},
		"posttest": map[string]interface{}{"title": "Post-test", "score": nil, "maxScore": 100, "status": "not_started"},
	}
	for rows.Next() {
		var id, maxScore int
		var assessmentType, status, note, title string
		var score sql.NullInt64
		if err := rows.Scan(&id, &assessmentType, &score, &maxScore, &status, &note, &title); err != nil {
			return nil, err
		}
		var scoreValue interface{}
		if score.Valid {
			scoreValue = int(score.Int64)
		}
		result[assessmentType] = map[string]interface{}{
			"id": id, "title": title, "score": scoreValue, "maxScore": maxScore, "status": status, "note": note,
		}
	}
	return result, rows.Err()
}

func (r *Repository) CreateSessionAssignmentInput(a *models.SessionAssignmentInput) (int, error) {
	if a.CourseID == 0 && a.SessionID != 0 {
		_ = r.db.QueryRow(`SELECT course_id FROM course_sessions WHERE id=$1`, a.SessionID).Scan(&a.CourseID)
	}
	var id int
	err := r.db.QueryRow(`
		INSERT INTO session_assignments (course_id, session_id, title, description, due_date, status, submitted_count, total_students)
		SELECT $1, NULLIF($2,0), $3, $4, $5, 'pending', 0,
			COALESCE((SELECT cardinality(cls.students) FROM courses c JOIN classes cls ON cls.code=c.class_code WHERE c.id=$1), 0)
		RETURNING id
	`, a.CourseID, a.SessionID, a.Title, a.Description, a.DueDate).Scan(&id)
	return id, err
}

// UpdateSessionAssignment updates an existing assignment (Aslab)
func (r *Repository) UpdateSessionAssignmentInput(id int, a *models.SessionAssignmentInput) error {
	_, err := r.db.Exec(`
		UPDATE session_assignments SET title=$1, description=$2, due_date=$3 WHERE id=$4
	`, a.Title, a.Description, a.DueDate, id)
	return err
}

// UpsertAssistantSessionAttendance saves aslab attendance for a session
func (r *Repository) UpsertAssistantSessionAttendance(sessionID int, a *models.AssistantSessionAttendance) error {
	_, err := r.db.Exec(`
		INSERT INTO assistant_session_attendance (session_id, status, check_in_time, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (session_id) DO UPDATE SET status=EXCLUDED.status, check_in_time=EXCLUDED.check_in_time, updated_at=now()
	`, sessionID, a.Status, a.CheckInTime)
	return err
}

// GetAssistantSessionAttendance returns aslab attendance for a session
func (r *Repository) GetAssistantSessionAttendance(sessionID int) (*models.AssistantSessionAttendance, error) {
	var a models.AssistantSessionAttendance
	err := r.db.QueryRow(`SELECT status, check_in_time FROM assistant_session_attendance WHERE session_id=$1`, sessionID).
		Scan(&a.Status, &a.CheckInTime)
	if err == sql.ErrNoRows {
		return &models.AssistantSessionAttendance{Status: "", CheckInTime: ""}, nil
	}
	return &a, err
}

// RejectAssistantReportWithNote rejects a report with an optional note
func (r *Repository) RejectAssistantReportWithNote(id int, note string) error {
	_, err := r.db.Exec(`UPDATE assistant_reports SET status='Ditolak', rejection_note=$1 WHERE id=$2`, note, id)
	return err
}

// UpdateCourseSession updates session date and sort_order (Admin)
func (r *Repository) UpdateCourseSession(id int, input *models.UpdateSessionInput) error {
	_, err := r.db.Exec(`
		UPDATE course_sessions SET session_date=$1, sort_order=$2 WHERE id=$3
	`, input.SessionDate, input.SortOrder, id)
	return err
}

// DeleteCourseSession menghapus satu sesi. Baris turunan (soal, jawaban, materi,
// tugas, penilaian) ikut terhapus lewat ON DELETE CASCADE pada foreign key-nya.
func (r *Repository) DeleteCourseSession(id int) error {
	_, err := r.db.Exec(`DELETE FROM course_sessions WHERE id=$1`, id)
	return err
}

// GetCourseSessions returns sessions for a course ordered by sort_order
func (r *Repository) GetCourseSessions(courseID int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT id, course_id, session_number, title, topic, session_date, session_time, session_type, COALESCE(conference_link,''), room, description, sort_order
		FROM course_sessions WHERE course_id=$1 ORDER BY sort_order, session_number
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]interface{}{}
	for rows.Next() {
		var id, courseId, sessionNumber, sortOrder int
		var title, topic, sessionDate, sessionTime, sessionType, conferenceLink, room, description string
		if err := rows.Scan(&id, &courseId, &sessionNumber, &title, &topic, &sessionDate, &sessionTime, &sessionType, &conferenceLink, &room, &description, &sortOrder); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id": id, "courseId": courseId, "sessionNumber": sessionNumber,
			"title": title, "topic": topic, "date": sessionDate, "time": sessionTime,
			"sessionType": sessionType, "conferenceLink": conferenceLink, "room": room,
			"description": description, "sortOrder": sortOrder,
		})
	}
	return items, rows.Err()
}

// SaveUploadedFile saves file metadata and returns the file_url
func (r *Repository) SaveMaterialFile(courseID, sessionID int, title, fileURL, fileType, fileSize string) (int, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO course_materials (course_id, session_id, title, material_type, size, file_url, upload_date, status)
		VALUES ($1, NULLIF($2,0), $3, $4, $5, $6, to_char(CURRENT_DATE,'DD Mon YYYY'), 'available')
		RETURNING id
	`, courseID, sessionID, title, fileType, fileSize, fileURL).Scan(&id)
	return id, err
}

// ExportReportsData returns all reports for XLSX export
func (r *Repository) ExportReportsData() ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT course_code, course_name, class_name, week, topic, nim, name, submitted_at, status, COALESCE(score,0), rejection_note
		FROM assistant_reports ORDER BY course_code, week
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []map[string]interface{}
	for rows.Next() {
		var courseCode, courseName, className, topic, nim, name, submittedAt, status, rejectionNote string
		var week, score int
		if err := rows.Scan(&courseCode, &courseName, &className, &week, &topic, &nim, &name, &submittedAt, &status, &score, &rejectionNote); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"courseCode": courseCode, "courseName": courseName, "class": className,
			"week": week, "topic": topic, "nim": nim, "name": name,
			"submittedAt": submittedAt, "status": status, "score": score, "rejectionNote": rejectionNote,
		})
	}
	return items, rows.Err()
}

// GetAssignmentSubmissions returns students from the class with their submission data
func (r *Repository) GetAssignmentSubmissions(assignmentID int) ([]map[string]interface{}, error) {
	// First try: get students from class
	rows, err := r.db.Query(`
		SELECT s.id, s.student_id, s.name, s.email,
			COALESCE(sub.answer_text,'') answer_text,
			COALESCE(sub.file_url,'') file_url,
			COALESCE(sub.file_name,'') file_name,
			COALESCE(sub.file_size,'') file_size,
			COALESCE(to_char(sub.submitted_at,'YYYY-MM-DD HH24:MI'),'') submitted_at,
			sub.score,
			COALESCE(sub.feedback,'') feedback
		FROM students s
		JOIN (
			SELECT unnest(cls.students) AS sid
			FROM session_assignments sa
			JOIN courses c ON c.id = sa.course_id
			JOIN classes cls ON cls.code = c.class_code
			WHERE sa.id = $1
		) class_students ON class_students.sid = s.id
		LEFT JOIN student_submissions sub ON sub.assignment_id = $1 AND sub.student_id = s.id
		ORDER BY s.name
	`, assignmentID)
	if err != nil {
		// Fallback: just get submissions directly
		rows, err = r.db.Query(`
			SELECT s.id, s.student_id, s.name, s.email,
				sub.answer_text, sub.file_url, sub.file_name, sub.file_size,
				to_char(sub.submitted_at,'YYYY-MM-DD HH24:MI'),
				sub.score, COALESCE(sub.feedback,'')
			FROM student_submissions sub
			JOIN students s ON s.id = sub.student_id
			WHERE sub.assignment_id = $1
			ORDER BY s.name
		`, assignmentID)
		if err != nil {
			return []map[string]interface{}{}, nil
		}
	}
	defer rows.Close()
	var items []map[string]interface{}
	for rows.Next() {
		var id int
		var nim, name, email, answerText, fileUrl, fileName, fileSize, submittedAt, feedback string
		var score *int
		if err := rows.Scan(&id, &nim, &name, &email, &answerText, &fileUrl, &fileName, &fileSize, &submittedAt, &score, &feedback); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id": id, "nim": nim, "name": name, "email": email,
			"answerText": answerText, "fileUrl": fileUrl, "fileName": fileName,
			"fileSize": fileSize, "submittedAt": submittedAt, "score": score, "feedback": feedback,
		})
	}
	if items == nil {
		items = []map[string]interface{}{}
	}
	return items, rows.Err()
}

// SubmitStudentAssignment upserts a student submission
func (r *Repository) SubmitStudentAssignment(assignmentID, studentID int, answerText, fileURL, fileName, fileSize string) error {
	_, err := r.db.Exec(`
		INSERT INTO student_submissions (assignment_id, student_id, answer_text, file_url, file_name, file_size, submitted_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (assignment_id, student_id) DO UPDATE SET
			answer_text=EXCLUDED.answer_text, file_url=EXCLUDED.file_url,
			file_name=EXCLUDED.file_name, file_size=EXCLUDED.file_size, submitted_at=now()
	`, assignmentID, studentID, answerText, fileURL, fileName, fileSize)
	if err != nil {
		return err
	}
	// Update submitted_count
	_, err = r.db.Exec(`
		UPDATE session_assignments SET submitted_count = (
			SELECT count(*) FROM student_submissions WHERE assignment_id = $1
		) WHERE id = $1
	`, assignmentID)
	return err
}

// GradeStudentSubmission updates score and feedback for a student submission by submission ID
func (r *Repository) GradeStudentSubmission(submissionID int, score int, feedback string) error {
	_, err := r.db.Exec(`UPDATE student_submissions SET score=$1, feedback=$2 WHERE id=$3`, score, feedback, submissionID)
	return err
}

// SyncSubmissionGradeToReport syncs grade from student_submissions to assistant_reports
func (r *Repository) SyncSubmissionGradeToReport(submissionID int) error {
	// Get submission details
	var studentID, assignmentID int
	err := r.db.QueryRow(`SELECT student_id, assignment_id FROM student_submissions WHERE id=$1`, submissionID).Scan(&studentID, &assignmentID)
	if err != nil {
		return err
	}

	// Get student nim
	var nim string
	err = r.db.QueryRow(`SELECT student_id FROM students WHERE id=$1`, studentID).Scan(&nim)
	if err != nil {
		return err
	}

	// Get assignment details (course_id, session_id)
	var courseID, sessionID sql.NullInt64
	err = r.db.QueryRow(`SELECT course_id, session_id FROM session_assignments WHERE id=$1`, assignmentID).Scan(&courseID, &sessionID)
	if err != nil {
		return err
	}

	// Get course code
	var courseCode string
	err = r.db.QueryRow(`SELECT class_code FROM courses WHERE id=$1`, courseID.Int64).Scan(&courseCode)
	if err != nil {
		return err
	}

	// Get score and feedback from submission
	var score sql.NullInt64
	var feedback string
	err = r.db.QueryRow(`SELECT score, feedback FROM student_submissions WHERE id=$1`, submissionID).Scan(&score, &feedback)
	if err != nil {
		return err
	}

	// Only update if we have valid session_id
	if !sessionID.Valid || sessionID.Int64 == 0 {
		return nil // Skip if no session
	}

	// Get session number
	var sessionNumber int
	err = r.db.QueryRow(`SELECT session_number FROM course_sessions WHERE id=$1`, sessionID.Int64).Scan(&sessionNumber)
	if err != nil {
		return err
	}

	// Update assistant_reports with specific nim, course_code, and week
	_, err = r.db.Exec(`
		UPDATE assistant_reports
		SET score=$1, feedback=$2
		WHERE nim=$3 AND course_code=$4 AND week=$5
	`, score, feedback, nim, courseCode, sessionNumber)
	return err
}

// GetSessionStudents returns students from the class linked to a course_session
func (r *Repository) GetSessionStudents(sessionID int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.student_id AS nim, s.name, '' AS status, '' AS time
		FROM course_sessions cs
		JOIN courses c ON c.id = cs.course_id
		JOIN classes cls ON cls.code = c.class_code
		JOIN students s ON s.id = ANY(cls.students)
		WHERE cs.id = $1
		ORDER BY s.name
	`, sessionID)
	if err != nil {
		return []map[string]interface{}{}, nil
	}
	defer rows.Close()
	var items []map[string]interface{}
	for rows.Next() {
		var id int
		var nim, name, status, t string
		if err := rows.Scan(&id, &nim, &name, &status, &t); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{"id": id, "nim": nim, "name": name, "status": status, "time": t})
	}
	if items == nil {
		items = []map[string]interface{}{}
	}
	return items, rows.Err()
}

// GetAssignmentTraceFlow returns trace flow of assignment submissions from upstream to downstream
func (r *Repository) GetAssignmentTraceFlow(assignmentID int) (map[string]interface{}, error) {
	var payload []byte
	err := r.db.QueryRow(`
		SELECT json_build_object(
			'assignment', json_build_object(
				'id', sa.id,
				'title', sa.title,
				'description', sa.description,
				'dueDate', sa.due_date,
				'courseId', sa.course_id,
				'courseName', c.name,
				'instructor', c.instructor,
				'assistant', c.assistant,
				'totalStudents', sa.total_students,
				'submittedCount', sa.submitted_count,
				'gradedCount', (SELECT count(*) FROM student_submissions WHERE assignment_id = $1 AND score IS NOT NULL),
				'pendingCount', sa.total_students - sa.submitted_count
			),
			'submissions', COALESCE((
				SELECT json_agg(json_build_object(
					'id', sub.id,
					'studentId', sub.student_id,
					'studentName', s.name,
					'studentNim', s.student_id,
					'answerText', COALESCE(sub.answer_text, ''),
					'fileUrl', COALESCE(sub.file_url, ''),
					'fileName', COALESCE(sub.file_name, ''),
					'fileSize', COALESCE(sub.file_size, ''),
					'submittedAt', COALESCE(to_char(sub.submitted_at, 'YYYY-MM-DD HH24:MI'), ''),
					'score', sub.score,
					'feedback', COALESCE(sub.feedback, ''),
					'status', CASE
						WHEN sub.id IS NULL THEN 'Belum Dikumpulkan'
						WHEN sub.score IS NULL THEN 'Dikumpulkan'
						ELSE 'Dinilai'
					END
				) ORDER BY s.name)
				FROM student_submissions sub
				JOIN students s ON s.id = sub.student_id
				WHERE sub.assignment_id = $1
			), '[]'::json),
			'timeline', COALESCE((
				SELECT json_agg(json_build_object(
					'date', DATE(sub.submitted_at),
					'count', count(*)
				) ORDER BY DATE(sub.submitted_at))
				FROM student_submissions sub
				WHERE sub.assignment_id = $1
				GROUP BY DATE(sub.submitted_at)
			), '[]'::json)
		)
		FROM session_assignments sa
		JOIN courses c ON c.id = sa.course_id
		WHERE sa.id = $1
	`, assignmentID).Scan(&payload)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetAssignmentStats returns statistics for an assignment
func (r *Repository) GetAssignmentStats(assignmentID int) (map[string]interface{}, error) {
	var payload []byte
	err := r.db.QueryRow(`
		SELECT json_build_object(
			'totalStudents', sa.total_students,
			'submitted', sa.submitted_count,
			'notSubmitted', sa.total_students - sa.submitted_count,
			'graded', (SELECT count(*) FROM student_submissions WHERE assignment_id = $1 AND score IS NOT NULL),
			'notGraded', (SELECT count(*) FROM student_submissions WHERE assignment_id = $1 AND score IS NULL),
			'submissionRate', CASE WHEN sa.total_students = 0 THEN 0 ELSE ROUND((sa.submitted_count::float / sa.total_students) * 100, 2) END,
			'averageScore', COALESCE(ROUND(AVG(sub.score)::numeric, 2), 0),
			'highestScore', COALESCE(MAX(sub.score), 0),
			'lowestScore', COALESCE(MIN(sub.score), 0),
			'scoreDistribution', COALESCE((
				SELECT json_object_agg(
					CASE
						WHEN sub.score >= 85 THEN 'A (85-100)'
						WHEN sub.score >= 70 THEN 'B (70-84)'
						WHEN sub.score >= 60 THEN 'C (60-69)'
						WHEN sub.score >= 50 THEN 'D (50-59)'
						ELSE 'E (<50)'
					END,
					count(*)
				)
				FROM student_submissions sub
				WHERE sub.assignment_id = $1 AND sub.score IS NOT NULL
			), '{}'::json)
		)
		FROM session_assignments sa
		LEFT JOIN student_submissions sub ON sub.assignment_id = sa.id
		WHERE sa.id = $1
		GROUP BY sa.id, sa.total_students, sa.submitted_count
	`, assignmentID).Scan(&payload)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetAssignmentGradeImpact returns impact of assignment scores on student grades
func (r *Repository) GetAssignmentGradeImpact(assignmentID int) (map[string]interface{}, error) {
	var payload []byte
	err := r.db.QueryRow(`
		SELECT json_build_object(
			'assignmentId', sa.id,
			'assignmentTitle', sa.title,
			'totalStudents', sa.total_students,
			'averageScore', COALESCE(ROUND(AVG(sub.score)::numeric, 2), 0),
			'impactAnalysis', COALESCE((
				SELECT json_agg(json_build_object(
					'studentId', s.id,
					'studentName', s.name,
					'studentNim', s.student_id,
					'assignmentScore', COALESCE(sub.score, 0),
					'submissionStatus', CASE
						WHEN sub.id IS NULL THEN 'Belum Dikumpulkan'
						WHEN sub.score IS NULL THEN 'Dikumpulkan'
						ELSE 'Dinilai'
					END,
					'potentialGradeImpact', CASE
						WHEN sub.score IS NULL THEN 'Tidak Ada Nilai'
						WHEN sub.score >= 85 THEN 'Meningkatkan Nilai'
						WHEN sub.score >= 70 THEN 'Mempertahankan Nilai'
						ELSE 'Menurunkan Nilai'
					END
				) ORDER BY s.name)
				FROM students s
				LEFT JOIN student_submissions sub ON sub.student_id = s.id AND sub.assignment_id = $1
				WHERE s.id IN (
					SELECT unnest(cls.students)
					FROM courses c
					JOIN classes cls ON cls.code = c.class_code
					WHERE c.id = sa.course_id
				)
			), '[]'::json),
			'summary', json_build_object(
				'excellentCount', (SELECT count(*) FROM student_submissions WHERE assignment_id = $1 AND score >= 85),
				'goodCount', (SELECT count(*) FROM student_submissions WHERE assignment_id = $1 AND score >= 70 AND score < 85),
				'averageCount', (SELECT count(*) FROM student_submissions WHERE assignment_id = $1 AND score >= 60 AND score < 70),
				'poorCount', (SELECT count(*) FROM student_submissions WHERE assignment_id = $1 AND score < 60),
				'notSubmittedCount', sa.total_students - sa.submitted_count
			)
		)
		FROM session_assignments sa
		WHERE sa.id = $1
	`, assignmentID).Scan(&payload)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetReportTraceFlow returns trace flow of assistant reports with impact on grades
func (r *Repository) GetReportTraceFlow(courseID int) (map[string]interface{}, error) {
	var payload []byte
	err := r.db.QueryRow(`
		SELECT json_build_object(
			'course', json_build_object(
				'courseCode', course_code,
				'courseName', course_name,
				'className', class_name,
				'totalReports', total_reports,
				'reviewed', reviewed,
				'pending', pending,
				'approved', approved,
				'needsRevision', needs_revision
			),
			'reports', COALESCE((
				SELECT json_agg(json_build_object(
					'id', ar.id,
					'nim', ar.nim,
					'name', ar.name,
					'week', ar.week,
					'topic', ar.topic,
					'submittedAt', ar.submitted_at,
					'status', ar.status,
					'score', ar.score,
					'fileName', ar.file_name,
					'fileSize', ar.file_size,
					'gradeImpact', CASE
						WHEN ar.score IS NULL THEN 'Belum Dinilai'
						WHEN ar.score >= 85 THEN 'Meningkatkan Nilai'
						WHEN ar.score >= 70 THEN 'Mempertahankan Nilai'
						ELSE 'Menurunkan Nilai'
					END
				) ORDER BY ar.week, ar.name)
				FROM assistant_reports ar
				WHERE ar.course_code = ars.course_code
			), '[]'::json),
			'summary', json_build_object(
				'excellentCount', (SELECT count(*) FROM assistant_reports WHERE course_code = ars.course_code AND score >= 85),
				'goodCount', (SELECT count(*) FROM assistant_reports WHERE course_code = ars.course_code AND score >= 70 AND score < 85),
				'averageCount', (SELECT count(*) FROM assistant_reports WHERE course_code = ars.course_code AND score >= 60 AND score < 70),
				'poorCount', (SELECT count(*) FROM assistant_reports WHERE course_code = ars.course_code AND score < 60),
				'notReviewedCount', ars.pending
			)
		)
		FROM assistant_report_summary ars
		WHERE ars.id = $1
	`, courseID).Scan(&payload)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetSubmissionDetail returns detail of a single submission with student and assignment info
func (r *Repository) GetSubmissionDetail(submissionID int) (map[string]interface{}, error) {
	var payload []byte
	err := r.db.QueryRow(`
		SELECT json_build_object(
			'submission', json_build_object(
				'id', sub.id,
				'assignmentId', sub.assignment_id,
				'studentId', sub.student_id,
				'answerText', COALESCE(sub.answer_text, ''),
				'fileUrl', COALESCE(sub.file_url, ''),
				'fileName', COALESCE(sub.file_name, ''),
				'fileSize', COALESCE(sub.file_size, ''),
				'submittedAt', COALESCE(to_char(sub.submitted_at, 'YYYY-MM-DD HH24:MI'), ''),
				'score', sub.score,
				'feedback', COALESCE(sub.feedback, '')
			),
			'student', json_build_object(
				'id', s.id,
				'name', s.name,
				'nim', s.student_id,
				'email', s.email
			),
			'assignment', json_build_object(
				'id', sa.id,
				'title', sa.title,
				'description', sa.description,
				'dueDate', sa.due_date,
				'courseId', sa.course_id,
				'courseName', c.name,
				'instructor', c.instructor,
				'assistant', c.assistant
			)
		)
		FROM student_submissions sub
		JOIN students s ON s.id = sub.student_id
		JOIN session_assignments sa ON sa.id = sub.assignment_id
		JOIN courses c ON c.id = sa.course_id
		WHERE sub.id = $1
	`, submissionID).Scan(&payload)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// StudentAssignmentSubmission mengembalikan pengumpulan milik satu mahasiswa untuk
// sebuah tugas, dipakai halaman "Lihat Detail" di sisi mahasiswa. Bila belum ada
// pengumpulan, dikembalikan bentuk kosong agar antarmuka tetap bisa ditampilkan.
func (r *Repository) StudentAssignmentSubmission(assignmentID int, studentID int) (map[string]interface{}, error) {
	var payload []byte
	err := r.db.QueryRow(`
		SELECT json_build_object(
			'submission', json_build_object(
				'id', sub.id,
				'assignmentId', sub.assignment_id,
				'studentId', sub.student_id,
				'answerText', COALESCE(sub.answer_text, ''),
				'fileUrl', COALESCE(sub.file_url, ''),
				'fileName', COALESCE(sub.file_name, ''),
				'fileSize', COALESCE(sub.file_size, ''),
				'submittedAt', COALESCE(to_char(sub.submitted_at, 'YYYY-MM-DD HH24:MI'), ''),
				'score', sub.score,
				'feedback', COALESCE(sub.feedback, '')
			),
			'student', json_build_object('id', s.id, 'name', s.name, 'nim', s.student_id, 'email', s.email),
			'assignment', json_build_object(
				'id', sa.id, 'title', sa.title, 'description', sa.description, 'dueDate', sa.due_date,
				'courseId', sa.course_id, 'courseName', c.name, 'instructor', c.instructor, 'assistant', c.assistant
			)
		)
		FROM student_submissions sub
		JOIN students s ON s.id = sub.student_id
		JOIN session_assignments sa ON sa.id = sub.assignment_id
		JOIN courses c ON c.id = sa.course_id
		WHERE sub.assignment_id = $1 AND sub.student_id = $2
	`, assignmentID, studentID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		// Tugas ada tapi belum dikumpulkan: kembalikan info tugas dengan submission kosong.
		var info []byte
		infoErr := r.db.QueryRow(`
			SELECT json_build_object(
				'submission', NULL,
				'student', (SELECT json_build_object('id', s.id, 'name', s.name, 'nim', s.student_id, 'email', s.email) FROM students s WHERE s.id=$2),
				'assignment', json_build_object(
					'id', sa.id, 'title', sa.title, 'description', sa.description, 'dueDate', sa.due_date,
					'courseId', sa.course_id, 'courseName', c.name, 'instructor', c.instructor, 'assistant', c.assistant
				)
			)
			FROM session_assignments sa JOIN courses c ON c.id = sa.course_id WHERE sa.id = $1
		`, assignmentID, studentID).Scan(&info)
		if infoErr != nil {
			return nil, infoErr
		}
		var empty map[string]interface{}
		if err := json.Unmarshal(info, &empty); err != nil {
			return nil, err
		}
		return empty, nil
	}
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}
