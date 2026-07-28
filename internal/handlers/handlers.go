package handlers

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"studi-ku-backend/internal/models"
	"studi-ku-backend/internal/repositories"
)

type Handler struct {
	repo      *repositories.Repository
	uploadDir string
}

func New(repo *repositories.Repository) *Handler {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if strings.TrimSpace(uploadDir) == "" {
		uploadDir = "uploads"
	}
	return &Handler{repo: repo, uploadDir: uploadDir}
}

func Register(r *gin.Engine, h *Handler) {
	r.GET("/api/health", func(c *gin.Context) { ok(c, "healthy", gin.H{"status": "ok"}) })
	api := r.Group("/api")
	{
		api.POST("/auth/login", h.Login)
		api.Use(h.AuthRequired())
		registerPageData(api, h)
		api.PUT("/profile/password", h.ChangePassword)

		api.GET("/student/dashboard", requireRole("student"), h.GetDashboard)
		api.GET("/student/courses", requireRole("student"), h.GetStudentCourses)
		api.GET("/student/assignments", requireRole("student"), h.GetAssignments)
		api.GET("/student/grades", requireRole("student"), h.GetGrades)
		api.PUT("/student/assignments/:id/submit", requireRole("student"), h.SubmitAssignment)
		api.POST("/student/assignments/submit", requireRole("student"), h.SubmitStudentAssignment)
		api.PUT("/lecturer/grades/students/:id", requireRole("kalab"), h.UpdateStudentGrade)
		api.POST("/lecturer/courses", requireRole("laboran", "aslab"), h.CreateLecturerCourse)
		api.PUT("/lecturer/courses/:id", requireRole("admin", "laboran", "aslab", "kalab"), h.UpdateLecturerCourse)
		api.DELETE("/lecturer/courses/:id", requireRole("laboran"), h.DeleteLecturerCourse)
		api.DELETE("/assistant/materials/:id", requireRole("aslab"), h.DeleteMaterial)
		api.GET("/materials", requireRole("admin", "kalab", "laboran", "aslab"), h.GetMaterials)
		api.POST("/materials", requireRole("aslab"), h.CreateMaterial)
		api.POST("/assistant/materials", requireRole("aslab"), h.CreateMaterialMetadata)
		api.PUT("/materials/:id/submit", requireRole("aslab"), h.SubmitMaterial)
		api.PUT("/materials/:id/approve", requireRole("laboran", "kalab"), h.ApproveMaterial)
		api.PUT("/materials/:id/reject", requireRole("laboran", "kalab"), h.RejectMaterial)
		api.GET("/materials/:id/file", requireRole("admin", "kalab", "laboran", "aslab", "student"), h.MaterialFile)
		api.GET("/materials/:id/download", requireRole("admin", "kalab", "laboran", "aslab", "student"), h.DownloadMaterial)

		api.GET("/admin/courses", requireRole("admin", "laboran", "kalab"), h.GetAdminCourses)
		api.POST("/admin/courses", requireRole("laboran"), h.CreateCourse)
		api.PUT("/admin/courses/:id", requireRole("admin", "laboran", "kalab"), h.UpdateCourse)
		api.DELETE("/admin/courses/:id", requireRole("laboran"), h.DeleteCourse)

		api.GET("/admin/academic-years", requireRole("admin", "laboran", "kalab"), h.GetAcademicYears)
		api.POST("/admin/academic-years", requireRole("admin"), h.CreateAcademicYear)
		api.PUT("/admin/academic-years/:id", requireRole("admin"), h.UpdateAcademicYear)
		api.DELETE("/admin/academic-years/:id", requireRole("admin"), h.DeleteAcademicYear)

		api.GET("/admin/students", requireRole("admin", "laboran"), h.GetStudents)
		api.POST("/admin/students", requireRole("admin"), h.CreateStudent)
		api.GET("/admin/students/import/template", requireRole("admin"), h.DownloadStudentImportTemplate)
		api.POST("/admin/students/import/preview", requireRole("admin"), h.PreviewStudentImport)
		api.POST("/admin/students/import", requireRole("admin"), h.ImportStudents)
		api.PUT("/admin/students/:id", requireRole("admin"), h.UpdateStudent)
		api.DELETE("/admin/students/:id", requireRole("admin"), h.DeleteStudent)
		api.PUT("/admin/students/:id/reset-password", requireRole("admin"), h.ResetStudentPassword)

		api.GET("/admin/lecturers", requireRole("admin"), h.GetLecturers)
		api.POST("/admin/lecturers", requireRole("admin"), h.CreateLecturer)
		api.GET("/admin/lecturers/import/template", requireRole("admin"), h.DownloadLecturerImportTemplate)
		api.POST("/admin/lecturers/import/preview", requireRole("admin"), h.PreviewLecturerImport)
		api.POST("/admin/lecturers/import", requireRole("admin"), h.ImportLecturers)
		api.PUT("/admin/lecturers/:id", requireRole("admin"), h.UpdateLecturer)
		api.DELETE("/admin/lecturers/:id", requireRole("admin"), h.DeleteLecturer)
		api.PUT("/admin/lecturers/:id/reset-password", requireRole("admin"), h.ResetLecturerPassword)

		api.GET("/admin/lab-assistants", requireRole("admin"), h.GetAssistants)
		api.POST("/admin/lab-assistants", requireRole("admin"), h.CreateAssistant)
		api.GET("/admin/lab-assistants/import/template", requireRole("admin"), h.DownloadAssistantImportTemplate)
		api.POST("/admin/lab-assistants/import/preview", requireRole("admin"), h.PreviewAssistantImport)
		api.POST("/admin/lab-assistants/import", requireRole("admin"), h.ImportAssistants)
		api.PUT("/admin/lab-assistants/:id", requireRole("admin"), h.UpdateAssistant)
		api.DELETE("/admin/lab-assistants/:id", requireRole("admin"), h.DeleteAssistant)
		api.PUT("/admin/lab-assistants/:id/reset-password", requireRole("admin"), h.ResetAssistantPassword)

		api.GET("/admin/classes", requireRole("admin", "laboran"), h.GetClasses)
		api.POST("/admin/classes", requireRole("admin", "laboran"), h.CreateClass)
		api.PUT("/admin/classes/:id", requireRole("admin", "laboran"), h.UpdateClass)
		api.DELETE("/admin/classes/:id", requireRole("admin", "laboran"), h.DeleteClass)

		api.POST("/admin/assignments", requireRole("admin"), h.CreateAdminAssignment)
		api.PUT("/admin/assignments/:id", requireRole("admin"), h.UpdateAdminAssignment)
		api.DELETE("/admin/assignments/:id", requireRole("admin"), h.DeleteAdminAssignment)

		api.POST("/assistant/session-assignments", requireRole("aslab"), h.CreateSessionAssignment)
		api.PUT("/assistant/session-assignments/:id", requireRole("aslab"), h.UpdateSessionAssignment)
		api.DELETE("/assistant/session-assignments/:id", requireRole("aslab"), h.DeleteSessionAssignment)
		api.PUT("/assistant/reports/:id/review", requireRole("laboran", "kalab"), h.ReviewAssistantReport)
		api.GET("/reports", requireRole("admin", "kalab", "laboran", "aslab"), h.GetReports)
		api.GET("/reports/:id/file", requireRole("admin", "kalab", "laboran", "aslab"), h.ReportFile)
		api.GET("/reports/:id/html", requireRole("admin", "kalab", "laboran", "aslab"), h.ReportHTML)
		api.GET("/submissions/:id/download", requireRole("admin", "kalab", "laboran", "aslab"), h.ReportFile)
		api.PUT("/reports/:id/approve", requireRole("laboran", "kalab"), h.ApproveReport)
		api.PUT("/reports/:id/reject", requireRole("laboran", "kalab"), h.RejectReport)
		api.PUT("/assistant/attendance/sessions/:id", requireRole("aslab"), h.UpdateAssistantAttendanceSession)
		api.PUT("/assistant/course-sessions/:id/attendance", requireRole("aslab"), h.UpdateCourseSessionAttendance)
		api.GET("/assistant/sessions/:id/assessments", requireRole("aslab", "student"), h.GetSessionAssessments)
		api.PUT("/assistant/sessions/:id/assessments", requireRole("aslab"), h.UpsertSessionAssessment)
		api.PUT("/assistant/sessions/:id/student-assessments", requireRole("aslab"), h.UpsertAssistantStudentSessionAssessment)
		api.PUT("/student/sessions/:id/assessments", requireRole("student"), h.UpsertStudentSessionAssessment)
		api.GET("/assistant/sessions/:id/pretest", requireRole("aslab"), h.GetAssistantSessionPretest)
		api.POST("/assistant/sessions/:id/pretest/questions", requireRole("aslab"), h.CreatePretestQuestion)
		api.PUT("/assistant/sessions/:id/pretest/questions/:questionId", requireRole("aslab"), h.UpdatePretestQuestion)
		api.DELETE("/assistant/sessions/:id/pretest/questions/:questionId", requireRole("aslab"), h.DeletePretestQuestion)
		api.GET("/student/assignments/:id/submission", requireRole("student"), h.GetStudentAssignmentSubmission)
		api.GET("/student/sessions/:id/pretest", requireRole("student"), h.GetStudentSessionPretest)
		api.POST("/student/sessions/:id/pretest/submit", requireRole("student"), h.SubmitStudentPretest)
		// Rute quiz melayani pretest dan post-test lewat query string ?type=.
		// Rute /pretest di atas dipertahankan sebagai jalur lama.
		api.GET("/assistant/sessions/:id/quiz", requireRole("aslab"), h.GetAssistantSessionQuiz)
		api.POST("/assistant/sessions/:id/quiz/questions", requireRole("aslab"), h.CreateQuizQuestion)
		api.PUT("/assistant/sessions/:id/quiz/questions/:questionId", requireRole("aslab"), h.UpdateQuizQuestion)
		api.DELETE("/assistant/sessions/:id/quiz/questions/:questionId", requireRole("aslab"), h.DeleteQuizQuestion)
		api.GET("/student/sessions/:id/quiz", requireRole("student"), h.GetStudentSessionQuiz)
		api.POST("/student/sessions/:id/quiz/submit", requireRole("student"), h.SubmitStudentQuiz)
		api.GET("/sessions/:id/ngain", requireRole("admin", "kalab", "laboran", "aslab"), h.GetSessionNGain)
		api.GET("/courses/:id/ngain", requireRole("admin", "kalab", "laboran", "aslab"), h.GetCourseNGain)
		api.GET("/student/sessions/:id/ngain", requireRole("student"), h.GetStudentOwnSessionNGain)
		api.POST("/assistant/sessions/:id/reports", requireRole("aslab"), h.SubmitAssistantSessionReport)
		api.PUT("/lecturer/reports/:id/approve", requireRole("kalab"), h.ApproveAssistantReport)
		api.PUT("/lecturer/reports/:id/reject", requireRole("kalab"), h.RejectAssistantReport)
		api.GET("/reports/workflow", requireRole("aslab", "laboran", "kalab", "admin"), h.GetReportWorkflow)
		api.POST("/reports/workflow/submit", requireRole("aslab"), h.SubmitReportWorkflow)
		api.POST("/reports/workflow/approve", requireRole("laboran", "kalab"), h.ApproveReportWorkflow)
		api.POST("/reports/workflow/reject", requireRole("laboran", "kalab"), h.RejectReportWorkflow)
		api.POST("/reports/workflow/reset", requireRole("aslab"), h.ResetReportWorkflow)
		api.PUT("/assistant/submissions/:id/grade", requireRole("aslab", "laboran"), h.GradeStudentSubmission)
		api.PUT("/assistant/sessions/:id/student-attendance", requireRole("aslab"), h.UpdateStudentAttendanceByCourseSession)
		api.POST("/assistant/sessions/:id/assignments", requireRole("aslab"), h.CreateSessionAssignment)
		api.PUT("/assistant/assignments/:id", requireRole("aslab"), h.UpdateSessionAssignment)
		api.GET("/assistant/assignments/:id/submissions", requireRole("aslab", "laboran", "kalab"), h.GetAssignmentSubmissions)
		api.GET("/assistant/assignments/:id/students", requireRole("aslab"), h.GetSessionStudents)
		api.POST("/assistant/sessions/:id/assistant-attendance", requireRole("aslab"), h.UpsertAssistantSessionAttendance)
		api.GET("/assistant/sessions/:id/assistant-attendance", requireRole("aslab"), h.GetAssistantSessionAttendance)
		api.POST("/upload", requireRole("admin", "kalab", "laboran", "aslab", "student"), h.UploadFile)
		api.POST("/files/upload", requireRole("admin", "kalab", "laboran", "aslab", "student"), h.UploadFile)
		api.GET("/files/*filepath", requireRole("admin", "kalab", "laboran", "aslab", "student"), h.ServeFile)
		api.GET("/admin/courses/:id/sessions", requireRole("admin", "laboran", "kalab"), h.GetCourseSessions)
		api.PUT("/admin/sessions/:id", requireRole("admin", "laboran", "kalab"), h.UpdateCourseSession)
		api.GET("/admin/reports/export", requireRole("admin", "laboran", "kalab"), h.ExportReportsXLSX)
		api.GET("/assignments/:id/trace", requireRole("admin", "kalab", "laboran", "aslab"), h.GetAssignmentTraceFlow)
		api.GET("/assignments/:id/stats", requireRole("admin", "kalab", "laboran", "aslab"), h.GetAssignmentStats)
		api.GET("/assignments/:id/grade-impact", requireRole("admin", "kalab", "laboran", "aslab"), h.GetAssignmentGradeImpact)
		api.GET("/reports/:id/trace", requireRole("admin", "kalab", "laboran", "aslab"), h.GetReportTraceFlow)
		api.GET("/submissions/:id", requireRole("admin", "kalab", "laboran", "aslab", "student"), h.GetSubmissionDetail)
	}
}

