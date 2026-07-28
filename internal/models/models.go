package models

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Token string `json:"token"`
}

type AuthUser struct {
	ID    int
	Email string
	Role  string
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"`
}

type Schedule struct {
	Day       string `json:"day"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

type AttendanceSummary struct {
	Present    int `json:"present"`
	Total      int `json:"total"`
	Percentage int `json:"percentage"`
}

type StudentCourse struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Lecturer   string            `json:"lecturer"`
	Tutor      string            `json:"tutor"`
	Schedule   Schedule          `json:"schedule"`
	Room       string            `json:"room"`
	Attendance AttendanceSummary `json:"attendance"`
	Color      string            `json:"color"`
}

type DashboardCourse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Time       string `json:"time"`
	Room       string `json:"room"`
	Instructor string `json:"instructor"`
	Color      string `json:"color"`
}

type DashboardDeadline struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Course  string `json:"course"`
	DueDate string `json:"dueDate"`
	Urgent  bool   `json:"urgent"`
}

type Assignment struct {
	ID            int     `json:"id"`
	Title         string  `json:"title" binding:"required"`
	Course        string  `json:"course"`
	Assistant     string  `json:"assistant"`
	DueDate       string  `json:"dueDate" binding:"required"`
	DueTime       string  `json:"dueTime"`
	Status        string  `json:"status"`
	CourseColor   string  `json:"courseColor"`
	SubmittedDate *string `json:"submittedDate,omitempty"`
	Score         *int    `json:"score,omitempty"`
	AnswerText    string  `json:"answerText,omitempty"`
	FileURL       string  `json:"fileUrl,omitempty"`
	FileName      string  `json:"fileName,omitempty"`
	FileSize      string  `json:"fileSize,omitempty"`
}

type Grade struct {
	ID         int    `json:"id"`
	CourseName string `json:"courseName"`
	Code       string `json:"code"`
	Semester   string `json:"semester"`
	Credits    int    `json:"credits"`
	Grade      string `json:"grade"`
	Score      int    `json:"score"`
	Color      string `json:"color"`
}

type AdminCourse struct {
	ID           int    `json:"id"`
	Name         string `json:"name" binding:"required"`
	Instructor   string `json:"instructor"`
	Assistant    string `json:"assistant" binding:"required"`
	StudyProgram string `json:"studyProgram" binding:"required"`
	AcademicYear string `json:"academicYear" binding:"required"`
	ClassCode    string `json:"classCode" binding:"required"`
	Status       string `json:"status"`
	Day          string `json:"day"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	Room         string `json:"room" binding:"required"`
	Sessions     int    `json:"sessions" binding:"required,min=1"`
	Credits      int    `json:"credits" binding:"required,min=1"`
	Students     int    `json:"students"`
}

type AcademicYear struct {
	ID            int    `json:"id"`
	Name          string `json:"name" binding:"required"`
	StartDate     string `json:"startDate" binding:"required"`
	EndDate       string `json:"endDate" binding:"required"`
	Semester      string `json:"semester" binding:"required"`
	Status        string `json:"status"`
	TotalCourses  int    `json:"totalCourses"`
	TotalStudents int    `json:"totalStudents"`
}

type Student struct {
	ID                int      `json:"id"`
	Name              string   `json:"name" binding:"required"`
	Email             string   `json:"email" binding:"required,email"`
	Password          string   `json:"password"`
	DefaultPassword   string   `json:"defaultPassword"`
	StudentID         string   `json:"studentId" binding:"required"`
	Program           string   `json:"program" binding:"required"`
	Semester          int      `json:"semester" binding:"required,min=1"`
	Courses           []string `json:"courses"`
	Status            string   `json:"status"`
	JoinDate          string   `json:"joinDate"`
	IsPasswordChanged bool     `json:"isPasswordChanged"`
}

type Lecturer struct {
	ID                int      `json:"id"`
	Name              string   `json:"name" binding:"required"`
	Email             string   `json:"email" binding:"required,email"`
	Password          string   `json:"password"`
	DefaultPassword   string   `json:"defaultPassword"`
	NIDN              string   `json:"nidn" binding:"required"`
	Courses           []string `json:"courses"`
	IsPasswordChanged bool     `json:"isPasswordChanged"`
}

