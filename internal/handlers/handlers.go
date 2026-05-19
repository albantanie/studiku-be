package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"studi-ku-backend/internal/models"
	"studi-ku-backend/internal/repositories"
)

type Handler struct{ repo *repositories.Repository }

func New(repo *repositories.Repository) *Handler { return &Handler{repo: repo} }

func Register(r *gin.Engine, h *Handler) {
	r.GET("/api/health", func(c *gin.Context) { ok(c, "healthy", gin.H{"status": "ok"}) })
	api := r.Group("/api")
	{
		api.POST("/auth/login", h.Login)
		registerPageData(api, h)

		api.GET("/student/dashboard", h.GetDashboard)
		api.GET("/student/courses", h.GetStudentCourses)
		api.GET("/student/assignments", h.GetAssignments)
		api.GET("/student/grades", h.GetGrades)
		api.PUT("/student/assignments/:id/submit", h.SubmitAssignment)
		api.PUT("/lecturer/grades/students/:id", h.UpdateStudentGrade)
		api.POST("/lecturer/courses", h.CreateLecturerCourse)
		api.PUT("/lecturer/courses/:id", h.UpdateLecturerCourse)
		api.DELETE("/lecturer/courses/:id", h.DeleteLecturerCourse)
		api.DELETE("/assistant/materials/:id", h.DeleteMaterial)

		api.GET("/admin/courses", h.GetAdminCourses)
		api.POST("/admin/courses", h.CreateCourse)
		api.PUT("/admin/courses/:id", h.UpdateCourse)
		api.DELETE("/admin/courses/:id", h.DeleteCourse)

		api.GET("/admin/academic-years", h.GetAcademicYears)
		api.POST("/admin/academic-years", h.CreateAcademicYear)
		api.PUT("/admin/academic-years/:id", h.UpdateAcademicYear)
		api.DELETE("/admin/academic-years/:id", h.DeleteAcademicYear)

		api.GET("/admin/students", h.GetStudents)
		api.POST("/admin/students", h.CreateStudent)
		api.PUT("/admin/students/:id", h.UpdateStudent)
		api.DELETE("/admin/students/:id", h.DeleteStudent)
		api.PUT("/admin/students/:id/reset-password", h.ResetStudentPassword)

		api.GET("/admin/lecturers", h.GetLecturers)
		api.POST("/admin/lecturers", h.CreateLecturer)
		api.PUT("/admin/lecturers/:id", h.UpdateLecturer)
		api.DELETE("/admin/lecturers/:id", h.DeleteLecturer)
		api.PUT("/admin/lecturers/:id/reset-password", h.ResetLecturerPassword)

		api.GET("/admin/lab-assistants", h.GetAssistants)
		api.POST("/admin/lab-assistants", h.CreateAssistant)
		api.PUT("/admin/lab-assistants/:id", h.UpdateAssistant)
		api.DELETE("/admin/lab-assistants/:id", h.DeleteAssistant)
		api.PUT("/admin/lab-assistants/:id/reset-password", h.ResetAssistantPassword)

		api.GET("/admin/classes", h.GetClasses)
		api.POST("/admin/classes", h.CreateClass)
		api.PUT("/admin/classes/:id", h.UpdateClass)
		api.DELETE("/admin/classes/:id", h.DeleteClass)

		api.POST("/admin/assignments", h.CreateAdminAssignment)
		api.PUT("/admin/assignments/:id", h.UpdateAdminAssignment)
		api.DELETE("/admin/assignments/:id", h.DeleteAdminAssignment)

		api.PUT("/assistant/reports/:id/review", h.ReviewAssistantReport)
		api.PUT("/assistant/submissions/:id/grade", h.ReviewAssistantReport)
		api.GET("/reports/workflow", h.GetReportWorkflow)
		api.POST("/reports/workflow/submit", h.SubmitReportWorkflow)
		api.POST("/reports/workflow/approve", h.ApproveReportWorkflow)
		api.POST("/reports/workflow/reject", h.RejectReportWorkflow)
		api.POST("/reports/workflow/reset", h.ResetReportWorkflow)
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
	respond(c, user, err)
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