func (h *Handler) Login(c *gin.Context) {
	var payload models.LoginRequest
	if bind(c, &payload) {
		return
	}

	user, err := h.repo.Login(payload.Email, payload.Password)
	if errors.Is(err, sql.ErrNoRows) {
		fail(c, http.StatusUnauthorized, "email or password is invalid")
		return
	}
	if err == nil {
		user.Token, err = issueAuthToken(user)
	}
	respond(c, user, err)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	authUser, ok := currentAuthUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "authentication required")
		return
	}
	var payload models.ChangePasswordRequest
	if bind(c, &payload) {
		return
	}
	if payload.NewPassword != payload.ConfirmPassword {
		fail(c, http.StatusBadRequest, "password baru dan konfirmasi tidak sama")
		return
	}
	respond(c, gin.H{"changed": true}, h.repo.ChangePassword(authUser.ID, authUser.Role, payload.CurrentPassword, payload.NewPassword))
}

func (h *Handler) GetDashboard(c *gin.Context) {
	data, err := h.repo.Dashboard()
	respond(c, data, err)
}
func (h *Handler) GetStudentCourses(c *gin.Context) {
	data, err := h.repo.StudentCourses()
	respond(c, data, err)
}
func (h *Handler) GetAssignments(c *gin.Context) {
	studentID, _ := strconv.Atoi(c.Query("studentId"))
	data, err := h.repo.Assignments(studentID)
	respond(c, data, err)
}
func (h *Handler) GetGrades(c *gin.Context) { data, err := h.repo.Grades(); respond(c, data, err) }
func (h *Handler) GetAdminCourses(c *gin.Context) {
	data, err := h.repo.AdminCourses()
	respond(c, data, err)
}
func (h *Handler) GetAcademicYears(c *gin.Context) {
	data, err := h.repo.AcademicYears()
	respond(c, data, err)
}
func (h *Handler) GetStudents(c *gin.Context) { data, err := h.repo.Students(); respond(c, data, err) }
func (h *Handler) GetLecturers(c *gin.Context) {
	data, err := h.repo.Lecturers()
	respond(c, data, err)
}
func (h *Handler) GetAssistants(c *gin.Context) {
	data, err := h.repo.Assistants()
	respond(c, data, err)
}
func (h *Handler) GetClasses(c *gin.Context) { data, err := h.repo.Classes(); respond(c, data, err) }
func (h *Handler) SubmitAssignment(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.AssignmentSubmission
	if bind(c, &payload) {
		return
	}
	err = h.repo.SubmitAssignment(id, &payload)
	payload.ID = id
	respond(c, payload, err)
}