type LabAssistant struct {
	ID                int     `json:"id"`
	Name              string  `json:"name" binding:"required"`
	Email             string  `json:"email" binding:"required,email"`
	Phone             string  `json:"phone"`
	StudentID         string  `json:"studentId"`
	Role              string  `json:"role"`
	Lab               string  `json:"lab"`
	Supervisor        string  `json:"supervisor"`
	Semester          int     `json:"semester"`
	GPA               float64 `json:"gpa"`
	AssignedCourses   int     `json:"assignedCourses"`
	WeeklyHours       int     `json:"weeklyHours"`
	Status            string  `json:"status"`
	JoinDate          string  `json:"joinDate"`
	Password          string  `json:"password"`
	DefaultPassword   string  `json:"defaultPassword"`
	IsPasswordChanged bool    `json:"isPasswordChanged"`
}

type ClassData struct {
	ID            int     `json:"id"`
	Code          string  `json:"code" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	AcademicYear  string  `json:"academicYear" binding:"required"`
	Assistant     string  `json:"assistant"`
	Schedule      string  `json:"schedule"`
	Room          string  `json:"room"`
	TotalStudents int     `json:"totalStudents"`
	Capacity      int     `json:"capacity" binding:"required,min=1"`
	Students      []int64 `json:"students"`
}

type LecturerCourse struct {
	ID           int    `json:"id"`
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Class        string `json:"class" binding:"required"`
	Semester     string `json:"semester"`
	AcademicYear string `json:"academicYear"`
	Students     int    `json:"students"`
	Schedule     string `json:"schedule"`
	Room         string `json:"room"`
	SKS          int    `json:"sks" binding:"required,min=1"`
	Description  string `json:"description"`
	Materials    int    `json:"materials"`
	Assignments  int    `json:"assignments"`
}

type AdminAssignment struct {
	ID            int    `json:"id"`
	Title         string `json:"title" binding:"required"`
	Course        string `json:"course" binding:"required"`
	Instructor    string `json:"instructor"`
	DueDate       string `json:"dueDate" binding:"required"`
	TotalStudents int    `json:"totalStudents"`
	Submitted     int    `json:"submitted"`
	Graded        int    `json:"graded"`
	Pending       int    `json:"pending"`
	Status        string `json:"status"`
	Type          string `json:"type"`
}

type StudentGradeUpdate struct {
	Tugas1     int     `json:"tugas1"`
	Tugas2     int     `json:"tugas2"`
	Tugas3     int     `json:"tugas3"`
	UjianAkhir int     `json:"ujianAkhir"`
	NilaiAkhir float64 `json:"nilaiAkhir"`
	Grade      string  `json:"grade"`
}

type ReportReview struct {
	ID       int    `json:"id"`
	Score    int    `json:"score" binding:"min=0,max=100"`
	Status   string `json:"status"`
	Feedback string `json:"feedback"`
	Note     string `json:"note"`
}

type AssignmentSubmission struct {
	ID       int    `json:"id"`
	Answer   string `json:"answer"`
	FileName string `json:"fileName"`
}

type ReportWorkflowItem struct {
	CourseID  int    `json:"courseId"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}

type ReportWorkflowAction struct {
	CourseID int `json:"courseId" binding:"required"`
}

