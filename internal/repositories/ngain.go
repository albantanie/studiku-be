package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// N-Gain (normalized gain) mengikuti Hake (1998):
//
//	        skor post-test - skor pretest
//	g = ---------------------------------------
//	     skor maksimum ideal - skor pretest
//
// Pembilang adalah kenaikan yang benar-benar dicapai mahasiswa, penyebut adalah
// kenaikan maksimum yang masih mungkin dicapai. Jadi N-Gain mengukur berapa
// persen dari ruang perbaikan yang tersisa berhasil dimanfaatkan, bukan sekadar
// selisih mentah post-test dan pretest.

// Status hasil perhitungan N-Gain untuk satu mahasiswa.
const (
	// NGainStatusOK berarti pretest dan post-test lengkap dan nilai naik.
	NGainStatusOK = "ok"
	// NGainStatusRegression berarti post-test lebih rendah dari pretest (N-Gain negatif).
	NGainStatusRegression = "regression"
	// NGainStatusCeiling berarti pretest sudah menyentuh skor maksimum sehingga
	// penyebut rumus bernilai nol dan N-Gain tidak terdefinisi.
	NGainStatusCeiling = "ceiling"
	// NGainStatusIncomplete berarti salah satu dari pretest atau post-test belum dikerjakan.
	NGainStatusIncomplete = "incomplete"
)

// Ambang kategori Hake (1998).
const (
	nGainHighThreshold   = 0.7
	nGainMediumThreshold = 0.3
)

// NGainValue adalah hasil perhitungan N-Gain satu mahasiswa.
//
// Struct ini sengaja hanya membawa hasil, bukan uraian cara menghitungnya.
// Antarmuka menampilkan nilai akhir dan kategorinya saja; penjabaran rumus
// dengan angka mahasiswa tidak dikirim maupun ditampilkan.
type NGainValue struct {
	Pre           float64 `json:"pre"`
	Post          float64 `json:"post"`
	MaxScore      float64 `json:"maxScore"`
	Gain          float64 `json:"gain"`    // pembilang: post - pre
	MaxGain       float64 `json:"maxGain"` // penyebut: maxScore - pre
	NGain         float64 `json:"nGain"`
	Percent       float64 `json:"percent"`
	Category      string  `json:"category"`
	CategoryColor string  `json:"categoryColor"`
	Effectiveness string  `json:"effectiveness"`
	Status        string  `json:"status"`
	Valid         bool    `json:"valid"`
	Note          string  `json:"note"`
	HasPre        bool    `json:"hasPre"`
	HasPost       bool    `json:"hasPost"`
}

// NGainSummary meringkas satu kelas atau satu rombongan mahasiswa.
type NGainSummary struct {
	StudentCount   int     `json:"studentCount"`
	ComputedCount  int     `json:"computedCount"`
	IncompleteCnt  int     `json:"incompleteCount"`
	CeilingCount   int     `json:"ceilingCount"`
	AveragePre     float64 `json:"averagePre"`
	AveragePost    float64 `json:"averagePost"`
	MaxScore       float64 `json:"maxScore"`
	AverageNGain   float64 `json:"averageNGain"`   // rata-rata N-Gain tiap mahasiswa
	AveragePercent float64 `json:"averagePercent"` // rata-rata N-Gain dalam persen
	// NGainOfAverages adalah <g> versi asli Hake: dihitung dari rata-rata kelas,
	// bukan dari rata-rata N-Gain individu. Keduanya ditampilkan karena bisa berbeda.
	NGainOfAverages float64        `json:"nGainOfAverages"`
	Category        string         `json:"category"`
	CategoryColor   string         `json:"categoryColor"`
	Effectiveness   string         `json:"effectiveness"`
	Distribution    []NGainBucket  `json:"distribution"`
	Interpretation  string         `json:"interpretation"`
	Thresholds      NGainReference `json:"thresholds"`
}

// NGainBucket adalah satu baris sebaran kategori.
type NGainBucket struct {
	Category string  `json:"category"`
	Color    string  `json:"color"`
	Range    string  `json:"range"`
	Count    int     `json:"count"`
	Percent  float64 `json:"percent"`
}