func (h *Handler) SubmitStudentAssignment(c *gin.Context) {
	var payload models.StudentSubmissionInput
	if bind(c, &payload) {
		return
	}
	err := h.repo.SubmitStudentAssignment(payload.AssignmentID, payload.StudentID, payload.AnswerText, payload.FileURL, payload.FileName, payload.FileSize)
	respond(c, payload, err)
}

func (h *Handler) CreateCourse(c *gin.Context) {
	var payload models.AdminCourse
	if bind(c, &payload) {
		return
	}
	err := h.repo.CreateCourse(&payload)
	created(c, payload, err)
}
func (h *Handler) UpdateCourse(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.AdminCourse
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateCourse(id, &payload)
	respond(c, payload, err)
}
func (h *Handler) DeleteCourse(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteCourse(id)
	respond(c, gin.H{"id": id}, err)
}

func (h *Handler) CreateAcademicYear(c *gin.Context) {
	var payload models.AcademicYear
	if bind(c, &payload) {
		return
	}
	err := h.repo.CreateAcademicYear(&payload)
	created(c, payload, err)
}
func (h *Handler) UpdateAcademicYear(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.AcademicYear
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateAcademicYear(id, &payload)
	respond(c, payload, err)
}
func (h *Handler) DeleteAcademicYear(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteAcademicYear(id)
	respond(c, gin.H{"id": id}, err)
}

func (h *Handler) CreateStudent(c *gin.Context) {
	var payload models.Student
	if bind(c, &payload) {
		return
	}
	err := h.repo.CreateStudent(&payload)
	payload.Password = ""
	payload.DefaultPassword = ""
	created(c, payload, err)
}
func (h *Handler) UpdateStudent(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.Student
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateStudent(id, &payload)
	payload.Password = ""
	payload.DefaultPassword = ""
	respond(c, payload, err)
}
func (h *Handler) DeleteStudent(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteStudent(id)
	respond(c, gin.H{"id": id}, err)
}
func (h *Handler) ResetStudentPassword(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.ResetStudentPassword(id)
	respond(c, data, err)
}

func (h *Handler) CreateLecturer(c *gin.Context) {
	var payload models.Lecturer
	if bind(c, &payload) {
		return
	}
	err := h.repo.CreateLecturer(&payload)
	payload.Password = ""
	payload.DefaultPassword = ""
	created(c, payload, err)
}
func (h *Handler) UpdateLecturer(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.Lecturer
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateLecturer(id, &payload)
	payload.Password = ""
	payload.DefaultPassword = ""
	respond(c, payload, err)
}
func (h *Handler) DeleteLecturer(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteLecturer(id)
	respond(c, gin.H{"id": id}, err)
}
func (h *Handler) ResetLecturerPassword(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.ResetLecturerPassword(id)
	respond(c, data, err)
}

func (h *Handler) CreateAssistant(c *gin.Context) {
	var payload models.LabAssistant
	if bind(c, &payload) {
		return
	}
	err := h.repo.CreateAssistant(&payload)
	payload.Password = ""
	payload.DefaultPassword = ""
	created(c, payload, err)
}
func (h *Handler) UpdateAssistant(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.LabAssistant
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateAssistant(id, &payload)
	payload.Password = ""
	payload.DefaultPassword = ""
	respond(c, payload, err)
}
func (h *Handler) DeleteAssistant(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteAssistant(id)
	respond(c, gin.H{"id": id}, err)
}
func (h *Handler) ResetAssistantPassword(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.ResetAssistantPassword(id)
	respond(c, data, err)
}

func (h *Handler) CreateClass(c *gin.Context) {
	var payload models.ClassData
	if bind(c, &payload) {
		return
	}
	err := h.repo.CreateClass(&payload)
	created(c, payload, err)
}
func (h *Handler) UpdateClass(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.ClassData
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateClass(id, &payload)
	respond(c, payload, err)
}
func (h *Handler) DeleteClass(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteClass(id)
	respond(c, gin.H{"id": id}, err)
}

func (h *Handler) CreateLecturerCourse(c *gin.Context) {
	var payload models.LecturerCourse
	if bind(c, &payload) {
		return
	}
	err := h.repo.CreateLecturerCourse(&payload)
	created(c, payload, err)
}
func (h *Handler) UpdateLecturerCourse(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.LecturerCourse
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateLecturerCourse(id, &payload)
	payload.ID = id
	respond(c, payload, err)
}
func (h *Handler) DeleteLecturerCourse(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteLecturerCourse(id)
	respond(c, gin.H{"id": id}, err)
}

func (h *Handler) CreateAdminAssignment(c *gin.Context) {
	var payload models.AdminAssignment
	if bind(c, &payload) {
		return
	}
	err := h.repo.CreateAdminAssignment(&payload)
	created(c, payload, err)
}
func (h *Handler) UpdateAdminAssignment(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.AdminAssignment
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateAdminAssignment(id, &payload)
	payload.ID = id
	respond(c, payload, err)
}
func (h *Handler) DeleteAdminAssignment(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteAdminAssignment(id)
	respond(c, gin.H{"id": id}, err)
}

func (h *Handler) DeleteMaterial(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteMaterial(id)
	respond(c, gin.H{"id": id}, err)
}

func (h *Handler) GetMaterials(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	data, err := h.repo.MaterialsForRole(authUser.Role, authUser.ID)
	respond(c, data, err)
}

func (h *Handler) CreateMaterial(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		fail(c, http.StatusBadRequest, "judul materi wajib diisi")
		return
	}
	courseID, err := strconv.Atoi(c.PostForm("courseId"))
	if err != nil || courseID <= 0 {
		fail(c, http.StatusBadRequest, "courseId wajib valid")
		return
	}
	var sessionID *int
	if raw := strings.TrimSpace(c.PostForm("sessionId")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			fail(c, http.StatusBadRequest, "sessionId tidak valid")
			return
		}
		sessionID = &parsed
	}
	file, err := c.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "file PDF wajib diupload")
		return
	}
	if !strings.EqualFold(filepath.Ext(file.Filename), ".pdf") {
		fail(c, http.StatusBadRequest, "materi hanya menerima file PDF")
		return
	}
	relativePath, err := h.saveUploadedPDF(file.Filename, "materials", file)
	if err != nil {
		respond(c, nil, err)
		return
	}
	item, err := h.repo.CreateMaterial(courseID, sessionID, title, c.PostForm("description"), relativePath, humanFileSize(file.Size), authUser.ID)
	created(c, item, err)
}

func (h *Handler) SubmitMaterial(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	id, err := pathID(c)
	if err != nil {
		return
	}
	respond(c, gin.H{"id": id, "status": "submitted"}, h.repo.SubmitMaterial(id, authUser.ID))
}

func (h *Handler) ApproveMaterial(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	id, err := pathID(c)
	if err != nil {
		return
	}
	respond(c, gin.H{"id": id, "status": "approved"}, h.repo.ApproveMaterial(id, authUser.ID))
}

