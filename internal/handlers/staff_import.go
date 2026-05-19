package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"studi-ku-backend/internal/models"
)

var lecturerImportHeaders = []string{"nama", "email", "nidn", "mata_kuliah"}
var assistantImportHeaders = []string{"nim", "nama", "email", "no_hp", "lab", "supervisor", "semester", "ipk", "status"}

type lecturerImportRow struct {
	Name    string
	Email   string
	NIDN    string
	Courses []string
}

type assistantImportRow struct {
	StudentID  string
	Name       string
	Email      string
	Phone      string
	Lab        string
	Supervisor string
	Semester   int
	GPA        float64
	Status     string
}

func (h *Handler) DownloadLecturerImportTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)
	for i, header := range lecturerImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}
	_ = f.SetCellValue(sheet, "A2", "Dr. Budi Santoso")
	_ = f.SetCellValue(sheet, "B2", "budi.dosen@example.com")
	_ = f.SetCellValue(sheet, "C2", "0123456789")
	_ = f.SetCellValue(sheet, "D2", "Algoritma;Basis Data")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="template_import_dosen.xlsx"`)
	if err := f.Write(c.Writer); err != nil {
		fail(c, http.StatusInternalServerError, "gagal membuat template xlsx")
		return
	}
}

func (h *Handler) DownloadAssistantImportTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)
	for i, header := range assistantImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}
	_ = f.SetCellValue(sheet, "A2", "240001")
	_ = f.SetCellValue(sheet, "B2", "Andi Pratama")
	_ = f.SetCellValue(sheet, "C2", "andi.aslab@example.com")
	_ = f.SetCellValue(sheet, "D2", "081234567890")
	_ = f.SetCellValue(sheet, "E2", "Laboratorium Pemrograman")
	_ = f.SetCellValue(sheet, "F2", "Dr. Budi Santoso")
	_ = f.SetCellValue(sheet, "G2", "6")
	_ = f.SetCellValue(sheet, "H2", "3.50")
	_ = f.SetCellValue(sheet, "I2", "Aktif")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="template_import_aslab.xlsx"`)
	if err := f.Write(c.Writer); err != nil {
		fail(c, http.StatusInternalServerError, "gagal membuat template xlsx")
		return
	}
}

func (h *Handler) PreviewLecturerImport(c *gin.Context) {
	rows, err := parseLecturerImportXLSX(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	preview := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		preview = append(preview, map[string]interface{}{
			"name":    row.Name,
			"email":   row.Email,
			"nidn":    row.NIDN,
			"courses": row.Courses,
		})
	}
	ok(c, "success", gin.H{"preview": preview})
}

func (h *Handler) ImportLecturers(c *gin.Context) {
	rows, err := parseLecturerImportXLSX(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	items := make([]models.Lecturer, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.Lecturer{
			Name:              row.Name,
			Email:             strings.ToLower(row.Email),
			NIDN:              row.NIDN,
			Courses:           row.Courses,
			Password:          "password",
			DefaultPassword:   "password",
			IsPasswordChanged: false,
		})
	}
	if err := h.repo.ImportLecturers(items); err != nil {
		fail(c, http.StatusBadRequest, fmt.Sprintf("gagal simpan data import: %v", err))
		return
	}
	all, err := h.repo.Lecturers()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, "success", gin.H{"imported": len(items), "lecturers": all})
}

func (h *Handler) PreviewAssistantImport(c *gin.Context) {
	rows, err := parseAssistantImportXLSX(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	preview := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		preview = append(preview, map[string]interface{}{
			"studentId":  row.StudentID,
			"name":       row.Name,
			"email":      row.Email,
			"phone":      row.Phone,
			"lab":        row.Lab,
			"supervisor": row.Supervisor,
			"semester":   row.Semester,
			"gpa":        row.GPA,
			"status":     row.Status,
		})
	}
	ok(c, "success", gin.H{"preview": preview})
}

func (h *Handler) ImportAssistants(c *gin.Context) {
	rows, err := parseAssistantImportXLSX(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	items := make([]models.LabAssistant, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.LabAssistant{
			Name:              row.Name,
			Email:             strings.ToLower(row.Email),
			Phone:             row.Phone,
			StudentID:         row.StudentID,
			Lab:               row.Lab,
			Supervisor:        row.Supervisor,
			Semester:          row.Semester,
			GPA:               row.GPA,
			AssignedCourses:   0,
			WeeklyHours:       0,
			Status:            row.Status,
			Password:          "password",
			DefaultPassword:   "password",
			IsPasswordChanged: false,
		})
	}
	if err := h.repo.ImportAssistants(items); err != nil {
		fail(c, http.StatusBadRequest, fmt.Sprintf("gagal simpan data import: %v", err))
		return
	}
	all, err := h.repo.Assistants()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, "success", gin.H{"imported": len(items), "assistants": all})
}

