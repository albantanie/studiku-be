package handlers

import "github.com/gin-gonic/gin"

func registerPageData(api *gin.RouterGroup, h *Handler) {
	api.GET("/admin/dashboard", requireRole("admin"), h.AdminDashboard)
	api.GET("/admin/attendance", requireRole("admin"), h.AdminAttendance)
	api.GET("/admin/grades", requireRole("admin"), h.AdminGrades)
	api.GET("/lecturer/dashboard", requireRole("kalab"), h.LecturerDashboard)
	api.GET("/lecturer/courses", requireRole("kalab", "laboran"), h.LecturerCourses)
	api.GET("/lecturer/attendance", requireRole("kalab"), h.LecturerAttendance)
	api.GET("/lecturer/grades", requireRole("kalab"), h.LecturerGrades)
	api.GET("/assistant/dashboard", requireRole("aslab"), h.AssistantDashboard)
	api.GET("/assistant/practicals", requireRole("aslab"), h.AssistantPracticals)
	api.GET("/assistant/attendance", requireRole("aslab"), h.AssistantAttendance)
	api.GET("/assistant/reports", requireRole("aslab", "laboran", "kalab", "admin"), h.AssistantReports)
	api.GET("/student/courses/:id/detail", requireRole("student"), h.StudentCourseDetail)
	api.GET("/student/sessions/:id/detail", requireRole("student"), h.StudentSessionDetail)
	api.GET("/student/materials", requireRole("student"), h.StudentMaterials)
	api.GET("/admin/reports", requireRole("admin"), h.AdminReports)
	api.GET("/admin/assignments", requireRole("admin"), h.AdminAssignments)
	api.GET("/admin/class-management", requireRole("admin", "laboran"), h.AdminClassManagement)
	api.GET("/admin/class-students", requireRole("admin", "laboran"), h.AdminClassStudents)
	api.GET("/admin/course-form-options", requireRole("admin", "laboran"), h.AdminCourseFormOptions)
	api.GET("/student/grade-semesters", requireRole("student"), h.StudentGradeSemesters)
	api.GET("/admin/student-programs", requireRole("admin"), h.AdminStudentPrograms)
	api.GET("/admin/labs", requireRole("admin"), h.AdminLabs)
	api.GET("/admin/user-tabs", requireRole("admin"), h.AdminUserTabs)
	api.GET("/admin/import-lecturer-preview", requireRole("admin"), h.AdminImportLecturerPreview)
	api.GET("/admin/import-file-types", requireRole("admin"), h.AdminImportFileTypes)
}

func (h *Handler) pageData(c *gin.Context, key string) {
	data, err := h.repo.PageData(key)
	respond(c, data, err)
}

func (h *Handler) AdminDashboard(c *gin.Context)      { h.pageData(c, "admin_dashboard") }
func (h *Handler) LecturerDashboard(c *gin.Context)   { h.pageData(c, "lecturer_dashboard") }
func (h *Handler) LecturerCourses(c *gin.Context)     { h.pageData(c, "lecturer_courses") }
func (h *Handler) LecturerAttendance(c *gin.Context)  { h.pageData(c, "lecturer_attendance") }
func (h *Handler) LecturerGrades(c *gin.Context)      { h.pageData(c, "lecturer_grades") }
func (h *Handler) AssistantDashboard(c *gin.Context)  { h.pageData(c, "assistant_dashboard") }
func (h *Handler) AssistantPracticals(c *gin.Context) { h.pageData(c, "assistant_practicals") }
func (h *Handler) AssistantAttendance(c *gin.Context) { h.pageData(c, "assistant_attendance") }
func (h *Handler) AssistantReports(c *gin.Context)    { h.pageData(c, "assistant_reports") }
func (h *Handler) AdminAttendance(c *gin.Context)     { h.pageData(c, "admin_attendance") }
func (h *Handler) AdminGrades(c *gin.Context)         { h.pageData(c, "admin_grades") }
func (h *Handler) StudentCourseDetail(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.StudentCourseDetail(id)
	respond(c, data, err)
}
func (h *Handler) StudentSessionDetail(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		return
	}
	data, err := h.repo.StudentSessionDetail(id)
	respond(c, data, err)
}
func (h *Handler) StudentMaterials(c *gin.Context)       { h.pageData(c, "student_materials") }
func (h *Handler) AdminReports(c *gin.Context)           { h.pageData(c, "admin_reports") }
func (h *Handler) AdminAssignments(c *gin.Context)       { h.pageData(c, "admin_assignments") }
func (h *Handler) AdminClassManagement(c *gin.Context)   { h.pageData(c, "admin_class_management") }
func (h *Handler) AdminClassStudents(c *gin.Context)     { h.pageData(c, "admin_class_students") }
func (h *Handler) AdminCourseFormOptions(c *gin.Context) { h.pageData(c, "admin_course_form_options") }
func (h *Handler) StudentGradeSemesters(c *gin.Context)  { h.pageData(c, "student_grade_semesters") }
func (h *Handler) AdminStudentPrograms(c *gin.Context)   { h.pageData(c, "admin_student_programs") }
func (h *Handler) AdminLabs(c *gin.Context)              { h.pageData(c, "admin_labs") }
func (h *Handler) AdminUserTabs(c *gin.Context)          { h.pageData(c, "admin_user_tabs") }
func (h *Handler) AdminImportLecturerPreview(c *gin.Context) {
	h.pageData(c, "admin_import_lecturer_preview")
}
func (h *Handler) AdminImportFileTypes(c *gin.Context) { h.pageData(c, "admin_import_file_types") }