type AttendanceRecordInput struct {
	ID     int    `json:"id"`
	NIM    string `json:"nim" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Status string `json:"status" binding:"required"`
	Time   string `json:"time"`
}

type AssistantAttendanceUpdate struct {
	Records []AttendanceRecordInput `json:"records" binding:"required"`
}

type SessionAssessmentInput struct {
	Type     string `json:"type" binding:"required"`
	Title    string `json:"title"`
	Score    *int   `json:"score"`
	MaxScore int    `json:"maxScore"`
	Status   string `json:"status"`
	Note     string `json:"note"`
}

type StudentSessionAssessmentInput struct {
	Type      string `json:"type" binding:"required"`
	StudentID int    `json:"studentId"`
	Score     *int   `json:"score"`
	MaxScore  int    `json:"maxScore"`
	Status    string `json:"status"`
	Note      string `json:"note"`
}

// Soal pretest dan post-test memakai bentuk yang sama. Field Type menentukan
// jenis tes, dan boleh kosong pada rute lama yang selalu berarti pretest.
type QuizQuestionInput struct {
	Type          string   `json:"type"`
	Question      string   `json:"question" binding:"required"`
	Options       []string `json:"options" binding:"required"`
	CorrectOption int      `json:"correctOption" binding:"min=0"`
	Points        int      `json:"points"`
	Explanation   string   `json:"explanation"`
	SortOrder     int      `json:"sortOrder"`
}

// AnswerIndex sengaja tidak memakai binding required karena indeks 0 adalah
// pilihan A yang sah, sedangkan required menolak nilai nol.
type QuizAnswerInput struct {
	QuestionID  int `json:"questionId" binding:"required"`
	AnswerIndex int `json:"answerIndex"`
}

type QuizSubmissionInput struct {
	Type    string            `json:"type"`
	Answers []QuizAnswerInput `json:"answers" binding:"required"`
}

type QuizQuestion struct {
	ID            int      `json:"id"`
	SessionID     int      `json:"sessionId"`
	Type          string   `json:"type"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectOption *int     `json:"correctOption,omitempty"`
	Points        int      `json:"points"`
	Explanation   string   `json:"explanation,omitempty"`
	SortOrder     int      `json:"sortOrder"`
}

type QuizSubmission struct {
	ID          int               `json:"id"`
	SessionID   int               `json:"sessionId"`
	StudentID   int               `json:"studentId"`
	Type        string            `json:"type"`
	Answers     []QuizAnswerInput `json:"answers"`
	Score       int               `json:"score"`
	MaxScore    int               `json:"maxScore"`
	Status      string            `json:"status"`
	SubmittedAt string            `json:"submittedAt,omitempty"`
}

// Nama lama dipertahankan sebagai alias supaya pemanggil lama tetap kompilasi.
type PretestQuestionInput = QuizQuestionInput
type PretestAnswerInput = QuizAnswerInput
type PretestSubmissionInput = QuizSubmissionInput
type PretestQuestion = QuizQuestion
type PretestSubmission = QuizSubmission

type SessionAssignment struct {
	ID             int    `json:"id"`
	CourseID       int    `json:"courseId"`
	SessionID      int    `json:"sessionId"`
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description"`
	Deadline       string `json:"deadline"`
	DueDate        string `json:"dueDate"`
	Status         string `json:"status"`
	SubmittedCount int    `json:"submittedCount"`
	TotalStudents  int    `json:"totalStudents"`
	SessionNumber  int    `json:"sessionNumber"`
}

type MaterialItem struct {
	ID            int    `json:"id"`
	CourseID      int    `json:"courseId"`
	SessionID     *int   `json:"sessionId,omitempty"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	Size          string `json:"size"`
	UploadDate    string `json:"uploadDate"`
	Status        string `json:"status"`
	FileURL       string `json:"fileUrl"`
	CourseName    string `json:"courseName"`
	CreatedBy     *int   `json:"createdBy,omitempty"`
	RejectionNote string `json:"rejectionNote,omitempty"`
}

type ReportItem struct {
	ID             int    `json:"id"`
	CourseID       *int   `json:"courseId,omitempty"`
	NIM            string `json:"nim"`
	Name           string `json:"name"`
	CourseCode     string `json:"courseCode"`
	CourseName     string `json:"courseName"`
	Class          string `json:"class"`
	Week           int    `json:"week"`
	Topic          string `json:"topic"`
	SubmittedAt    string `json:"submittedAt"`
	Status         string `json:"status"`
	Score          *int   `json:"score,omitempty"`
	FileName       string `json:"fileName"`
	FileSize       string `json:"fileSize"`
	FileURL        string `json:"fileUrl"`
	RejectionNote  string `json:"rejectionNote"`
	ReturnedToRole string `json:"returnedToRole"`
}

type ReportDocument struct {
	Institution   ReportInstitution `json:"institution"`
	Report        ReportItem        `json:"report"`
	Program       string            `json:"program"`
	AcademicYear  string            `json:"academicYear"`
	Semester      string            `json:"semester"`
	Instructor    string            `json:"instructor"`
	Assistant     string            `json:"assistant"`
	Credits       int               `json:"credits"`
	TotalSessions int               `json:"totalSessions"`
	TotalStudents int               `json:"totalStudents"`
	PassingGrade  float64           `json:"passingGrade"`
	Students      []ReportStudent   `json:"students"`
	Personnel     []ReportPerson    `json:"personnel"`
	Activities    []ReportActivity  `json:"activityLogs"`
	Signers       []ReportSigner    `json:"signers"`
}

type ReportInstitution struct {
	UniversityName   string `json:"universityName"`
	FacultyName      string `json:"facultyName"`
	StudyProgramName string `json:"studyProgramName"`
	LaboratoryName   string `json:"laboratoryName"`
	CampusAAddress   string `json:"campusAAddress"`
	CampusBAddress   string `json:"campusBAddress"`
	Website          string `json:"website"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	LogoPath         string `json:"logoPath"`
}

