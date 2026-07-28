package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"studi-ku-backend/internal/models"
)

// Satu bank soal melayani dua jenis tes. Kolom assessment_type pada
// session_pretest_questions dan session_pretest_submissions yang membedakan,
// sehingga pretest dan post-test dinilai oleh mesin yang sama persis. Itu
// syarat penting supaya perbandingan N-Gain adil.

const (
	quizTypePretest  = "pretest"
	quizTypePosttest = "posttest"
)

// normalizeQuizType memvalidasi jenis tes dan mengisi nilai bawaan pretest.
func normalizeQuizType(quizType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(quizType)) {
	case "", quizTypePretest:
		return quizTypePretest, nil
	case quizTypePosttest, "post-test", "post_test":
		return quizTypePosttest, nil
	default:
		return "", errors.New("jenis tes harus pretest atau posttest")
	}
}

// quizTitle mengembalikan judul yang ditampilkan untuk sebuah jenis tes.
func quizTitle(quizType string) string {
	if quizType == quizTypePosttest {
		return "Post-test"
	}
	return "Pretest"
}

// counterpartType mengembalikan jenis tes pasangannya.
func counterpartType(quizType string) string {
	if quizType == quizTypePosttest {
		return quizTypePretest
	}
	return quizTypePosttest
}

// SessionQuiz menyusun seluruh data satu jenis tes pada sebuah sesi: daftar soal,
// jawaban mahasiswa yang bersangkutan, rekap kelas untuk aslab, serta N-Gain bila
// pretest dan post-test sudah sama-sama selesai.
func (r *Repository) SessionQuiz(sessionID int, studentID int, role string, quizType string) (json.RawMessage, error) {
	quizType, err := normalizeQuizType(quizType)
	if err != nil {
		return nil, err
	}

	questions, err := r.quizQuestions(sessionID, role, quizType)
	if err != nil {
		return nil, err
	}

	submission, err := r.quizSubmission(sessionID, studentID, quizType)
	if err != nil {
		return nil, err
	}

	results := []map[string]interface{}{}
	if role != "student" {
		results, err = r.quizResults(sessionID, quizType)
		if err != nil {
			return nil, err
		}
	}

	other := counterpartType(quizType)
	otherCount, err := r.quizQuestionCount(sessionID, other)
	if err != nil {
		return nil, err
	}
	otherSubmission, err := r.quizSubmission(sessionID, studentID, other)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"questionCount":  len(questions),
		"submittedCount": len(results),
	}

	payload := map[string]interface{}{
		"sessionId":  sessionID,
		"type":       quizType,
		"title":      quizTitle(quizType),
		"questions":  questions,
		"submission": submission,
		"results":    results,
		"summary":    summary,
		"counterpart": map[string]interface{}{
			"type":          other,
			"title":         quizTitle(other),
			"questionCount": otherCount,
			"status":        otherSubmission["status"],
			"score":         otherSubmission["score"],
		},
		"guide": nGainReference(),
	}

	// N-Gain pribadi hanya bermakna bila mahasiswa yang diminta memang ada.
	if studentID > 0 {
		value, err := r.StudentSessionNGain(sessionID, studentID)
		if err != nil {
			return nil, err
		}
		payload["ngain"] = value
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// CreateQuizQuestion menambah satu soal pada jenis tes tertentu.
func (r *Repository) CreateQuizQuestion(sessionID int, createdBy int, quizType string, payload *models.QuizQuestionInput) (map[string]interface{}, error) {
	quizType, err := normalizeQuizType(quizType)
	if err != nil {
		return nil, err
	}
	options, err := json.Marshal(payload.Options)
	if err != nil {
		return nil, err
	}
	points := payload.Points
	if points <= 0 {
		points = 10
	}
	var id int
	err = r.db.QueryRow(`
		INSERT INTO session_pretest_questions (session_id,assessment_type,question,options,correct_option,points,explanation,sort_order,created_by,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,0),now())
		RETURNING id
	`, sessionID, quizType, payload.Question, options, payload.CorrectOption, points, payload.Explanation, payload.SortOrder, createdBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": id, "sessionId": sessionID, "type": quizType, "question": payload.Question, "options": payload.Options,
		"correctOption": payload.CorrectOption, "points": points, "explanation": payload.Explanation, "sortOrder": payload.SortOrder,
	}, nil
}

// UpdateQuizQuestion mengubah satu soal. Jenis tes ikut diperbarui supaya soal
// bisa dipindahkan dari pretest ke post-test bila diperlukan.
func (r *Repository) UpdateQuizQuestion(questionID int, quizType string, payload *models.QuizQuestionInput) (map[string]interface{}, error) {
	quizType, err := normalizeQuizType(quizType)
	if err != nil {
		return nil, err
	}
	options, err := json.Marshal(payload.Options)
	if err != nil {
		return nil, err
	}
	points := payload.Points
	if points <= 0 {
		points = 10
	}
	var sessionID int
	err = r.db.QueryRow(`
		UPDATE session_pretest_questions
		SET question=$1, options=$2, correct_option=$3, points=$4, explanation=$5, sort_order=$6, assessment_type=$7, updated_at=now()
		WHERE id=$8
		RETURNING session_id
	`, payload.Question, options, payload.CorrectOption, points, payload.Explanation, payload.SortOrder, quizType, questionID).Scan(&sessionID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": questionID, "sessionId": sessionID, "type": quizType, "question": payload.Question, "options": payload.Options,
		"correctOption": payload.CorrectOption, "points": points, "explanation": payload.Explanation, "sortOrder": payload.SortOrder,
	}, nil
}

// DeleteQuizQuestion menghapus satu soal apa pun jenis tesnya.
func (r *Repository) DeleteQuizQuestion(questionID int) error {
	_, err := r.db.Exec(`DELETE FROM session_pretest_questions WHERE id=$1`, questionID)
	return err
}

// SubmitQuiz menilai jawaban mahasiswa di sisi server lalu menyimpannya ke tabel
// jawaban, tabel penanda sesi, dan tabel hasil per mahasiswa yang dibaca N-Gain.
func (r *Repository) SubmitQuiz(sessionID int, studentID int, quizType string, payload *models.QuizSubmissionInput) (map[string]interface{}, error) {
	quizType, err := normalizeQuizType(quizType)
	if err != nil {
		return nil, err
	}
	if studentID <= 0 {
		return nil, errors.New("student wajib dipilih")
	}

	questions, err := r.loadQuizQuestions(sessionID, quizType)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, errors.New("belum ada soal " + quizTitle(quizType))
	}

	// Post-test hanya bermakna sebagai pembanding, jadi pretest wajib lebih dulu.
	if quizType == quizTypePosttest {
		prior, err := r.quizSubmission(sessionID, studentID, quizTypePretest)
		if err != nil {
			return nil, err
		}
		if prior["status"] != "completed" {
			return nil, errors.New("kerjakan pretest terlebih dahulu agar N-Gain dapat dihitung")
		}
	}

	answerMap := map[int]int{}
	for _, answer := range payload.Answers {
		answerMap[answer.QuestionID] = answer.AnswerIndex
	}

	totalPoints := 0
	scorePoints := 0
	for _, question := range questions {
		points := question.Points
		if points <= 0 {
			points = 10
		}
		totalPoints += points
		if selected, ok := answerMap[question.ID]; ok && question.CorrectOption != nil && selected == *question.CorrectOption {
			scorePoints += points
		}
	}
	if totalPoints <= 0 {
		totalPoints = len(questions)
	}
	score := 0
	if totalPoints > 0 {
		score = int(float64(scorePoints) * 100 / float64(totalPoints))
	}

	answersJSON, err := json.Marshal(payload.Answers)
	if err != nil {
		return nil, err
	}

	var submissionID int
	err = r.db.QueryRow(`
		INSERT INTO session_pretest_submissions (session_id,student_id,assessment_type,answers,score,max_score,status,submitted_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,100,'completed',now(),now())
		ON CONFLICT (session_id,student_id,assessment_type) DO UPDATE SET
			answers=EXCLUDED.answers,
			score=EXCLUDED.score,
			max_score=EXCLUDED.max_score,
			status='completed',
			submitted_at=COALESCE(session_pretest_submissions.submitted_at, now()),
			updated_at=now()
		RETURNING id
	`, sessionID, studentID, quizType, answersJSON, score).Scan(&submissionID)
	if err != nil {
		return nil, err
	}

	if _, err := r.db.Exec(`
		INSERT INTO session_assessments (session_id,assessment_type,title,score,max_score,status,note,updated_at)
		VALUES ($1,$2,$3,NULL,100,'completed','',now())
		ON CONFLICT (session_id,assessment_type) DO UPDATE SET
			title=EXCLUDED.title,
			max_score=EXCLUDED.max_score,
			status='completed',
			note='',
			updated_at=now()
	`, sessionID, quizType, quizTitle(quizType)); err != nil {
		return nil, err
	}

	if _, err := r.db.Exec(`
		INSERT INTO session_assessment_results (session_id,student_id,assessment_type,score,max_score,status,note,submitted_at,updated_at)
		VALUES ($1,$2,$3,$4,100,'completed','',now(),now())
		ON CONFLICT (session_id,student_id,assessment_type) DO UPDATE SET
			score=EXCLUDED.score,
			max_score=EXCLUDED.max_score,
			status='completed',
			note='',
			submitted_at=COALESCE(session_assessment_results.submitted_at, now()),
			updated_at=now()
	`, sessionID, studentID, quizType, score); err != nil {
		return nil, err
	}

	value, err := r.StudentSessionNGain(sessionID, studentID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id": submissionID, "sessionId": sessionID, "studentId": studentID, "type": quizType,
		"score": score, "maxScore": 100, "status": "completed", "answers": payload.Answers,
		"ngain": value,
	}, nil
}