func (h *Handler) RejectMaterial(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.ReportActionRequest
	if bind(c, &payload) {
		return
	}
	respond(c, gin.H{"id": id, "status": "rejected"}, h.repo.RejectMaterial(id, authUser.ID, payload.Note))
}

func (h *Handler) CreateMaterialMetadata(c *gin.Context) {
	var payload models.CreateMaterialInput
	if bind(c, &payload) {
		return
	}
	id, err := h.repo.SaveMaterialFile(payload.CourseID, payload.SessionID, payload.Title, payload.FileURL, payload.FileType, payload.FileSize)
	created(c, gin.H{"id": id}, err)
}

func (h *Handler) MaterialFile(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	id, err := pathID(c)
	if err != nil {
		return
	}
	path, err := h.repo.MaterialFilePath(id, authUser.Role)
	if errors.Is(err, sql.ErrNoRows) {
		materials, listErr := h.repo.MaterialsForRole(authUser.Role, authUser.ID)
		if listErr != nil {
			respond(c, nil, listErr)
			return
		}
		for _, material := range materials {
			if material.ID == id {
				c.Header("Content-Type", "application/pdf")
				c.Header("Content-Disposition", `inline; filename="materi.pdf"`)
				c.Data(http.StatusOK, "application/pdf", simplePDFBytes(fmt.Sprintf("Materi %s\nKursus %s\nStatus %s", material.Title, material.CourseName, material.Status)))
				return
			}
		}
	}
	if err != nil {
		respond(c, nil, err)
		return
	}
	h.serveUploadFile(c, path)
}

func (h *Handler) DownloadMaterial(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var fileURL, title string
	err = h.repo.DB().QueryRow(`SELECT COALESCE(NULLIF(file_url,''),'/api/files/placeholder.pdf'), title FROM course_materials WHERE id=$1`, id).Scan(&fileURL, &title)
	if err != nil {
		fail(c, http.StatusNotFound, "material not found")
		return
	}
	if strings.HasPrefix(fileURL, "/api/files/") {
		fp := filepath.Base(strings.TrimPrefix(fileURL, "/api/files/"))
		fullPath := filepath.Join(uploadDir, fp)
		if _, err := os.Stat(fullPath); err == nil {
			ext := strings.ToLower(filepath.Ext(fp))
			mimeType := mime.TypeByExtension(ext)
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			c.Header("Content-Type", mimeType)
			c.Header("Content-Disposition", "inline; filename=\""+title+"\"")
			c.File(fullPath)
			return
		}
	}
	c.Redirect(http.StatusFound, fileURL)
}

func (h *Handler) GetReports(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	data, err := h.repo.AssistantReportsForRole(authUser.Role, authUser.ID)
	respond(c, data, err)
}

func (h *Handler) ReportFile(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	id, err := pathID(c)
	if err != nil {
		return
	}
	path, err := h.repo.ReportFilePath(id, authUser.Role)
	if errors.Is(err, sql.ErrNoRows) {
		reports, listErr := h.repo.AssistantReportsForRole(authUser.Role, authUser.ID)
		if listErr != nil {
			respond(c, nil, listErr)
			return
		}
		for _, report := range reports {
			if report.ID == id {
				doc, docErr := h.repo.ReportDocument(id, authUser.Role, authUser.ID)
				if docErr != nil {
					respond(c, nil, docErr)
					return
				}
				pdf, pdfErr := renderReportPDFWithDompdf(doc)
				if pdfErr != nil {
					respond(c, nil, pdfErr)
					return
				}
				c.Header("Content-Type", "application/pdf")
				c.Header("Content-Disposition", `attachment; filename="laporan-praktikum.pdf"`)
				c.Data(http.StatusOK, "application/pdf", pdf)
				return
			}
		}
	}
	if err != nil {
		respond(c, nil, err)
		return
	}
	h.serveUploadFile(c, path)
}

func (h *Handler) ReportHTML(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	id, err := pathID(c)
	if err != nil {
		return
	}
	doc, err := h.reportDocumentForUser(id, authUser)
	if err != nil {
		respond(c, nil, err)
		return
	}
	html, err := renderReportHTML(doc)
	if err != nil {
		respond(c, nil, err)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Content-Disposition", `inline; filename="laporan.html"`)
	c.String(http.StatusOK, html)
}

func (h *Handler) reportDocumentForUser(id int, authUser models.AuthUser) (*models.ReportDocument, error) {
	reports, err := h.repo.AssistantReportsForRole(authUser.Role, authUser.ID)
	if err != nil {
		return nil, err
	}
	for _, report := range reports {
		if report.ID == id {
			return h.repo.ReportDocument(id, authUser.Role, authUser.ID)
		}
	}
	return nil, sql.ErrNoRows
}

func (h *Handler) ApproveReport(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	id, err := pathID(c)
	if err != nil {
		return
	}
	respond(c, gin.H{"id": id}, h.repo.ApproveReportByRole(id, authUser.Role, authUser.ID))
}

func (h *Handler) RejectReport(c *gin.Context) {
	authUser, _ := currentAuthUser(c)
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.ReportActionRequest
	if bind(c, &payload) {
		return
	}
	if strings.TrimSpace(payload.Note) == "" {
		fail(c, http.StatusBadRequest, "catatan penolakan wajib diisi")
		return
	}
	respond(c, gin.H{"id": id}, h.repo.RejectReportByRole(id, authUser.Role, authUser.ID, payload.Note))
}

func (h *Handler) saveUploadedPDF(originalName string, folder string, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > maxUploadSize {
		return "", fmt.Errorf("ukuran file maksimal 20 MB")
	}
	// Validate PDF magic bytes
	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil || !bytes.HasPrefix(magic, []byte("%PDF")) {
		return "", fmt.Errorf("file bukan PDF yang valid")
	}

	dir := filepath.Join(uploadDir, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d-%s", time.Now().UnixNano(), safeFileName(originalName))
	relativePath := filepath.Join(folder, name)
	fullPath := filepath.Join(uploadDir, relativePath)
	return relativePath, saveUploadedFile(fileHeader, fullPath)
}

func (h *Handler) serveUploadFile(c *gin.Context, relativePath string) {
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `inline; filename="file.pdf"`)
	c.File(filepath.Join(uploadDir, relativePath))
}

func safeFileName(name string) string {
	clean := filepath.Base(name)
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, clean)
}

func humanFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

