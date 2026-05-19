package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"studi-ku-backend/internal/models"
)

var studentImportHeaders = []string{"nim", "nama", "email", "no_hp", "jenis_kelamin", "program_studi", "angkatan", "status"}

type studentImportRow struct {
	StudentID    string
	Name         string
	Email        string
	Phone        string
	Gender       string
	Program      string
	BatchYear    int
	Status       string
	RowNumber    int
	SourceRowRaw map[string]string
}

func (h *Handler) DownloadStudentImportTemplate(c *gin.Context) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := f.GetSheetName(0)
	for i, header := range studentImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}

	_ = f.SetCellValue(sheet, "A2", "240001")
	_ = f.SetCellValue(sheet, "B2", "Budi Santoso")
	_ = f.SetCellValue(sheet, "C2", "budi@example.com")
	_ = f.SetCellValue(sheet, "D2", "081234567890")
	_ = f.SetCellValue(sheet, "E2", "L")
	_ = f.SetCellValue(sheet, "F2", "Teknik Informatika")
	_ = f.SetCellValue(sheet, "G2", "2024")
	_ = f.SetCellValue(sheet, "H2", "aktif")

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="template_import_mahasiswa.xlsx"`)
	if err := f.Write(c.Writer); err != nil {
		fail(c, http.StatusInternalServerError, "gagal membuat template xlsx")
		return
	}
}

func (h *Handler) PreviewStudentImport(c *gin.Context) {
	rows, err := parseStudentImportXLSX(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	ok(c, "success", gin.H{"preview": previewStudents(rows), "rows": rows})
}

func (h *Handler) ImportStudents(c *gin.Context) {
	rows, err := parseStudentImportXLSX(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	students := make([]models.Student, 0, len(rows))
	for _, row := range rows {
		students = append(students, models.Student{
			Name:            row.Name,
			Email:           strings.ToLower(row.Email),
			StudentID:       row.StudentID,
			Program:         row.Program,
			Semester:        semesterFromBatch(row.BatchYear, time.Now().Year()),
			Status:          normalizeStatus(row.Status),
			Courses:         []string{},
			Password:        "password",
			DefaultPassword: "password",
		})
	}

	if err := h.repo.ImportStudents(students); err != nil {
		fail(c, http.StatusBadRequest, fmt.Sprintf("gagal simpan data import: %v", err))
		return
	}

	all, err := h.repo.Students()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	ok(c, "success", gin.H{"imported": len(students), "students": all})
}

func parseStudentImportXLSX(c *gin.Context) ([]studentImportRow, error) {
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

	headers := normalizeHeaders(rows[0])
	headerIndex := map[string]int{}
	for idx, h := range headers {
		headerIndex[h] = idx
	}
	for _, required := range studentImportHeaders {
		if _, ok := headerIndex[required]; !ok {
			return nil, fmt.Errorf("kolom wajib tidak ditemukan: %s", required)
		}
	}

	result := make([]studentImportRow, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		rowNum := i + 1
		row := rows[i]
		data := map[string]string{}
		for _, h := range studentImportHeaders {
			data[h] = pickCell(row, headerIndex[h])
		}

		if allBlank(data) {
			continue
		}

		if err := requireFields(data, "nim", "nama", "email", "no_hp", "jenis_kelamin", "program_studi", "angkatan", "status"); err != nil {
			return nil, fmt.Errorf("baris %d: %v", rowNum, err)
		}
		if !strings.Contains(data["email"], "@") {
			return nil, fmt.Errorf("baris %d: email tidak valid", rowNum)
		}
		if err := validatePhone(data["no_hp"]); err != nil {
			return nil, fmt.Errorf("baris %d: %v", rowNum, err)
		}
		if err := validateGender(data["jenis_kelamin"]); err != nil {
			return nil, fmt.Errorf("baris %d: %v", rowNum, err)
		}
		angkatan, err := strconv.Atoi(data["angkatan"])
		if err != nil || angkatan < 1900 || angkatan > 3000 {
			return nil, fmt.Errorf("baris %d: angkatan harus angka valid", rowNum)
		}

		result = append(result, studentImportRow{
			StudentID:    data["nim"],
			Name:         data["nama"],
			Email:        data["email"],
			Phone:        data["no_hp"],
			Gender:       data["jenis_kelamin"],
			Program:      data["program_studi"],
			BatchYear:    angkatan,
			Status:       data["status"],
			RowNumber:    rowNum,
			SourceRowRaw: data,
		})
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("tidak ada data valid untuk diimport")
	}
	return result, nil
}

func previewStudents(rows []studentImportRow) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"name":      row.Name,
			"email":     strings.ToLower(row.Email),
			"studentId": row.StudentID,
			"program":   row.Program,
			"angkatan":  row.BatchYear,
			"status":    normalizeStatus(row.Status),
		})
	}
	return out
}

func normalizeHeaders(headers []string) []string {
	out := make([]string, len(headers))
	for i, h := range headers {
		x := strings.TrimSpace(strings.ToLower(h))
		x = strings.ReplaceAll(x, " ", "_")
		out[i] = x
	}
	return out
}

func pickCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func allBlank(m map[string]string) bool {
	for _, v := range m {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func requireFields(data map[string]string, fields ...string) error {
	for _, f := range fields {
		if strings.TrimSpace(data[f]) == "" {
			return fmt.Errorf("kolom %s wajib diisi", f)
		}
	}
	return nil
}

func normalizeStatus(s string) string {
	x := strings.ToLower(strings.TrimSpace(s))
	if x == "aktif" {
		return "Aktif"
	}
	if x == "cuti" {
		return "Cuti"
	}
	if x == "" {
		return "Aktif"
	}
	return strings.ToUpper(x[:1]) + x[1:]
}

func semesterFromBatch(batchYear, currentYear int) int {
	if batchYear <= 0 {
		return 1
	}
	semester := (currentYear-batchYear)*2 + 1
	if semester < 1 {
		return 1
	}
	if semester > 14 {
		return 14
	}
	return semester
}

func validatePhone(v string) error {
	x := strings.TrimSpace(v)
	if x == "" {
		return fmt.Errorf("no_hp wajib diisi")
	}
	if strings.HasPrefix(x, "+") {
		x = x[1:]
	}
	if len(x) < 10 || len(x) > 15 {
		return fmt.Errorf("no_hp harus 10-15 digit")
	}
	for _, ch := range x {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("no_hp hanya boleh angka")
		}
	}
	return nil
}

func validateGender(v string) error {
	x := strings.ToLower(strings.TrimSpace(v))
	switch x {
	case "l", "p", "laki-laki", "laki", "perempuan":
		return nil
	default:
		return fmt.Errorf("jenis_kelamin harus L atau P")
	}
}
