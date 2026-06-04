package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"

	"studi-ku-backend/internal/models"
)

func (r *Repository) SessionPretest(sessionID int, studentID int, role string) (json.RawMessage, error) {
	questions, err := r.pretestQuestions(sessionID, role)
	if err != nil {
		return nil, err
	}

	submission, err := r.pretestSubmission(sessionID, studentID)
	if err != nil {
		return nil, err
	}

	results := []map[string]interface{}{}
	if role != "student" {
		results, err = r.pretestResults(sessionID)
		if err != nil {
			return nil, err
		}
	}

	summary := map[string]interface{}{
		"questionCount":  len(questions),
		"submittedCount": len(results),
	}

	payload := map[string]interface{}{
		"sessionId":  sessionID,
		"questions":  questions,
		"submission": submission,
		"results":    results,
		"summary":    summary,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func (r *Repository) CreatePretestQuestion(sessionID int, createdBy int, payload *models.PretestQuestionInput) (map[string]interface{}, error) {
	options, err := json.Marshal(payload.Options)
	if err != nil {
		return nil, err
	}
	points := payload.Points
	if points <= 0 {
		points = 10
	}
	sortOrder := payload.SortOrder
	var id int
	err = r.db.QueryRow(`
		INSERT INTO session_pretest_questions (session_id,question,options,correct_option,points,explanation,sort_order,created_by,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NULLIF($8,0),now())
		RETURNING id
	`, sessionID, payload.Question, options, payload.CorrectOption, points, payload.Explanation, sortOrder, createdBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": id, "sessionId": sessionID, "question": payload.Question, "options": payload.Options,
		"correctOption": payload.CorrectOption, "points": points, "explanation": payload.Explanation, "sortOrder": sortOrder,
	}, nil
}

func (r *Repository) UpdatePretestQuestion(questionID int, payload *models.PretestQuestionInput) (map[string]interface{}, error) {
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
		SET question=$1, options=$2, correct_option=$3, points=$4, explanation=$5, sort_order=$6, updated_at=now()
		WHERE id=$7
		RETURNING session_id
	`, payload.Question, options, payload.CorrectOption, points, payload.Explanation, payload.SortOrder, questionID).Scan(&sessionID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": questionID, "sessionId": sessionID, "question": payload.Question, "options": payload.Options,
		"correctOption": payload.CorrectOption, "points": points, "explanation": payload.Explanation, "sortOrder": payload.SortOrder,
	}, nil
}

func (r *Repository) DeletePretestQuestion(questionID int) error {
	_, err := r.db.Exec(`DELETE FROM session_pretest_questions WHERE id=$1`, questionID)
	return err
}

func (r *Repository) SubmitPretest(sessionID int, studentID int, payload *models.PretestSubmissionInput) (map[string]interface{}, error) {
	if studentID <= 0 {
		return nil, errors.New("student wajib dipilih")
	}

	questions, err := r.loadPretestQuestions(sessionID)
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, errors.New("belum ada soal pretest")
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
		INSERT INTO session_pretest_submissions (session_id,student_id,answers,score,max_score,status,submitted_at,updated_at)
		VALUES ($1,$2,$3,$4,100,'completed',now(),now())
		ON CONFLICT (session_id,student_id) DO UPDATE SET
			answers=EXCLUDED.answers,
			score=EXCLUDED.score,
			max_score=EXCLUDED.max_score,
			status='completed',
			submitted_at=COALESCE(session_pretest_submissions.submitted_at, now()),
			updated_at=now()
		RETURNING id
	`, sessionID, studentID, answersJSON, score).Scan(&submissionID)
	if err != nil {
		return nil, err
	}

	if _, err := r.db.Exec(`
		INSERT INTO session_assessments (session_id,assessment_type,title,score,max_score,status,note,updated_at)
		VALUES ($1,'pretest','Pretest',NULL,100,'completed','',now())
		ON CONFLICT (session_id,assessment_type) DO UPDATE SET
			title='Pretest',
			max_score=EXCLUDED.max_score,
			status='completed',
			note='',
			updated_at=now()
	`, sessionID); err != nil {
		return nil, err
	}

	if _, err := r.db.Exec(`
		INSERT INTO session_assessment_results (session_id,student_id,assessment_type,score,max_score,status,note,submitted_at,updated_at)
		VALUES ($1,$2,'pretest',$3,100,'completed','',now(),now())
		ON CONFLICT (session_id,student_id,assessment_type) DO UPDATE SET
			score=EXCLUDED.score,
			max_score=EXCLUDED.max_score,
			status='completed',
			note='',
			submitted_at=COALESCE(session_assessment_results.submitted_at, now()),
			updated_at=now()
	`, sessionID, studentID, score); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id": submissionID, "sessionId": sessionID, "studentId": studentID,
		"score": score, "maxScore": 100, "status": "completed", "answers": payload.Answers,
	}, nil
}