func saveUploadedFile(fileHeader *multipart.FileHeader, path string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func reportPDFBytes(doc *models.ReportDocument) []byte {
	report := doc.Report
	institution := doc.Institution

	lines := []string{
		fmt.Sprintf("PERIODE : %s %s", strings.ToUpper(emptyDash(doc.AcademicYear)), strings.ToUpper(emptyDash(doc.Semester))),
		"",
		fmt.Sprintf("Mata kuliah      : %s", report.CourseName),
		fmt.Sprintf("Dosen Pengajar   : %s", emptyDash(doc.Instructor)),
		fmt.Sprintf("Kelas / Kelompok : %s", report.Class),
		fmt.Sprintf("Kode Mata kuliah : %s", report.CourseCode),
		fmt.Sprintf("Nama Kelas       : %s", report.Class),
		fmt.Sprintf("SKS              : %d", doc.Credits),
		"",
		"PRESENSI MAHASISWA",
		"No  NPM             Nama                           Pertemuan Alfa Hadir Ijin Sakit Presentase",
	}
	for _, student := range doc.Students {
		lines = append(lines, fmt.Sprintf("%-3d %-15s %-30s %9d %4d %5d %4d %5d %9.2f",
			student.No, truncatePDFText(student.NIM, 15), truncatePDFText(student.Name, 30),
			student.Meetings, student.Absent, student.Present, student.Permit, student.Sick, student.AttendancePercent))
	}
	lines = append(lines,
		"",
		"NILAI AKHIR MAHASISWA",
		"No  NPM             Nama Mahasiswa                 Tugas(30%) UAS(35%) Praktikum(35%) Nilai Grade Lulus",
	)
	for _, student := range doc.Students {
		lulus := ""
		if student.Passed {
			lulus = "Lulus"
		}
		lines = append(lines, fmt.Sprintf("%-3d %-15s %-30s %10.2f %8.2f %14.2f %6.2f %-5s %s",
			student.No, truncatePDFText(student.NIM, 15), truncatePDFText(student.Name, 30),
			student.AssignmentScore, student.Posttest, student.Praktikum, student.FinalScore, student.Grade, lulus))
	}

	if strings.TrimSpace(report.RejectionNote) != "" {
		lines = append(lines, "", "CATATAN PENOLAKAN", report.RejectionNote)
	}
	studyProgramSigner := signerByRole(doc.Signers, "head_of_study_program")
	labSigner := signerByRole(doc.Signers, "head_of_laboratory")
	lines = append(lines,
		"",
		fmt.Sprintf("Jakarta, %s", time.Now().Format("02 January 2006")),
		"",
		"Mengetahui,",
		fmt.Sprintf("Ka. Program Studi %-25s Ka. Laboratorium %s", truncatePDFText(emptyDash(institution.StudyProgramName), 25), emptyDash(institution.LaboratoryName)),
		"",
		"",
		"",
		fmt.Sprintf("%-40s %s", emptyDash(studyProgramSigner.Name), emptyDash(labSigner.Name)),
		fmt.Sprintf("%s: %-34s %s: %s", emptyDash(studyProgramSigner.IdentifierType), emptyDash(studyProgramSigner.IdentifierNumber), emptyDash(labSigner.IdentifierType), emptyDash(labSigner.IdentifierNumber)),
	)
	return simplePDFBytes(strings.Join(lines, "\n"))
}

type reportHTMLView struct {
	Doc          *models.ReportDocument
	Passed       int
	Failed       int
	PassRate     float64
	AverageScore float64
	GeneratedAt  string
	Kaprodi      models.ReportSigner
	Kalab        models.ReportSigner
}

func renderReportHTML(doc *models.ReportDocument) (string, error) {
	view := reportHTMLView{
		Doc:         doc,
		GeneratedAt: time.Now().Format("2006-01-02"),
		Kaprodi:     signerByRole(doc.Signers, "head_of_study_program"),
		Kalab:       signerByRole(doc.Signers, "head_of_laboratory"),
	}
	for _, student := range doc.Students {
		if student.Passed {
			view.Passed++
		}
		view.AverageScore += student.FinalScore
	}
	view.Failed = len(doc.Students) - view.Passed
	if len(doc.Students) > 0 {
		view.PassRate = float64(view.Passed) * 100 / float64(len(doc.Students))
		view.AverageScore = view.AverageScore / float64(len(doc.Students))
	}

	tpl, err := template.New("report").Funcs(template.FuncMap{
		"dash":  emptyDash,
		"upper": strings.ToUpper,
		"status": func(passed bool) string {
			if passed {
				return "Lulus"
			}
			return "Tidak Lulus"
		},
		"done": statusDone,
		// ngain menghitung N-Gain satu mahasiswa, ngainSummary merangkum sekelas.
		"ngain": reportNGain,
		"ngainSummary": func(students []models.ReportStudent) repositories.NGainSummary {
			values := make([]repositories.NGainValue, 0, len(students))
			for _, student := range students {
				values = append(values, reportNGain(student.Pretest, student.Posttest))
			}
			return repositories.SummarizeNGainValues(values, 100)
		},
		"minus": func(a float64, b float64) float64 { return a - b },
		"add":   func(a int, b int) int { return a + b },
	}).Parse(reportHTMLTemplate)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, view); err != nil {
		return "", err
	}
	return out.String(), nil
}

func renderReportPDFWithDompdf(doc *models.ReportDocument) ([]byte, error) {
	html, err := renderReportHTML(doc)
	if err != nil {
		return nil, err
	}
	scriptPath := filepath.Join("scripts", "render_report_pdf.php")
	command := exec.Command("php", scriptPath)
	command.Stdin = strings.NewReader(html)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("render dompdf: %s", message)
	}
	if len(output) == 0 {
		return nil, errors.New("render dompdf: empty pdf output")
	}
	return output, nil
}

