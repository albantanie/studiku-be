package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"studi-ku-backend/internal/models"
)

type Repository struct{ db *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Login(email string, password string) (*models.LoginUser, error) {
	var user models.LoginUser

	err := r.db.QueryRow(`
		SELECT id, name, email, role
		FROM (
			SELECT id, name, email, 'student' AS role, password FROM students
			UNION ALL
			SELECT id, name, email, 'lecturer' AS role, password FROM lecturers
			UNION ALL
			SELECT id, name, email, 'assistant' AS role, password FROM lab_assistants
			UNION ALL
			SELECT id, name, email, 'admin' AS role, password FROM admins
		) users
		WHERE lower(email) = lower($1)
		AND password = $2
		LIMIT 1
	`, email, password).Scan(&user.ID, &user.Name, &user.Email, &user.Role)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *Repository) Dashboard() (map[string]interface{}, error) {
	courses, err := r.StudentCourses()
	if err != nil {
		return nil, err
	}
	assignments, err := r.Assignments()
	if err != nil {
		return nil, err
	}

	today := make([]models.DashboardCourse, 0)
	for _, c := range courses {
		if len(today) == 2 {
			break
		}
		today = append(today, models.DashboardCourse{ID: c.ID, Name: c.Name, Time: c.Schedule.StartTime + " - " + c.Schedule.EndTime, Room: "Lab 301", Instructor: c.Lecturer, Color: c.Color})
	}
	deadlines := make([]models.DashboardDeadline, 0)
	for i, a := range assignments {
		if len(deadlines) == 3 {
			break
		}
		if a.Status == "pending" {
			deadlines = append(deadlines, models.DashboardDeadline{ID: a.ID, Title: a.Title, Course: a.Course, DueDate: a.DueDate, Urgent: i == 0})
		}
	}
	return map[string]interface{}{"todayCourses": today, "upcomingDeadlines": deadlines, "weeklyTaskCount": 5}, nil
}

func (r *Repository) StudentCourses() ([]models.StudentCourse, error) {
	rows, err := r.db.Query(`SELECT id, name, instructor, assistant, day, start_time, end_time, attendance_present, attendance_total, color FROM courses ORDER BY id LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.StudentCourse{}
	for rows.Next() {
		var c models.StudentCourse
		if err := rows.Scan(&c.ID, &c.Name, &c.Lecturer, &c.Tutor, &c.Schedule.Day, &c.Schedule.StartTime, &c.Schedule.EndTime, &c.Attendance.Present, &c.Attendance.Total, &c.Color); err != nil {
			return nil, err
		}
		if c.Attendance.Total > 0 {
			c.Attendance.Percentage = c.Attendance.Present * 100 / c.Attendance.Total
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *Repository) Assignments() ([]models.Assignment, error) {
	rows, err := r.db.Query(`SELECT id,title,course,assistant,due_date,due_time,status,course_color,submitted_date,score FROM assignments ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.Assignment{}
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(&a.ID, &a.Title, &a.Course, &a.Assistant, &a.DueDate, &a.DueTime, &a.Status, &a.CourseColor, &a.SubmittedDate, &a.Score); err != nil {
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
	rows, err := r.db.Query(`SELECT id,name,instructor,assistant,study_program,academic_year,class_code,status,day,start_time,end_time,room,sessions,credits,students FROM courses ORDER BY id`)
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
	return r.db.QueryRow(`INSERT INTO courses (name,instructor,assistant,study_program,academic_year,class_code,status,day,start_time,end_time,room,sessions,credits,students) VALUES ($1,$2,$3,$4,$5,$6,COALESCE(NULLIF($7,''),'Aktif'),$8,$9,$10,$11,$12,$13,$14) RETURNING id`, c.Name, c.Instructor, c.Assistant, c.StudyProgram, c.AcademicYear, c.ClassCode, c.Status, c.Day, c.StartTime, c.EndTime, c.Room, c.Sessions, c.Credits, c.Students).Scan(&c.ID)
}
func (r *Repository) UpdateCourse(id int, c *models.AdminCourse) error {
	_, err := r.db.Exec(`UPDATE courses SET name=$1,instructor=$2,assistant=$3,study_program=$4,academic_year=$5,class_code=$6,status=$7,day=$8,start_time=$9,end_time=$10,room=$11,sessions=$12,credits=$13,students=$14,updated_at=now() WHERE id=$15`, c.Name, c.Instructor, c.Assistant, c.StudyProgram, c.AcademicYear, c.ClassCode, c.Status, c.Day, c.StartTime, c.EndTime, c.Room, c.Sessions, c.Credits, c.Students, id)
	return err
}
func (r *Repository) DeleteCourse(id int) error {
	_, err := r.db.Exec(`DELETE FROM courses WHERE id=$1`, id)
	return err
}

func (r *Repository) AcademicYears() ([]models.AcademicYear, error) {
	rows, err := r.db.Query(`SELECT id,name,start_date::text,end_date::text,semester,status,total_courses,total_students FROM academic_years ORDER BY start_date DESC`)
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
		items = append(items, s)
	}
	return items, rows.Err()
}
func (r *Repository) CreateStudent(s *models.Student) error {
	if s.DefaultPassword == "" {
		s.DefaultPassword = "password"
	}
	if s.Password == "" {
		s.Password = "password"
	}
	return r.db.QueryRow(`INSERT INTO students (name,email,password,default_password,student_id,program,semester,courses,status,join_date,is_password_changed) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE(NULLIF($9,''),'Aktif'),COALESCE(NULLIF($10,'')::date,CURRENT_DATE),$11) RETURNING id`, s.Name, s.Email, s.Password, s.DefaultPassword, s.StudentID, s.Program, s.Semester, pq.Array(s.Courses), s.Status, s.JoinDate, s.IsPasswordChanged).Scan(&s.ID)
}
func (r *Repository) UpdateStudent(id int, s *models.Student) error {
	_, err := r.db.Exec(`UPDATE students SET name=$1,email=$2,password=$3,default_password=$4,student_id=$5,program=$6,semester=$7,courses=$8,status=$9,join_date=$10,is_password_changed=$11,updated_at=now() WHERE id=$12`, s.Name, s.Email, s.Password, s.DefaultPassword, s.StudentID, s.Program, s.Semester, pq.Array(s.Courses), s.Status, s.JoinDate, s.IsPasswordChanged, id)
	return err
}
func (r *Repository) DeleteStudent(id int) error {
	_, err := r.db.Exec(`DELETE FROM students WHERE id=$1`, id)
	return err
}
func (r *Repository) ResetStudentPassword(id int) (*models.Student, error) {
	_, err := r.db.Exec(`UPDATE students SET password=default_password,is_password_changed=false,updated_at=now() WHERE id=$1`, id)
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
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) CreateLecturer(x *models.Lecturer) error {
	if x.DefaultPassword == "" {
		x.DefaultPassword = "password"
	}
	if x.Password == "" {
		x.Password = x.DefaultPassword
	}
	return r.db.QueryRow(`INSERT INTO lecturers (name,email,password,default_password,nidn,courses,is_password_changed) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, x.Name, x.Email, x.Password, x.DefaultPassword, x.NIDN, pq.Array(x.Courses), x.IsPasswordChanged).Scan(&x.ID)
}
func (r *Repository) UpdateLecturer(id int, x *models.Lecturer) error {
	if x.DefaultPassword == "" {
		x.DefaultPassword = "password"
	}
	if x.Password == "" {
		x.Password = x.DefaultPassword
	}
	_, err := r.db.Exec(`UPDATE lecturers SET name=$1,email=$2,password=$3,default_password=$4,nidn=$5,courses=$6,is_password_changed=$7,updated_at=now() WHERE id=$8`, x.Name, x.Email, x.Password, x.DefaultPassword, x.NIDN, pq.Array(x.Courses), x.IsPasswordChanged, id)
	return err
}
func (r *Repository) DeleteLecturer(id int) error {
	_, err := r.db.Exec(`DELETE FROM lecturers WHERE id=$1`, id)
	return err
}
func (r *Repository) ResetLecturerPassword(id int) (*models.Lecturer, error) {
	_, err := r.db.Exec(`UPDATE lecturers SET password=default_password,is_password_changed=false,updated_at=now() WHERE id=$1`, id)
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
	rows, err := r.db.Query(`SELECT id,name,email,phone,student_id,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date::text,password,default_password,is_password_changed FROM lab_assistants ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.LabAssistant{}
	for rows.Next() {
		var x models.LabAssistant
		if err := rows.Scan(&x.ID, &x.Name, &x.Email, &x.Phone, &x.StudentID, &x.Lab, &x.Supervisor, &x.Semester, &x.GPA, &x.AssignedCourses, &x.WeeklyHours, &x.Status, &x.JoinDate, &x.Password, &x.DefaultPassword, &x.IsPasswordChanged); err != nil {
			return nil, err
		}
		items = append(items, x)
	}
	return items, rows.Err()
}
func (r *Repository) CreateAssistant(x *models.LabAssistant) error {
	if x.DefaultPassword == "" {
		x.DefaultPassword = "password"
	}
	if x.Password == "" {
		x.Password = x.DefaultPassword
	}
	return r.db.QueryRow(`INSERT INTO lab_assistants (name,email,phone,student_id,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date,password,default_password,is_password_changed) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,COALESCE(NULLIF($11,''),'Aktif'),COALESCE(NULLIF($12,'')::date,CURRENT_DATE),$13,$14,$15) RETURNING id`, x.Name, x.Email, x.Phone, x.StudentID, x.Lab, x.Supervisor, x.Semester, x.GPA, x.AssignedCourses, x.WeeklyHours, x.Status, x.JoinDate, x.Password, x.DefaultPassword, x.IsPasswordChanged).Scan(&x.ID)
}
func (r *Repository) UpdateAssistant(id int, x *models.LabAssistant) error {
	if x.DefaultPassword == "" {
		x.DefaultPassword = "password"
	}
	if x.Password == "" {
		x.Password = x.DefaultPassword
	}
	_, err := r.db.Exec(`UPDATE lab_assistants SET name=$1,email=$2,phone=$3,student_id=$4,lab=$5,supervisor=$6,semester=$7,gpa=$8,assigned_courses=$9,weekly_hours=$10,status=$11,join_date=$12,password=$13,default_password=$14,is_password_changed=$15,updated_at=now() WHERE id=$16`, x.Name, x.Email, x.Phone, x.StudentID, x.Lab, x.Supervisor, x.Semester, x.GPA, x.AssignedCourses, x.WeeklyHours, x.Status, x.JoinDate, x.Password, x.DefaultPassword, x.IsPasswordChanged, id)
	return err
}
func (r *Repository) DeleteAssistant(id int) error {
	_, err := r.db.Exec(`DELETE FROM lab_assistants WHERE id=$1`, id)
	return err
}
func (r *Repository) ResetAssistantPassword(id int) (*models.LabAssistant, error) {
	_, err := r.db.Exec(`UPDATE lab_assistants SET password=default_password,is_password_changed=false,updated_at=now() WHERE id=$1`, id)
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
	return r.db.QueryRow(`INSERT INTO classes (code,name,academic_year,assistant,schedule,room,total_students,capacity,students) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, c.Code, c.Name, c.AcademicYear, c.Assistant, c.Schedule, c.Room, c.TotalStudents, c.Capacity, pq.Array(c.Students)).Scan(&c.ID)
}
func (r *Repository) UpdateClass(id int, c *models.ClassData) error {
	c.TotalStudents = len(c.Students)
	_, err := r.db.Exec(`UPDATE classes SET code=$1,name=$2,academic_year=$3,assistant=$4,schedule=$5,room=$6,total_students=$7,capacity=$8,students=$9,updated_at=now() WHERE id=$10`, c.Code, c.Name, c.AcademicYear, c.Assistant, c.Schedule, c.Room, c.TotalStudents, c.Capacity, pq.Array(c.Students), id)
	return err
}
func (r *Repository) DeleteClass(id int) error {
	_, err := r.db.Exec(`DELETE FROM classes WHERE id=$1`, id)
	return err
}

func (r *Repository) CreateLecturerCourse(c *models.LecturerCourse) error {
	if c.SKS == 0 {
		c.SKS = 3
	}
	return r.db.QueryRow(`INSERT INTO courses (name,instructor,assistant,study_program,academic_year,class_code,status,day,start_time,end_time,room,sessions,credits,students,description) VALUES ($1,'','', 'Teknik Informatika',$2,$3,'Aktif',$4,'','',$5,14,$6,$7,$8) RETURNING id`, c.Name, c.AcademicYear, c.Code, c.Schedule, c.Room, c.SKS, c.Students, c.Description).Scan(&c.ID)
}

func (r *Repository) UpdateLecturerCourse(id int, c *models.LecturerCourse) error {
	if c.SKS == 0 {
		c.SKS = 3
	}
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
	_, err := r.db.Exec(`DELETE FROM course_materials WHERE id=$1`, id)
	return err
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
