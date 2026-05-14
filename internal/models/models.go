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
	Instructor   string `json:"instructor" binding:"required"`
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
	StudentID         string  `json:"studentId" binding:"required"`
	Lab               string  `json:"lab" binding:"required"`
	Supervisor        string  `json:"supervisor"`
	Semester          int     `json:"semester" binding:"required,min=1"`
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
	ID            int    `json:"id"`
	Code          string `json:"code" binding:"required"`
	Name          string `json:"name" binding:"required"`
	AcademicYear  string `json:"academicYear" binding:"required"`
	Assistant     string `json:"assistant"`
	Schedule      string `json:"schedule"`
	Room          string `json:"room"`
	TotalStudents int    `json:"totalStudents"`
	Capacity      int    `json:"capacity" binding:"required,min=1"`
	Students      []int  `json:"students"`
}