// NGainReference adalah tabel acuan yang ditampilkan sebagai panduan di UI.
type NGainReference struct {
	Source        string              `json:"source"`
	Formula       string              `json:"formula"`
	Categories    []NGainRefRow       `json:"categories"`
	Effectiveness []NGainRefRow       `json:"effectiveness"`
	Steps         []NGainGuidedStep   `json:"steps"`
	Cautions      []string            `json:"cautions"`
	Glossary      []NGainGlossaryTerm `json:"glossary"`
}

// NGainRefRow adalah satu baris tabel interpretasi.
type NGainRefRow struct {
	Range string `json:"range"`
	Label string `json:"label"`
	Color string `json:"color"`
}

// NGainGuidedStep adalah satu langkah pada panduan berurutan di UI.
type NGainGuidedStep struct {
	Order  int    `json:"order"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// NGainGlossaryTerm menjelaskan satu istilah yang muncul di rumus.
type NGainGlossaryTerm struct {
	Term    string `json:"term"`
	Meaning string `json:"meaning"`
}

// NGainStudentRow adalah satu baris tabel hasil per mahasiswa.
type NGainStudentRow struct {
	StudentID         int        `json:"studentId"`
	NIM               string     `json:"nim"`
	Name              string     `json:"name"`
	PretestSessions   int        `json:"pretestSessions,omitempty"`
	PosttestSessions  int        `json:"posttestSessions,omitempty"`
	NGain             NGainValue `json:"ngain"`
	ImprovementPoints float64    `json:"improvementPoints"`
}

// round membulatkan ke sejumlah angka di belakang koma.
func round(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}

// formatID menulis angka dengan koma desimal seperti kebiasaan penulisan Indonesia.
func formatID(value float64, places int) string {
	return strings.Replace(fmt.Sprintf("%.*f", places, value), ".", ",", 1)
}

// nGainCategory memetakan nilai g ke kategori Hake.
func nGainCategory(g float64) (string, string) {
	switch {
	case g >= nGainHighThreshold:
		return "Tinggi", "green"
	case g >= nGainMediumThreshold:
		return "Sedang", "amber"
	default:
		return "Rendah", "red"
	}
}

// nGainEffectiveness memetakan N-Gain persen ke tafsiran keefektifan pembelajaran.
func nGainEffectiveness(percent float64) string {
	switch {
	case percent >= 76:
		return "Efektif"
	case percent >= 56:
		return "Cukup Efektif"
	case percent >= 40:
		return "Kurang Efektif"
	default:
		return "Tidak Efektif"
	}
}

// ComputeNGain menghitung N-Gain satu mahasiswa. Nilai pre dan post bernilai nil
// bila tes yang bersangkutan belum dikerjakan.
func ComputeNGain(pre *float64, post *float64, maxScore float64) NGainValue {
	if maxScore <= 0 {
		maxScore = 100
	}
	result := NGainValue{
		MaxScore: maxScore,
		HasPre:   pre != nil,
		HasPost:  post != nil,
	}

	if pre == nil || post == nil {
		result.Status = NGainStatusIncomplete
		result.Category = "Belum Lengkap"
		result.CategoryColor = "gray"
		result.Effectiveness = "Belum dapat dinilai"
		switch {
		case pre == nil && post == nil:
			result.Note = "Pretest dan post-test belum dikerjakan."
		case pre == nil:
			result.Note = "Pretest belum dikerjakan, N-Gain butuh nilai awal sebagai pembanding."
		default:
			result.Note = "Post-test belum dikerjakan, N-Gain butuh nilai akhir sebagai pembanding."
		}
		if pre != nil {
			result.Pre = *pre
		}
		if post != nil {
			result.Post = *post
		}
		return result
	}

	result.Pre = *pre
	result.Post = *post
	result.Gain = result.Post - result.Pre
	result.MaxGain = maxScore - result.Pre

	// Pretest sudah maksimum: tidak ada ruang perbaikan, penyebut nol.
	if result.MaxGain <= 0 {
		result.Status = NGainStatusCeiling
		result.Category = "Tidak Terdefinisi"
		result.CategoryColor = "gray"
		result.Effectiveness = "Tidak dapat dihitung"
		result.Note = "Nilai pretest sudah menyentuh skor maksimum sehingga penyebut rumus bernilai nol. Mahasiswa ini dikeluarkan dari rata-rata kelas."
		return result
	}

	result.NGain = round(result.Gain/result.MaxGain, 4)
	result.Percent = round(result.NGain*100, 2)
	result.Category, result.CategoryColor = nGainCategory(result.NGain)
	result.Effectiveness = nGainEffectiveness(result.Percent)
	result.Valid = true

	if result.Gain < 0 {
		result.Status = NGainStatusRegression
		result.Note = "Nilai post-test lebih rendah dari pretest sehingga N-Gain negatif. Periksa kembali kesulitan soal atau kondisi pelaksanaan tes."
	} else {
		result.Status = NGainStatusOK
	}
	return result
}

// nGainReference mengembalikan materi panduan statis yang dipakai di seluruh UI.
func nGainReference() NGainReference {
	return NGainReference{
		Source:  "Hake, R. R. (1998). Interactive-engagement versus traditional methods. American Journal of Physics, 66(1), 64-74.",
		Formula: "N-Gain (g) = (Post-test - Pretest) / (Skor Maksimum Ideal - Pretest)",
		Categories: []NGainRefRow{
			{Range: "g >= 0,70", Label: "Tinggi", Color: "green"},
			{Range: "0,30 <= g < 0,70", Label: "Sedang", Color: "amber"},
			{Range: "g < 0,30", Label: "Rendah", Color: "red"},
		},
		Effectiveness: []NGainRefRow{
			{Range: "N-Gain >= 76%", Label: "Efektif", Color: "green"},
			{Range: "56% - 75%", Label: "Cukup Efektif", Color: "lime"},
			{Range: "40% - 55%", Label: "Kurang Efektif", Color: "amber"},
			{Range: "< 40%", Label: "Tidak Efektif", Color: "red"},
		},
		Steps: []NGainGuidedStep{
			{Order: 1, Title: "Susun soal pretest", Detail: "Buat soal pretest pada sesi ini. Soal mengukur pemahaman awal, jadi berikan sebelum mahasiswa membuka materi."},
			{Order: 2, Title: "Mahasiswa mengerjakan pretest", Detail: "Sistem menilai otomatis dan menyimpan nilai awal dalam skala 0-100."},
			{Order: 3, Title: "Laksanakan praktikum", Detail: "Materi, tugas, dan kegiatan praktikum berjalan seperti biasa."},
			{Order: 4, Title: "Susun soal post-test", Detail: "Gunakan cakupan materi dan tingkat kesulitan yang setara dengan pretest agar perbandingannya adil."},
			{Order: 5, Title: "Mahasiswa mengerjakan post-test", Detail: "Sistem kembali menilai otomatis dan menyimpan nilai akhir dalam skala yang sama."},
			{Order: 6, Title: "Baca hasil N-Gain", Detail: "Sistem menghitung N-Gain per mahasiswa dan rata-rata kelas beserta kategorinya."},
		},
		Cautions: []string{
			"Pretest dan post-test harus memakai skor maksimum ideal yang sama, di sistem ini 100.",
			"Cakupan materi kedua tes harus setara. Post-test yang jauh lebih mudah akan melambungkan N-Gain secara semu.",
			"Mahasiswa dengan pretest bernilai 100 tidak punya ruang perbaikan sehingga N-Gain tidak terdefinisi dan dikeluarkan dari rata-rata.",
			"N-Gain negatif berarti nilai turun. Periksa kembali pelaksanaan tes sebelum menarik kesimpulan.",
			"Rata-rata N-Gain individu dan N-Gain dari rata-rata kelas adalah dua angka berbeda. Sebutkan yang mana yang dipakai saat melaporkan.",
		},
		Glossary: []NGainGlossaryTerm{
			{Term: "Pretest", Meaning: "Tes sebelum pembelajaran untuk mengukur pemahaman awal."},
			{Term: "Post-test", Meaning: "Tes setelah pembelajaran untuk mengukur pemahaman akhir."},
			{Term: "Skor Maksimum Ideal", Meaning: "Nilai tertinggi yang mungkin dicapai pada tes, di sistem ini 100."},
			{Term: "Gain", Meaning: "Selisih mentah post-test dikurangi pretest."},
			{Term: "N-Gain", Meaning: "Gain yang dinormalisasi terhadap ruang perbaikan yang masih tersisa."},
			{Term: "Kategori Hake", Meaning: "Pengelompokan N-Gain menjadi Tinggi, Sedang, dan Rendah."},
		},
	}
}

// NGainGuide adalah versi terekspor dari materi panduan N-Gain.
func NGainGuide() NGainReference {
	return nGainReference()
}

// SummarizeNGainValues meringkas sekumpulan nilai N-Gain menjadi statistik kelas.
// Dipakai oleh pembuat laporan yang sudah memegang nilai jadi, bukan baris database.
func SummarizeNGainValues(values []NGainValue, maxScore float64) NGainSummary {
	rows := make([]NGainStudentRow, 0, len(values))
	for _, value := range values {
		rows = append(rows, NGainStudentRow{NGain: value})
	}
	return summarizeNGain(rows, maxScore)
}

// summarizeNGain menghitung ringkasan kelas dari sekumpulan baris mahasiswa.
func summarizeNGain(rows []NGainStudentRow, maxScore float64) NGainSummary {
	if maxScore <= 0 {
		maxScore = 100
	}
	summary := NGainSummary{
		StudentCount: len(rows),
		MaxScore:     maxScore,
		Thresholds:   nGainReference(),
	}

	var sumPre, sumPost, sumNGain float64
	counts := map[string]int{"Tinggi": 0, "Sedang": 0, "Rendah": 0}

	for _, row := range rows {
		switch row.NGain.Status {
		case NGainStatusIncomplete:
			summary.IncompleteCnt++
			continue
		case NGainStatusCeiling:
			summary.CeilingCount++
			// Pretest sudah maksimum: nilainya tetap ikut rata-rata skor mentah,
			// tetapi tidak ikut rata-rata N-Gain karena tidak terdefinisi.
			sumPre += row.NGain.Pre
			sumPost += row.NGain.Post
			continue
		}
		summary.ComputedCount++
		sumPre += row.NGain.Pre
		sumPost += row.NGain.Post
		sumNGain += row.NGain.NGain
		counts[row.NGain.Category]++
	}

	scored := summary.ComputedCount + summary.CeilingCount
	if scored > 0 {
		summary.AveragePre = round(sumPre/float64(scored), 2)
		summary.AveragePost = round(sumPost/float64(scored), 2)
	}
	if summary.ComputedCount > 0 {
		summary.AverageNGain = round(sumNGain/float64(summary.ComputedCount), 4)
		summary.AveragePercent = round(summary.AverageNGain*100, 2)
		summary.Category, summary.CategoryColor = nGainCategory(summary.AverageNGain)
		summary.Effectiveness = nGainEffectiveness(summary.AveragePercent)
		summary.Interpretation = fmt.Sprintf(
			"Rata-rata N-Gain kelas %s termasuk kategori %s dan setara %s%% sehingga pembelajaran dinilai %s.",
			formatID(summary.AverageNGain, 2), summary.Category, formatID(summary.AveragePercent, 2), strings.ToLower(summary.Effectiveness))
	} else {
		summary.Category = "Belum Ada Data"
		summary.CategoryColor = "gray"
		summary.Effectiveness = "Belum dapat dinilai"
		summary.Interpretation = "Belum ada mahasiswa yang menyelesaikan pretest dan post-test sekaligus, jadi N-Gain kelas belum bisa dihitung."
	}

	if denominator := maxScore - summary.AveragePre; denominator > 0 && scored > 0 {
		summary.NGainOfAverages = round((summary.AveragePost-summary.AveragePre)/denominator, 4)
	}

	for _, bucket := range []struct{ label, color, span string }{
		{"Tinggi", "green", "g >= 0,70"},
		{"Sedang", "amber", "0,30 <= g < 0,70"},
		{"Rendah", "red", "g < 0,30"},
	} {
		count := counts[bucket.label]
		percent := 0.0
		if summary.ComputedCount > 0 {
			percent = round(float64(count)*100/float64(summary.ComputedCount), 2)
		}
		summary.Distribution = append(summary.Distribution, NGainBucket{
			Category: bucket.label, Color: bucket.color, Range: bucket.span, Count: count, Percent: percent,
		})
	}
	return summary
}

// nullableScore mengubah kolom nilai yang bisa NULL menjadi pointer float64.
// Nilai hanya dianggap ada bila statusnya completed.
func nullableScore(score sql.NullFloat64, status sql.NullString) *float64 {
	if !score.Valid {
		return nil
	}
	if status.Valid && status.String != "completed" {
		return nil
	}
	value := score.Float64
	return &value
}

// SessionNGain menghitung N-Gain seluruh mahasiswa pada satu sesi praktikum.
func (r *Repository) SessionNGain(sessionID int) (json.RawMessage, error) {
	rows, err := r.db.Query(`
		SELECT st.id, st.student_id, st.name,
			pre.score, pre.status, post.score, post.status,
			COALESCE(GREATEST(pre.max_score, post.max_score), 100)
		FROM course_sessions cs
		JOIN courses c ON c.id = cs.course_id
		JOIN classes cls ON cls.code = c.class_code
		JOIN students st ON st.id = ANY(cls.students)
		LEFT JOIN session_assessment_results pre
			ON pre.session_id = cs.id AND pre.student_id = st.id AND pre.assessment_type = 'pretest'
		LEFT JOIN session_assessment_results post
			ON post.session_id = cs.id AND post.student_id = st.id AND post.assessment_type = 'posttest'
		WHERE cs.id = $1
		ORDER BY st.student_id
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := []NGainStudentRow{}
	maxScore := 100.0
	for rows.Next() {
		var (
			studentID             int
			nim, name             string
			preScore, postScore   sql.NullFloat64
			preStatus, postStatus sql.NullString
			rowMaxScore           float64
		)
		if err := rows.Scan(&studentID, &nim, &name, &preScore, &preStatus, &postScore, &postStatus, &rowMaxScore); err != nil {
			return nil, err
		}
		if rowMaxScore > 0 {
			maxScore = rowMaxScore
		}
		value := ComputeNGain(nullableScore(preScore, preStatus), nullableScore(postScore, postStatus), rowMaxScore)
		students = append(students, NGainStudentRow{
			StudentID:         studentID,
			NIM:               nim,
			Name:              name,
			NGain:             value,
			ImprovementPoints: round(value.Gain, 2),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var sessionNumber int
	var sessionTitle string
	_ = r.db.QueryRow(`SELECT session_number, topic FROM course_sessions WHERE id=$1`, sessionID).
		Scan(&sessionNumber, &sessionTitle)

	payload := map[string]interface{}{
		"scope":         "session",
		"sessionId":     sessionID,
		"sessionNumber": sessionNumber,
		"sessionTitle":  sessionTitle,
		"students":      students,
		"summary":       summarizeNGain(students, maxScore),
		"guide":         nGainReference(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// CourseNGain menghitung N-Gain tingkat mata kuliah. Nilai pretest dan post-test
// tiap mahasiswa dirata-ratakan lebih dulu dari seluruh sesi yang sudah selesai,
// lalu rumus Hake diterapkan pada rata-rata tersebut.
func (r *Repository) CourseNGain(courseID int) (json.RawMessage, error) {
	rows, err := r.db.Query(`
		SELECT st.id, st.student_id, st.name,
			AVG(sar.score) FILTER (WHERE sar.assessment_type='pretest' AND sar.status='completed'),
			AVG(sar.score) FILTER (WHERE sar.assessment_type='posttest' AND sar.status='completed'),
			COUNT(*) FILTER (WHERE sar.assessment_type='pretest' AND sar.status='completed'),
			COUNT(*) FILTER (WHERE sar.assessment_type='posttest' AND sar.status='completed'),
			COALESCE(MAX(sar.max_score), 100)
		FROM courses c
		JOIN classes cls ON cls.code = c.class_code
		JOIN students st ON st.id = ANY(cls.students)
		LEFT JOIN course_sessions cs ON cs.course_id = c.id
		LEFT JOIN session_assessment_results sar
			ON sar.session_id = cs.id AND sar.student_id = st.id
		WHERE c.id = $1
		GROUP BY st.id, st.student_id, st.name
		ORDER BY st.student_id
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := []NGainStudentRow{}
	maxScore := 100.0
	for rows.Next() {
		var (
			studentID           int
			nim, name           string
			preAvg, postAvg     sql.NullFloat64
			preCount, postCount int
			rowMaxScore         float64
		)
		if err := rows.Scan(&studentID, &nim, &name, &preAvg, &postAvg, &preCount, &postCount, &rowMaxScore); err != nil {
			return nil, err
		}
		if rowMaxScore > 0 {
			maxScore = rowMaxScore
		}
		var pre, post *float64
		if preAvg.Valid {
			rounded := round(preAvg.Float64, 2)
			pre = &rounded
		}
		if postAvg.Valid {
			rounded := round(postAvg.Float64, 2)
			post = &rounded
		}
		value := ComputeNGain(pre, post, rowMaxScore)
		students = append(students, NGainStudentRow{
			StudentID:         studentID,
			NIM:               nim,
			Name:              name,
			PretestSessions:   preCount,
			PosttestSessions:  postCount,
			NGain:             value,
			ImprovementPoints: round(value.Gain, 2),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sessions, err := r.courseSessionNGainSummaries(courseID)
	if err != nil {
		return nil, err
	}

	var courseName, classCode string
	_ = r.db.QueryRow(`SELECT name, class_code FROM courses WHERE id=$1`, courseID).Scan(&courseName, &classCode)

	payload := map[string]interface{}{
		"scope":      "course",
		"courseId":   courseID,
		"courseName": courseName,
		"classCode":  classCode,
		"students":   students,
		"summary":    summarizeNGain(students, maxScore),
		"sessions":   sessions,
		"guide":      nGainReference(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// courseSessionNGainSummaries membuat ringkasan N-Gain untuk tiap sesi sebuah
// mata kuliah, dipakai untuk melihat sesi mana yang paling berdampak.
func (r *Repository) courseSessionNGainSummaries(courseID int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT cs.id, cs.session_number, cs.topic, st.id,
			pre.score, pre.status, post.score, post.status,
			COALESCE(GREATEST(pre.max_score, post.max_score), 100)
		FROM course_sessions cs
		JOIN courses c ON c.id = cs.course_id
		JOIN classes cls ON cls.code = c.class_code
		JOIN students st ON st.id = ANY(cls.students)
		LEFT JOIN session_assessment_results pre
			ON pre.session_id = cs.id AND pre.student_id = st.id AND pre.assessment_type = 'pretest'
		LEFT JOIN session_assessment_results post
			ON post.session_id = cs.id AND post.student_id = st.id AND post.assessment_type = 'posttest'
		WHERE c.id = $1
		ORDER BY cs.session_number, st.student_id
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type sessionBucket struct {
		number   int
		title    string
		maxScore float64
		rows     []NGainStudentRow
	}
	order := []int{}
	buckets := map[int]*sessionBucket{}

	for rows.Next() {
		var (
			sessionID, sessionNumber, studentID int
			title                               string
			preScore, postScore                 sql.NullFloat64
			preStatus, postStatus               sql.NullString
			rowMaxScore                         float64
		)
		if err := rows.Scan(&sessionID, &sessionNumber, &title, &studentID,
			&preScore, &preStatus, &postScore, &postStatus, &rowMaxScore); err != nil {
			return nil, err
		}
		bucket, ok := buckets[sessionID]
		if !ok {
			bucket = &sessionBucket{number: sessionNumber, title: title, maxScore: rowMaxScore}
			buckets[sessionID] = bucket
			order = append(order, sessionID)
		}
		value := ComputeNGain(nullableScore(preScore, preStatus), nullableScore(postScore, postStatus), rowMaxScore)
		bucket.rows = append(bucket.rows, NGainStudentRow{StudentID: studentID, NGain: value})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(order))
	for _, sessionID := range order {
		bucket := buckets[sessionID]
		items = append(items, map[string]interface{}{
			"sessionId":     sessionID,
			"sessionNumber": bucket.number,
			"sessionTitle":  bucket.title,
			"summary":       summarizeNGain(bucket.rows, bucket.maxScore),
		})
	}
	return items, nil
}

// StudentSessionNGain menghitung N-Gain satu mahasiswa pada satu sesi.
func (r *Repository) StudentSessionNGain(sessionID int, studentID int) (NGainValue, error) {
	var (
		preScore, postScore   sql.NullFloat64
		preStatus, postStatus sql.NullString
		maxScore              float64
	)
	err := r.db.QueryRow(`
		SELECT pre.score, pre.status, post.score, post.status,
			COALESCE(GREATEST(pre.max_score, post.max_score), 100)
		FROM (SELECT 1) dummy
		LEFT JOIN session_assessment_results pre
			ON pre.session_id = $1 AND pre.student_id = $2 AND pre.assessment_type = 'pretest'
		LEFT JOIN session_assessment_results post
			ON post.session_id = $1 AND post.student_id = $2 AND post.assessment_type = 'posttest'
	`, sessionID, studentID).Scan(&preScore, &preStatus, &postScore, &postStatus, &maxScore)
	if err != nil && err != sql.ErrNoRows {
		return NGainValue{}, err
	}
	return ComputeNGain(nullableScore(preScore, preStatus), nullableScore(postScore, postStatus), maxScore), nil
}