// quizQuestions menyusun daftar soal untuk ditampilkan. Kunci jawaban dan
// pembahasan disembunyikan dari mahasiswa.
func (r *Repository) quizQuestions(sessionID int, role string, quizType string) ([]map[string]interface{}, error) {
	questions, err := r.loadQuizQuestions(sessionID, quizType)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(questions))
	for _, question := range questions {
		item := map[string]interface{}{
			"id": question.ID, "sessionId": question.SessionID, "type": quizType, "question": question.Question,
			"options": question.Options, "points": question.Points, "sortOrder": question.SortOrder,
		}
		if role != "student" {
			correct := question.CorrectOption
			item["correctOption"] = correct
			item["explanation"] = question.Explanation
		}
		items = append(items, item)
	}
	return items, nil
}

// loadQuizQuestions membaca soal satu jenis tes dari database.
func (r *Repository) loadQuizQuestions(sessionID int, quizType string) ([]models.QuizQuestion, error) {
	rows, err := r.db.Query(`
		SELECT id,session_id,question,options,correct_option,points,explanation,sort_order
		FROM session_pretest_questions
		WHERE session_id=$1 AND assessment_type=$2
		ORDER BY sort_order, id
	`, sessionID, quizType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.QuizQuestion{}
	for rows.Next() {
		var question models.QuizQuestion
		var optionsBytes []byte
		var correctOption int
		if err := rows.Scan(&question.ID, &question.SessionID, &question.Question, &optionsBytes, &correctOption, &question.Points, &question.Explanation, &question.SortOrder); err != nil {
			return nil, err
		}
		question.CorrectOption = &correctOption
		question.Type = quizType
		if len(optionsBytes) > 0 {
			if err := json.Unmarshal(optionsBytes, &question.Options); err != nil {
				return nil, err
			}
		}
		items = append(items, question)
	}
	return items, rows.Err()
}

// quizQuestionCount menghitung jumlah soal satu jenis tes.
func (r *Repository) quizQuestionCount(sessionID int, quizType string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT count(*) FROM session_pretest_questions WHERE session_id=$1 AND assessment_type=$2
	`, sessionID, quizType).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// quizSubmission membaca jawaban satu mahasiswa. Bila belum ada, dikembalikan
// bentuk kosong agar antarmuka tidak perlu menangani nil.
func (r *Repository) quizSubmission(sessionID int, studentID int, quizType string) (map[string]interface{}, error) {
	empty := map[string]interface{}{
		"status": "not_started", "score": 0, "maxScore": 100, "type": quizType, "answers": []models.QuizAnswerInput{},
	}
	if studentID <= 0 {
		return empty, nil
	}

	var (
		id          int
		score       int
		maxScore    int
		status      string
		submittedAt sql.NullTime
		answersRaw  []byte
	)
	err := r.db.QueryRow(`
		SELECT id,score,max_score,status,submitted_at,answers
		FROM session_pretest_submissions
		WHERE session_id=$1 AND student_id=$2 AND assessment_type=$3
	`, sessionID, studentID, quizType).Scan(&id, &score, &maxScore, &status, &submittedAt, &answersRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return empty, nil
	}
	if err != nil {
		return nil, err
	}

	answers := []models.QuizAnswerInput{}
	if len(answersRaw) > 0 {
		if err := json.Unmarshal(answersRaw, &answers); err != nil {
			return nil, err
		}
	}
	result := map[string]interface{}{
		"id": id, "sessionId": sessionID, "studentId": studentID, "type": quizType,
		"score": score, "maxScore": maxScore, "status": status, "answers": answers,
	}
	if submittedAt.Valid {
		result["submittedAt"] = submittedAt.Time.Format(timeLayout())
	}
	return result, nil
}

// quizResults merekap jawaban seluruh mahasiswa pada satu jenis tes.
func (r *Repository) quizResults(sessionID int, quizType string) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.student_id, st.student_id, st.name, s.score, s.max_score, s.status, COALESCE(to_char(s.submitted_at, 'YYYY-MM-DD HH24:MI'), '')
		FROM session_pretest_submissions s
		JOIN students st ON st.id=s.student_id
		WHERE s.session_id=$1 AND s.assessment_type=$2
		ORDER BY st.student_id
	`, sessionID, quizType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []map[string]interface{}{}
	for rows.Next() {
		var id, studentDBID, score, maxScore int
		var nim, name, status, submittedAt string
		if err := rows.Scan(&id, &studentDBID, &nim, &name, &score, &maxScore, &status, &submittedAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id": id, "studentId": studentDBID, "nim": nim, "name": name, "type": quizType,
			"score": score, "maxScore": maxScore, "status": status, "submittedAt": submittedAt,
		})
	}
	return items, rows.Err()
}

func timeLayout() string {
	return "2006-01-02 15:04"
}

// Pembungkus lama supaya rute /pretest yang sudah dipakai klien tetap hidup.

func (r *Repository) SessionPretest(sessionID int, studentID int, role string) (json.RawMessage, error) {
	return r.SessionQuiz(sessionID, studentID, role, quizTypePretest)
}

func (r *Repository) CreatePretestQuestion(sessionID int, createdBy int, payload *models.QuizQuestionInput) (map[string]interface{}, error) {
	return r.CreateQuizQuestion(sessionID, createdBy, quizTypePretest, payload)
}

func (r *Repository) UpdatePretestQuestion(questionID int, payload *models.QuizQuestionInput) (map[string]interface{}, error) {
	return r.UpdateQuizQuestion(questionID, quizTypePretest, payload)
}

func (r *Repository) DeletePretestQuestion(questionID int) error {
	return r.DeleteQuizQuestion(questionID)
}

func (r *Repository) SubmitPretest(sessionID int, studentID int, payload *models.QuizSubmissionInput) (map[string]interface{}, error) {
	return r.SubmitQuiz(sessionID, studentID, quizTypePretest, payload)
}