const reportHTMLTemplate = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8" />
  <title>{{.Doc.Report.CourseName}} - Laporan Praktikum</title>
  <style>
    body{font-family:Arial,sans-serif;color:#111827;margin:32px;line-height:1.35}
    .cover{text-align:center;margin-bottom:28px}
    h1{font-size:24px;margin:0 0 8px;text-transform:uppercase}
    h2{font-size:16px;margin:24px 0 8px;text-transform:uppercase}
    h3{font-size:14px;margin:18px 0 8px}
    .muted{color:#4b5563;font-size:12px}
    table{width:100%;border-collapse:collapse;margin:8px 0 18px;font-size:11px}
    th,td{border:1px solid #d1d5db;padding:6px;vertical-align:top}
    th{background:#f3f4f6;text-align:left}
    .grid{display:grid;grid-template-columns:1fr 1fr;gap:24px}
    .no-border td{border:0;padding:3px 0}
    .signatures{display:grid;grid-template-columns:1fr 1fr;gap:48px;margin-top:32px;text-align:center}
    .signature-space{height:70px}
    .ngain-formula{font-size:11px;color:#374151;background:#f9fafb;border-left:3px solid #9ca3af;padding:8px 10px;margin:6px 0 12px}
    .report-head{margin-bottom:16px}
    .period{text-align:center;font-weight:bold;font-size:14px;text-transform:uppercase;margin-bottom:10px}
    .head-info{font-size:12px}
    .head-info td{border:0;padding:2px 6px 2px 0}
    .head-info .rlabel{padding-left:24px}
    td.c,th{text-align:center}
    td{text-align:left}
    @media print{body{margin:18mm}.page-break{page-break-before:always}}
  </style>
</head>
<body>
  <section class="report-head">
    <div class="period">PERIODE : {{upper (dash .Doc.AcademicYear)}} {{upper (dash .Doc.Semester)}}</div>
    <table class="no-border head-info">
      <tr><td>Mata kuliah</td><td>: {{dash .Doc.Report.CourseName}}</td><td class="rlabel">Nama Kelas</td><td>: {{dash .Doc.Report.Class}}</td></tr>
      <tr><td>Dosen Pengajar</td><td>: {{dash .Doc.Instructor}}</td><td class="rlabel">Kode Mata kuliah</td><td>: {{dash .Doc.Report.CourseCode}}</td></tr>
      <tr><td>Kelas / Kelompok</td><td>: {{dash .Doc.Report.Class}}</td><td class="rlabel">SKS</td><td>: {{.Doc.Credits}}</td></tr>
    </table>
  </section>

  <section>
    <h3>Presensi Mahasiswa</h3>
    <table>
      <thead><tr><th>No</th><th>NPM</th><th>Nama</th><th>Pertemuan</th><th>Alfa</th><th>Hadir</th><th>Ijin</th><th>Sakit</th><th>Presentase</th></tr></thead>
      <tbody>{{range .Doc.Students}}<tr><td>{{.No}}</td><td>{{.NIM}}</td><td>{{.Name}}</td><td class="c">{{.Meetings}}</td><td class="c">{{if gt .Absent 0}}{{.Absent}}{{end}}</td><td class="c">{{if gt .Present 0}}{{.Present}}{{end}}</td><td class="c">{{if gt .Permit 0}}{{.Permit}}{{end}}</td><td class="c">{{if gt .Sick 0}}{{.Sick}}{{end}}</td><td class="c">{{printf "%.2f" .AttendancePercent}}</td></tr>{{end}}</tbody>
    </table>
  </section>

  <section class="page-break">
    <h3>Nilai Akhir Mahasiswa</h3>
    <table>
      <thead><tr><th>No</th><th>NPM</th><th>Nama Mahasiswa</th><th>TUGAS (30%)</th><th>UJIAN AKHIR SEMESTER (35%)</th><th>PRAKTIKUM (35%)</th><th>Nilai</th><th>Grade</th><th>Lulus</th><th>Sunting KRS?</th><th>Info</th></tr></thead>
      <tbody>{{range .Doc.Students}}<tr><td>{{.No}}</td><td>{{.NIM}}</td><td>{{.Name}}</td><td class="c">{{printf "%.2f" .AssignmentScore}}</td><td class="c">{{printf "%.2f" .Posttest}}</td><td class="c">{{printf "%.2f" .Praktikum}}</td><td class="c">{{printf "%.2f" .FinalScore}}</td><td class="c">{{.Grade}}</td><td class="c">{{if .Passed}}&#10003;{{end}}</td><td></td><td></td></tr>{{end}}</tbody>
    </table>
  </section>

  {{if .Doc.Report.RejectionNote}}<section><h2>Catatan Penolakan</h2><p>{{.Doc.Report.RejectionNote}}</p></section>{{end}}

  <section class="signatures">
    <div><p>Ka. Program Studi {{dash .Doc.Institution.StudyProgramName}}</p><div class="signature-space"></div><strong>{{dash .Kaprodi.Name}}</strong><p>{{dash .Kaprodi.IdentifierType}}: {{dash .Kaprodi.IdentifierNumber}}</p></div>
    <div><p>Ka. Laboratorium {{dash .Doc.Institution.LaboratoryName}}</p><div class="signature-space"></div><strong>{{dash .Kalab.Name}}</strong><p>{{dash .Kalab.IdentifierType}}: {{dash .Kalab.IdentifierNumber}}</p></div>
  </section>
</body>
</html>`

func statusDone(done bool) string {
	if done {
		return "Selesai"
	}
	return "Belum"
}

// reportNGain menghitung N-Gain satu baris laporan. Nilai pretest dan post-test
// yang sama-sama nol berarti mahasiswa belum mengerjakan tes apa pun, bukan
// benar-benar mendapat nol, jadi barisnya ditandai belum lengkap.
func reportNGain(pre float64, post float64) repositories.NGainValue {
	if pre <= 0 && post <= 0 {
		return repositories.ComputeNGain(nil, nil, 100)
	}
	return repositories.ComputeNGain(&pre, &post, 100)
}

func signerByRole(signers []models.ReportSigner, role string) models.ReportSigner {
	for _, signer := range signers {
		if strings.EqualFold(signer.Role, role) {
			return signer
		}
	}
	return models.ReportSigner{IdentifierType: "NIDN"}
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func truncatePDFText(value string, max int) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func simplePDFBytes(text string) []byte {
	escaped := strings.ReplaceAll(text, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "(", "\\(")
	escaped = strings.ReplaceAll(escaped, ")", "\\)")
	lines := strings.Split(escaped, "\n")
	linesPerPage := 42
	contents := []string{}
	for start := 0; start < len(lines); start += linesPerPage {
		end := start + linesPerPage
		if end > len(lines) {
			end = len(lines)
		}
		content := "BT /F1 9 Tf 35 800 Td "
		for index, line := range lines[start:end] {
			if index > 0 {
				content += "0 -15 Td "
			}
			content += "(" + line + ") Tj "
		}
		content += "ET"
		contents = append(contents, content)
	}

	kids := []string{}
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}
	for index, content := range contents {
		pageObjectID := 4 + index*2
		contentObjectID := pageObjectID + 1
		kids = append(kids, fmt.Sprintf("%d 0 R", pageObjectID))
		objects = append(objects,
			fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>", contentObjectID),
			fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
		)
	}
	objects = append(objects[:1], append([]string{fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(contents))}, objects[1:]...)...)
	pdf := "%PDF-1.4\n"
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, len(pdf))
		pdf += fmt.Sprintf("%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xrefOffset := len(pdf)
	pdf += fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for i := 1; i < len(offsets); i++ {
		pdf += fmt.Sprintf("%010d 00000 n \n", offsets[i])
	}
	pdf += fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefOffset)
	return []byte(pdf)
}

func (h *Handler) CreateSessionAssignment(c *gin.Context) {
	var payload models.SessionAssignment
	if bind(c, &payload) {
		return
	}
	if payload.SessionID == 0 {
		if raw := c.Param("id"); raw != "" {
			sessionID, err := strconv.Atoi(raw)
			if err != nil {
				fail(c, http.StatusBadRequest, "invalid id")
				return
			}
			payload.SessionID = sessionID
		}
	}
	if payload.Deadline == "" {
		payload.Deadline = payload.DueDate
	}
	err := h.repo.CreateSessionAssignment(&payload)
	created(c, payload, err)
}

func (h *Handler) UpdateSessionAssignment(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.SessionAssignment
	if bind(c, &payload) {
		return
	}
	if payload.Deadline == "" {
		payload.Deadline = payload.DueDate
	}
	err = h.repo.UpdateSessionAssignment(id, &payload)
	respond(c, payload, err)
}

func (h *Handler) DeleteSessionAssignment(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	err = h.repo.DeleteSessionAssignment(id)
	respond(c, gin.H{"id": id}, err)
}
func (h *Handler) UpdateStudentGrade(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.StudentGradeUpdate
	if bind(c, &payload) {
		return
	}
	err = h.repo.UpdateStudentGrade(id, &payload)
	respond(c, payload, err)
}
func (h *Handler) ReviewAssistantReport(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.ReportReview
	if bind(c, &payload) {
		return
	}
	err = h.repo.ReviewAssistantReport(id, &payload)
	payload.ID = id
	respond(c, payload, err)
}
func (h *Handler) SubmitAssistantSessionReport(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	report, err := h.repo.SubmitAssistantSessionReport(id)
	respond(c, report, err)
}
func (h *Handler) ApproveAssistantReport(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}

	// Get lecturer name from header
	lecturerName := c.GetHeader("X-Lecturer-Name")
	if lecturerName == "" {
		fail(c, http.StatusUnauthorized, "lecturer name not provided")
		return
	}

	// Get report with course info
	report, err := h.repo.GetAssistantReportWithCourse(id)
	if err != nil {
		fail(c, http.StatusNotFound, "report not found")
		return
	}

	// Get course lecturer
	courseLecturer, err := h.repo.GetCourseLecturer(report["courseCode"].(string))
	if err != nil {
		fail(c, http.StatusInternalServerError, "failed to get course lecturer")
		return
	}

	// Check if current lecturer is the course lecturer
	if lecturerName != courseLecturer {
		fail(c, http.StatusForbidden, "only the course lecturer can approve this report")
		return
	}

	respond(c, gin.H{"id": id}, h.repo.SetAssistantReportStatus(id, "Disetujui"))
}
func (h *Handler) RejectAssistantReport(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}

	// Get lecturer name from header
	lecturerName := c.GetHeader("X-Lecturer-Name")
	if lecturerName == "" {
		fail(c, http.StatusUnauthorized, "lecturer name not provided")
		return
	}

	// Get report with course info
	report, err := h.repo.GetAssistantReportWithCourse(id)
	if err != nil {
		fail(c, http.StatusNotFound, "report not found")
		return
	}

	// Get course lecturer
	courseLecturer, err := h.repo.GetCourseLecturer(report["courseCode"].(string))
	if err != nil {
		fail(c, http.StatusInternalServerError, "failed to get course lecturer")
		return
	}

	// Check if current lecturer is the course lecturer
	if lecturerName != courseLecturer {
		fail(c, http.StatusForbidden, "only the course lecturer can reject this report")
		return
	}

	var payload models.RejectReportInput
	_ = c.ShouldBindJSON(&payload)
	respond(c, gin.H{"id": id}, h.repo.RejectAssistantReportWithNote(id, payload.RejectionNote))
}

func (h *Handler) GetAssignmentSubmissions(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetAssignmentSubmissions(id)
	respond(c, data, err)
}

func (h *Handler) GetSessionStudents(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetSessionStudents(id)
	respond(c, data, err)
}

func (h *Handler) GradeStudentSubmission(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload struct {
		Score    int    `json:"score"`
		Feedback string `json:"feedback"`
	}
	if bind(c, &payload) {
		return
	}
	fmt.Printf("[GradeStudentSubmission] Saving grade - submissionID: %d, score: %d, feedback: %s\n", id, payload.Score, payload.Feedback)
	err = h.repo.GradeStudentSubmission(id, payload.Score, payload.Feedback)
	if err == nil {
		fmt.Printf("[GradeStudentSubmission] Syncing to report - submissionID: %d\n", id)
		syncErr := h.repo.SyncSubmissionGradeToReport(id)
		if syncErr != nil {
			fmt.Printf("[GradeStudentSubmission] Sync error: %v\n", syncErr)
		} else {
			fmt.Printf("[GradeStudentSubmission] Sync success\n")
		}
	}
	respond(c, gin.H{"id": id, "score": payload.Score}, err)
}

func (h *Handler) UpsertAssistantSessionAttendance(c *gin.Context) {
	sessionID, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.AssistantSessionAttendance
	if bind(c, &payload) {
		return
	}
	respond(c, payload, h.repo.UpsertAssistantSessionAttendance(sessionID, &payload))
}

func (h *Handler) GetAssistantSessionAttendance(c *gin.Context) {
	sessionID, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetAssistantSessionAttendance(sessionID)
	respond(c, data, err)
}

func (h *Handler) GetCourseSessions(c *gin.Context) {
	courseID, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetCourseSessions(courseID)
	respond(c, data, err)
}

func (h *Handler) UpdateCourseSession(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.UpdateSessionInput
	if bind(c, &payload) {
		return
	}
	respond(c, gin.H{"id": id}, h.repo.UpdateCourseSession(id, &payload))
}

func (h *Handler) ExportReportsXLSX(c *gin.Context) {
	rows, err := h.repo.ExportReportsData()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	f := excelize.NewFile()
	sheet := "Laporan"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"No", "Kode Kursus", "Nama Kursus", "Kelas", "Sesi", "Topik", "NIM Aslab", "Nama Aslab", "Tanggal Kirim", "Status", "Nilai", "Catatan Penolakan"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, r := range rows {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), i+1)
		f.SetCellValue(sheet, cellName(2, row), r["courseCode"])
		f.SetCellValue(sheet, cellName(3, row), r["courseName"])
		f.SetCellValue(sheet, cellName(4, row), r["class"])
		f.SetCellValue(sheet, cellName(5, row), r["week"])
		f.SetCellValue(sheet, cellName(6, row), r["topic"])
		f.SetCellValue(sheet, cellName(7, row), r["nim"])
		f.SetCellValue(sheet, cellName(8, row), r["name"])
		f.SetCellValue(sheet, cellName(9, row), r["submittedAt"])
		f.SetCellValue(sheet, cellName(10, row), r["status"])
		f.SetCellValue(sheet, cellName(11, row), r["score"])
		f.SetCellValue(sheet, cellName(12, row), r["rejectionNote"])
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\"laporan-"+time.Now().Format("2006-01-02")+".xlsx\"")
	f.Write(c.Writer)
}

func cellName(col, row int) string {
	n, _ := excelize.CoordinatesToCellName(col, row)
	return n
}

var uploadDir = resolveUploadDir()

func resolveUploadDir() string {
	for _, dir := range []string{"./uploads", "/tmp/studiku-uploads"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			continue
		}
		// Verify we can actually write to the directory
		tmp, err := os.CreateTemp(dir, ".write-check-*")
		if err != nil {
			continue
		}
		tmp.Close()
		os.Remove(tmp.Name())
		return dir
	}
	return "/tmp/studiku-uploads" // last resort
}

const maxUploadSize = 20 * 1024 * 1024 // 20 MB

var allowedUploadExts = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

func (h *Handler) UploadFile(c *gin.Context) {
	if c.Request.ContentLength > maxUploadSize {
		fail(c, http.StatusRequestEntityTooLarge, "ukuran file maksimal 20 MB")
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "file wajib diupload")
		return
	}
	defer file.Close()

	if header.Size > maxUploadSize {
		fail(c, http.StatusRequestEntityTooLarge, "ukuran file maksimal 20 MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(filepath.Base(header.Filename)))
	if !allowedUploadExts[ext] {
		fail(c, http.StatusBadRequest, "tipe file tidak diizinkan")
		return
	}

	// Read magic bytes to validate actual file type
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	buf = buf[:n]
	detected := http.DetectContentType(buf)
	if ext == ".pdf" && !strings.HasPrefix(detected, "application/pdf") && !bytes.HasPrefix(buf, []byte("%PDF")) {
		fail(c, http.StatusBadRequest, "file bukan PDF yang valid")
		return
	}
	// Seek back to start
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fail(c, http.StatusInternalServerError, "cannot create upload dir")
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		fail(c, http.StatusInternalServerError, "cannot save file")
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		os.Remove(dst)
		fail(c, http.StatusInternalServerError, "cannot write file")
		return
	}

	fileURL := "/api/files/" + filename
	size := fmt.Sprintf("%.1f MB", float64(header.Size)/1024/1024)
	if header.Size < 1024*1024 {
		size = fmt.Sprintf("%.0f KB", float64(header.Size)/1024)
	}

	ok(c, "uploaded", gin.H{
		"fileUrl":  fileURL,
		"fileName": header.Filename,
		"fileSize": size,
		"mimeType": detected,
	})
}

func (h *Handler) ServeFile(c *gin.Context) {
	fp := c.Param("filepath")
	// Allow only safe filenames — no path separators
	fp = filepath.Base(fp)
	if fp == "." || fp == "/" {
		fail(c, http.StatusBadRequest, "invalid file")
		return
	}
	fullPath := filepath.Join(uploadDir, fp)
	// Confirm resolved path is still within uploadDir
	absUpload, _ := filepath.Abs(uploadDir)
	absFile, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absFile, absUpload+string(os.PathSeparator)) {
		fail(c, http.StatusForbidden, "akses ditolak")
		return
	}

	ext := strings.ToLower(filepath.Ext(fp))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	if _, err := os.Stat(fullPath); err != nil {
		fail(c, http.StatusNotFound, "file tidak ditemukan")
		return
	}

	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", `inline; filename="`+safeFileName(fp)+`"`)
	c.File(fullPath)
}
func (h *Handler) GetReportWorkflow(c *gin.Context) {
	data, err := h.repo.ReportWorkflow()
	respond(c, data, err)
}
func (h *Handler) SubmitReportWorkflow(c *gin.Context) {
	var payload models.ReportWorkflowAction
	if bind(c, &payload) {
		return
	}
	err := h.repo.UpsertReportWorkflow(payload.CourseID, "SUBMITTED")
	respond(c, payload, err)
}
func (h *Handler) ApproveReportWorkflow(c *gin.Context) {
	var payload models.ReportWorkflowAction
	if bind(c, &payload) {
		return
	}
	err := h.repo.UpsertReportWorkflow(payload.CourseID, "APPROVED")
	respond(c, payload, err)
}
func (h *Handler) RejectReportWorkflow(c *gin.Context) {
	var payload models.ReportWorkflowAction
	if bind(c, &payload) {
		return
	}
	err := h.repo.UpsertReportWorkflow(payload.CourseID, "REJECTED")
	respond(c, payload, err)
}
func (h *Handler) ResetReportWorkflow(c *gin.Context) {
	var payload models.ReportWorkflowAction
	if bind(c, &payload) {
		return
	}
	err := h.repo.UpsertReportWorkflow(payload.CourseID, "DRAFT")
	respond(c, payload, err)
}

func (h *Handler) UpdateAssistantAttendanceSession(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.AssistantAttendanceUpdate
	if bind(c, &payload) {
		return
	}
	respond(c, payload, h.repo.UpdateAssistantAttendanceSession(id, &payload))
}

func (h *Handler) UpdateCourseSessionAttendance(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.AssistantAttendanceUpdate
	if bind(c, &payload) {
		return
	}
	respond(c, payload, h.repo.UpdateCourseSessionAttendance(id, &payload))
}

func (h *Handler) UpdateStudentAttendanceByCourseSession(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.AssistantAttendanceUpdate
	if bind(c, &payload) {
		return
	}
	respond(c, payload, h.repo.UpsertStudentAttendanceByCourseSession(id, &payload))
}

func (h *Handler) GetSessionAssessments(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.SessionAssessments(id)
	respond(c, data, err)
}

func (h *Handler) UpsertSessionAssessment(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.SessionAssessmentInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.UpsertSessionAssessment(id, &payload)
	respond(c, data, err)
}

func (h *Handler) UpsertStudentSessionAssessment(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	var payload models.StudentSessionAssessmentInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.UpsertStudentSessionAssessment(id, authUser.ID, &payload)
	respond(c, data, err)
}

func (h *Handler) GetAssistantSessionPretest(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	data, err := h.repo.SessionPretest(id, authUser.ID, authUser.Role)
	respond(c, data, err)
}

func (h *Handler) GetStudentSessionPretest(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	data, err := h.repo.SessionPretest(id, authUser.ID, authUser.Role)
	respond(c, data, err)
}

func (h *Handler) CreatePretestQuestion(c *gin.Context) {
	sessionID, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	var payload models.PretestQuestionInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.CreatePretestQuestion(sessionID, authUser.ID, &payload)
	respond(c, data, err)
}

func (h *Handler) UpdatePretestQuestion(c *gin.Context) {
	_, err := pathID(c)
	if err != nil {
		return
	}
	questionID, err := pathParamID(c, "questionId")
	if err != nil {
		return
	}
	var payload models.PretestQuestionInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.UpdatePretestQuestion(questionID, &payload)
	respond(c, data, err)
}

func (h *Handler) DeletePretestQuestion(c *gin.Context) {
	_, err := pathID(c)
	if err != nil {
		return
	}
	questionID, err := pathParamID(c, "questionId")
	if err != nil {
		return
	}
	err = h.repo.DeletePretestQuestion(questionID)
	respond(c, gin.H{"id": questionID}, err)
}

func (h *Handler) SubmitStudentPretest(c *gin.Context) {
	sessionID, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	var payload models.PretestSubmissionInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.SubmitPretest(sessionID, authUser.ID, &payload)
	respond(c, data, err)
}

// quizTypeParam mengambil jenis tes dari query string, dengan isi body sebagai
// cadangan. Kosong berarti pretest, ditentukan di lapisan repository.
func quizTypeParam(c *gin.Context, fallback string) string {
	if value := c.Query("type"); value != "" {
		return value
	}
	return fallback
}

func (h *Handler) GetAssistantSessionQuiz(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	data, err := h.repo.SessionQuiz(id, authUser.ID, authUser.Role, quizTypeParam(c, ""))
	respond(c, data, err)
}

func (h *Handler) GetStudentSessionQuiz(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	data, err := h.repo.SessionQuiz(id, authUser.ID, authUser.Role, quizTypeParam(c, ""))
	respond(c, data, err)
}

func (h *Handler) CreateQuizQuestion(c *gin.Context) {
	sessionID, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	var payload models.QuizQuestionInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.CreateQuizQuestion(sessionID, authUser.ID, quizTypeParam(c, payload.Type), &payload)
	respond(c, data, err)
}

func (h *Handler) UpdateQuizQuestion(c *gin.Context) {
	if _, err := pathID(c); err != nil {
		return
	}
	questionID, err := pathParamID(c, "questionId")
	if err != nil {
		return
	}
	var payload models.QuizQuestionInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.UpdateQuizQuestion(questionID, quizTypeParam(c, payload.Type), &payload)
	respond(c, data, err)
}

func (h *Handler) DeleteQuizQuestion(c *gin.Context) {
	if _, err := pathID(c); err != nil {
		return
	}
	questionID, err := pathParamID(c, "questionId")
	if err != nil {
		return
	}
	err = h.repo.DeleteQuizQuestion(questionID)
	respond(c, gin.H{"id": questionID}, err)
}

func (h *Handler) SubmitStudentQuiz(c *gin.Context) {
	sessionID, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	var payload models.QuizSubmissionInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.SubmitQuiz(sessionID, authUser.ID, quizTypeParam(c, payload.Type), &payload)
	respond(c, data, err)
}

func (h *Handler) GetSessionNGain(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.SessionNGain(id)
	respond(c, data, err)
}

func (h *Handler) GetCourseNGain(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.CourseNGain(id)
	respond(c, data, err)
}

// GetStudentOwnSessionNGain hanya mengembalikan N-Gain milik mahasiswa yang login,
// tanpa membocorkan nilai teman sekelasnya.
func (h *Handler) GetStudentOwnSessionNGain(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	value, err := h.repo.StudentSessionNGain(id, authUser.ID)
	if err != nil {
		respond(c, nil, err)
		return
	}
	respond(c, gin.H{
		"sessionId": id,
		"ngain":     value,
		"guide":     repositories.NGainGuide(),
	}, nil)
}

func (h *Handler) UpsertAssistantStudentSessionAssessment(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	var payload models.StudentSessionAssessmentInput
	if bind(c, &payload) {
		return
	}
	data, err := h.repo.UpsertStudentSessionAssessment(id, payload.StudentID, &payload)
	respond(c, data, err)
}

func bind(c *gin.Context, out interface{}) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return true
	}
	return false
}
func pathID(c *gin.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid id")
		return 0, err
	}
	return id, nil
}

func pathParamID(c *gin.Context, name string) (int, error) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid id")
		return 0, err
	}
	return id, nil
}
func respond(c *gin.Context, data interface{}, err error) {
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, "success", data)
}
func created(c *gin.Context, data interface{}, err error) {
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, models.APIResponse{Success: true, Message: "created", Data: data})
}
func ok(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: msg, Data: data})
}
func fail(c *gin.Context, status int, message string) {
	c.JSON(status, models.APIResponse{Success: false, Error: message})
}

// GetAssignmentTraceFlow returns trace flow of assignment submissions
func (h *Handler) GetAssignmentTraceFlow(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetAssignmentTraceFlow(id)
	respond(c, data, err)
}

// GetAssignmentStats returns statistics for an assignment
func (h *Handler) GetAssignmentStats(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetAssignmentStats(id)
	respond(c, data, err)
}

// GetAssignmentGradeImpact returns impact of assignment scores on student grades
func (h *Handler) GetAssignmentGradeImpact(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetAssignmentGradeImpact(id)
	respond(c, data, err)
}

// GetReportTraceFlow returns trace flow of assistant reports with grade impact
func (h *Handler) GetReportTraceFlow(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetReportTraceFlow(id)
	respond(c, data, err)
}

// GetSubmissionDetail returns detail of a single submission
func (h *Handler) GetSubmissionDetail(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.GetSubmissionDetail(id)
	respond(c, data, err)
}

// GetStudentAssignmentSubmission mengambil pengumpulan milik mahasiswa yang login
// untuk sebuah tugas, dipakai tombol "Lihat Detail" di daftar tugas kursus.
func (h *Handler) GetStudentAssignmentSubmission(c *gin.Context) {
	assignmentID, err := pathID(c)
	if err != nil {
		return
	}
	authUser, _ := currentAuthUser(c)
	data, err := h.repo.StudentAssignmentSubmission(assignmentID, authUser.ID)
	respond(c, data, err)
}
