package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
		api.PUT("/lecturer/grades/students/:id", requireRole("kalab"), h.UpdateStudentGrade)
		api.POST("/lecturer/courses", requireRole("laboran", "aslab"), h.CreateLecturerCourse)
		api.PUT("/lecturer/courses/:id", requireRole("admin", "laboran", "aslab", "kalab"), h.UpdateLecturerCourse)
		api.DELETE("/lecturer/courses/:id", requireRole("laboran"), h.DeleteLecturerCourse)
		api.DELETE("/assistant/materials/:id", requireRole("aslab"), h.DeleteMaterial)
		api.GET("/materials", requireRole("admin", "kalab", "laboran", "aslab"), h.GetMaterials)
		api.POST("/materials", requireRole("aslab"), h.CreateMaterial)
		api.PUT("/materials/:id/submit", requireRole("aslab"), h.SubmitMaterial)
		api.GET("/materials/:id/file", requireRole("admin", "kalab", "laboran", "aslab", "student"), h.MaterialFile)

		api.GET("/admin/courses", requireRole("admin", "laboran", "kalab"), h.GetAdminCourses)
		api.POST("/admin/courses", requireRole("laboran"), h.CreateCourse)
		api.PUT("/admin/courses/:id", requireRole("admin", "laboran", "kalab"), h.UpdateCourse)
		api.DELETE("/admin/courses/:id", requireRole("laboran"), h.DeleteCourse)

		api.GET("/admin/academic-years", requireRole("admin", "laboran"), h.GetAcademicYears)
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
		api.GET("/submissions/:id/download", requireRole("admin", "kalab", "laboran", "aslab"), h.ReportFile)
		api.PUT("/reports/:id/approve", requireRole("laboran", "kalab"), h.ApproveReport)
		api.PUT("/reports/:id/reject", requireRole("laboran", "kalab"), h.RejectReport)
		api.PUT("/assistant/submissions/:id/grade", requireRole("aslab", "laboran"), h.ReviewAssistantReport)
		api.PUT("/assistant/attendance/sessions/:id", requireRole("aslab"), h.UpdateAssistantAttendanceSession)
		api.PUT("/assistant/course-sessions/:id/attendance", requireRole("aslab"), h.UpdateCourseSessionAttendance)
		api.POST("/assistant/sessions/:id/reports", requireRole("aslab"), h.SubmitAssistantSessionReport)
		api.PUT("/lecturer/reports/:id/approve", requireRole("kalab"), h.ApproveAssistantReport)
		api.PUT("/lecturer/reports/:id/reject", requireRole("kalab"), h.RejectAssistantReport)
		api.GET("/reports/workflow", requireRole("aslab", "laboran", "kalab", "admin"), h.GetReportWorkflow)
		api.POST("/reports/workflow/submit", requireRole("aslab"), h.SubmitReportWorkflow)
		api.POST("/reports/workflow/approve", requireRole("laboran", "kalab"), h.ApproveReportWorkflow)
		api.POST("/reports/workflow/reject", requireRole("laboran", "kalab"), h.RejectReportWorkflow)
		api.POST("/reports/workflow/reset", requireRole("aslab"), h.ResetReportWorkflow)
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
	data, err := h.repo.Assignments()
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
				c.Header("Content-Type", "application/pdf")
				c.Header("Content-Disposition", `inline; filename="laporan.pdf"`)
				c.Data(http.StatusOK, "application/pdf", simplePDFBytes(fmt.Sprintf("Laporan %s\nKelas %s\nTopik %s\nStatus %s\nCatatan %s", report.CourseName, report.Class, report.Topic, report.Status, report.RejectionNote)))
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
	dir := filepath.Join(h.uploadDir, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d-%s", time.Now().UnixNano(), safeFileName(originalName))
	relativePath := filepath.Join(folder, name)
	fullPath := filepath.Join(h.uploadDir, relativePath)
	return relativePath, saveUploadedFile(fileHeader, fullPath)
}

func (h *Handler) serveUploadFile(c *gin.Context, relativePath string) {
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `inline; filename="file.pdf"`)
	c.File(filepath.Join(h.uploadDir, relativePath))
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

func simplePDFBytes(text string) []byte {
	escaped := strings.ReplaceAll(text, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "(", "\\(")
	escaped = strings.ReplaceAll(escaped, ")", "\\)")
	lines := strings.Split(escaped, "\n")
	content := "BT /F1 12 Tf 50 780 Td "
	for index, line := range lines {
		if index > 0 {
			content += "0 -18 Td "
		}
		content += "(" + line + ") Tj "
	}
	content += "ET"
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
	}
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
	respond(c, gin.H{"id": id}, h.repo.SetAssistantReportStatus(id, "Disetujui"))
}
func (h *Handler) RejectAssistantReport(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	respond(c, gin.H{"id": id}, h.repo.SetAssistantReportStatus(id, "Ditolak"))
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