func parseLecturerImportXLSX(c *gin.Context) ([]lecturerImportRow, error) {
	rows, err := parseGenericXLSX(c)
	if err != nil {
		return nil, err
	}
	headers := normalizeHeaders(rows[0])
	hm := map[string]int{}
	for i, h := range headers {
		hm[h] = i
	}
	for _, required := range lecturerImportHeaders {
		if _, ok := hm[required]; !ok {
			return nil, fmt.Errorf("kolom wajib tidak ditemukan: %s", required)
		}
	}
	out := []lecturerImportRow{}
	for i := 1; i < len(rows); i++ {
		rowNum := i + 1
		row := rows[i]
		data := map[string]string{}
		for _, h := range lecturerImportHeaders {
			data[h] = pickCell(row, hm[h])
		}
		if allBlank(data) {
			continue
		}
		if err := requireFields(data, "nama", "email", "nidn"); err != nil {
			return nil, fmt.Errorf("baris %d: %v", rowNum, err)
		}
		if !strings.Contains(data["email"], "@") {
			return nil, fmt.Errorf("baris %d: email tidak valid", rowNum)
		}
		courses := []string{}
		for _, part := range strings.Split(data["mata_kuliah"], ";") {
			x := strings.TrimSpace(part)
			if x != "" {
				courses = append(courses, x)
			}
		}
		out = append(out, lecturerImportRow{
			Name:    data["nama"],
			Email:   data["email"],
			NIDN:    data["nidn"],
			Courses: courses,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("tidak ada data valid untuk diimport")
	}
	return out, nil
}

func parseAssistantImportXLSX(c *gin.Context) ([]assistantImportRow, error) {
	rows, err := parseGenericXLSX(c)
	if err != nil {
		return nil, err
	}
	headers := normalizeHeaders(rows[0])
	hm := map[string]int{}
	for i, h := range headers {
		hm[h] = i
	}
	for _, required := range assistantImportHeaders {
		if _, ok := hm[required]; !ok {
			return nil, fmt.Errorf("kolom wajib tidak ditemukan: %s", required)
		}
	}
	out := []assistantImportRow{}
	for i := 1; i < len(rows); i++ {
		rowNum := i + 1
		row := rows[i]
		data := map[string]string{}
		for _, h := range assistantImportHeaders {
			data[h] = pickCell(row, hm[h])
		}
		if allBlank(data) {
			continue
		}
		if err := requireFields(data, "nim", "nama", "email", "no_hp", "lab", "semester", "ipk", "status"); err != nil {
			return nil, fmt.Errorf("baris %d: %v", rowNum, err)
		}
		if !strings.Contains(data["email"], "@") {
			return nil, fmt.Errorf("baris %d: email tidak valid", rowNum)
		}
		if err := validatePhone(data["no_hp"]); err != nil {
			return nil, fmt.Errorf("baris %d: %v", rowNum, err)
		}
		semester, err := strconv.Atoi(data["semester"])
		if err != nil || semester < 1 || semester > 14 {
			return nil, fmt.Errorf("baris %d: semester harus angka 1-14", rowNum)
		}
		gpa, err := strconv.ParseFloat(data["ipk"], 64)
		if err != nil || gpa < 0 || gpa > 4 {
			return nil, fmt.Errorf("baris %d: ipk harus angka 0.00-4.00", rowNum)
		}
		out = append(out, assistantImportRow{
			StudentID:  data["nim"],
			Name:       data["nama"],
			Email:      data["email"],
			Phone:      data["no_hp"],
			Lab:        data["lab"],
			Supervisor: data["supervisor"],
			Semester:   semester,
			GPA:        gpa,
			Status:     normalizeStatus(data["status"]),
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("tidak ada data valid untuk diimport")
	}
	return out, nil
}

func parseGenericXLSX(c *gin.Context) ([][]string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("file wajib diunggah")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".xlsx" {
		return nil, fmt.Errorf("format file tidak valid: hanya .xlsx")
	}
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file upload")
	}
	defer func() { _ = src.Close() }()
	xl, err := excelize.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("gagal parsing xlsx: %v", err)
	}
	defer func() { _ = xl.Close() }()
	sheet := xl.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("sheet pertama tidak ditemukan")
	}
	rows, err := xl.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca isi sheet: %v", err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("file kosong: minimal header + 1 data")
	}
	return rows, nil
}