type ReportStudent struct {
	No                int     `json:"no"`
	NIM               string  `json:"nim"`
	Name              string  `json:"name"`
	AttendanceScore   float64 `json:"attendanceScore"`
	Pretest           float64 `json:"pretest"`
	AssignmentScore   float64 `json:"assignmentScore"`
	Posttest          float64 `json:"posttest"`
	Praktikum         float64 `json:"praktikum"`
	FinalScore        float64 `json:"finalScore"`
	Grade             string  `json:"grade"`
	Passed            bool    `json:"passed"`
	Meetings          int     `json:"meetings"`
	Present           int     `json:"present"`
	Absent            int     `json:"absent"`
	Permit            int     `json:"permit"`
	Sick              int     `json:"sick"`
	AttendancePercent float64 `json:"attendancePercent"`
	Progress          float64 `json:"progress"`
}

type ReportPerson struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	Identifier string `json:"identifier"`
	Note       string `json:"note"`
}

type ReportActivity struct {
	No          int    `json:"no"`
	Time        string `json:"time"`
	StudentName string `json:"studentName"`
	Activity    string `json:"activity"`
	CourseName  string `json:"courseName"`
	SessionName string `json:"sessionName"`
	Description string `json:"description"`
}

type ReportSigner struct {
	Role             string `json:"role"`
	Name             string `json:"name"`
	IdentifierType   string `json:"identifierType"`
	IdentifierNumber string `json:"identifierNumber"`
	SignaturePath    string `json:"signaturePath"`
}

type ReportActionRequest struct {
	Note string `json:"note"`
}

type SessionAssignmentInput struct {
	CourseID    int    `json:"courseId" binding:"required"`
	SessionID   int    `json:"sessionId"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	DueDate     string `json:"dueDate" binding:"required"`
}

type AssistantSessionAttendance struct {
	Status      string `json:"status" binding:"required"`
	CheckInTime string `json:"checkInTime"`
}

type RejectReportInput struct {
	RejectionNote string `json:"rejectionNote"`
}

type UpdateSessionInput struct {
	SessionDate string `json:"sessionDate"`
	SortOrder   int    `json:"sortOrder"`
}

type CreateMaterialInput struct {
	CourseID  int    `json:"courseId" binding:"required"`
	SessionID int    `json:"sessionId"`
	Title     string `json:"title" binding:"required"`
	FileURL   string `json:"fileUrl" binding:"required"`
	FileType  string `json:"fileType"`
	FileSize  string `json:"fileSize"`
}

type StudentSubmissionInput struct {
	AssignmentID int    `json:"assignmentId" binding:"required"`
	StudentID    int    `json:"studentId" binding:"required"`
	AnswerText   string `json:"answerText"`
	FileURL      string `json:"fileUrl"`
	FileName     string `json:"fileName"`
	FileSize     string `json:"fileSize"`
}