func (r *Repository) pretestQuestions(sessionID int, role string) ([]map[string]interface{}, error) {
	questions, err := r.loadPretestQuestions(sessionID)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(questions))
	for _, question := range questions {
		item := map[string]interface{}{
			"id": question.ID, "sessionId": question.SessionID, "question": question.Question,
			"options": question.Options, "points": question.Points, "explanation": question.Explanation, "sortOrder": question.SortOrder,
		}
		if role != "student" {
			correct := question.CorrectOption
			item["correctOption"] = &correct
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) loadPretestQuestions(sessionID int) ([]models.PretestQuestion, error) {
	rows, err := r.db.Query(`
		SELECT id,session_id,question,options,correct_option,points,explanation,sort_order
		FROM session_pretest_questions
		WHERE session_id=$1
		ORDER BY sort_order, id
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.PretestQuestion{}
	for rows.Next() {
		var question models.PretestQuestion
		var optionsBytes []byte
		var correctOption int
		if err := rows.Scan(&question.ID, &question.SessionID, &question.Question, &optionsBytes, &correctOption, &question.Points, &question.Explanation, &question.SortOrder); err != nil {
			return nil, err
		}
		question.CorrectOption = &correctOption
		if len(optionsBytes) > 0 {
			if err := json.Unmarshal(optionsBytes, &question.Options); err != nil {
				return nil, err
			}
		}
		items = append(items, question)
	}
	return items, rows.Err()
}

func (r *Repository) pretestSubmission(sessionID int, studentID int) (map[string]interface{}, error) {
	if studentID <= 0 {
		return map[string]interface{}{
			"status": "not_started", "score": 0, "maxScore": 100, "answers": []models.PretestAnswerInput{},
		}, nil
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
		WHERE session_id=$1 AND student_id=$2
	`, sessionID, studentID).Scan(&id, &score, &maxScore, &status, &submittedAt, &answersRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return map[string]interface{}{
			"status": "not_started", "score": 0, "maxScore": 100, "answers": []models.PretestAnswerInput{},
		}, nil
	}
	if err != nil {
		return nil, err
	}

	answers := []models.PretestAnswerInput{}
	if len(answersRaw) > 0 {
		if err := json.Unmarshal(answersRaw, &answers); err != nil {
			return nil, err
		}
	}
	result := map[string]interface{}{
		"id": id, "sessionId": sessionID, "studentId": studentID,
		"score": score, "maxScore": maxScore, "status": status, "answers": answers,
	}
	if submittedAt.Valid {
		result["submittedAt"] = submittedAt.Time.Format(timeLayout())
	}
	return result, nil
}

func (r *Repository) pretestResults(sessionID int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.student_id, st.student_id, st.name, s.score, s.max_score, s.status, COALESCE(to_char(s.submitted_at, 'YYYY-MM-DD HH24:MI'), '')
		FROM session_pretest_submissions s
		JOIN students st ON st.id=s.student_id
		WHERE s.session_id=$1
		ORDER BY st.student_id
	`, sessionID)
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
			"id": id, "studentId": studentDBID, "nim": nim, "name": name,
			"score": score, "maxScore": maxScore, "status": status, "submittedAt": submittedAt,
		})
	}
	return items, rows.Err()
}

func timeLayout() string {
	return "2006-01-02 15:04"
}
