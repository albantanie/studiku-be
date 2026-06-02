package repositories

import (
	"database/sql"
	"encoding/json"
)

func (r *Repository) PageData(key string) (json.RawMessage, error) {
	query, ok := pageDataQueries[key]
	if !ok {
		return nil, sql.ErrNoRows
	}
	if key == "assistant_attendance" {
		if err := r.syncAssistantAttendanceSessions(); err != nil {
			return nil, err
		}
	}
	if key == "assistant_practicals" {
		if err := r.syncAllCourseSessions(); err != nil {
			return nil, err
		}
		if err := r.syncAssistantAttendanceSessions(); err != nil {
			return nil, err
		}
	}

	var payload []byte
	if err := r.db.QueryRow(query).Scan(&payload); err != nil {
		return nil, err
	}
	return json.RawMessage(payload), nil
}

func (r *Repository) StudentCourseDetail(courseID int) (json.RawMessage, error) {
	var payload []byte
	if err := r.db.QueryRow(`
		SELECT json_build_object(
			'materials', COALESCE((
				SELECT json_agg(json_build_object(
					'id',id,'title',title,'week',COALESCE(week,0),'status',status,'fileUrl','/api/materials/' || id || '/file'
				) ORDER BY id)
				FROM course_materials
				WHERE course_id=$1 AND deleted_at IS NULL AND status IN ('submitted','available')
			), '[]'::json),
			'courseSessions', COALESCE((
				SELECT json_agg(json_build_object(
					'id',id,'sessionName',title,'topic',topic,'date',session_date,'time',session_time,
					'type',session_type,'conferenceLink',conference_link,
					'progress', (
						(CASE WHEN EXISTS (SELECT 1 FROM attendance_records ar JOIN attendance_sessions ats ON ats.id=ar.session_id WHERE ats.course_code=(SELECT class_code FROM courses WHERE id=$1) AND ats.session_number=course_sessions.session_number) THEN 33 ELSE 0 END) +
						(CASE WHEN EXISTS (SELECT 1 FROM course_materials cm WHERE cm.session_id=course_sessions.id AND cm.deleted_at IS NULL AND cm.status IN ('submitted','available')) THEN 33 ELSE 0 END) +
						(CASE WHEN EXISTS (SELECT 1 FROM session_assignments sa WHERE sa.session_id=course_sessions.id AND sa.status IN ('submitted','graded','Selesai')) THEN 34 ELSE 0 END)
					)
				) ORDER BY session_number)
				FROM course_sessions
				WHERE course_id=$1
			), '[]'::json),
			'assignments', COALESCE((
				SELECT json_agg(json_build_object('id',id,'title',title,'dueDate',due_date,'status',status,'score',score) ORDER BY id)
				FROM session_assignments
				WHERE course_id=$1
			), '[]'::json)
		)
	`, courseID).Scan(&payload); err != nil {
		return nil, err
	}
	return json.RawMessage(payload), nil
}

func (r *Repository) StudentSessionDetail(sessionID int) (json.RawMessage, error) {
	var payload []byte
	if err := r.db.QueryRow(`
		SELECT json_build_object(
			'materials', COALESCE((
				SELECT json_agg(json_build_object('id',id,'title',title,'type',material_type,'size',size,'fileUrl','/api/materials/' || id || '/file') ORDER BY id)
				FROM course_materials
				WHERE session_id=$1 AND deleted_at IS NULL AND status IN ('submitted','available')
			), '[]'::json),
			'assignments', COALESCE((
				SELECT json_agg(json_build_object('id',id,'title',title,'deadline',due_date,'status',status) ORDER BY id)
				FROM session_assignments
				WHERE session_id=$1
			), '[]'::json)
		)
	`, sessionID).Scan(&payload); err != nil {
		return nil, err
	}
	return json.RawMessage(payload), nil
}

var pageDataQueries = map[string]string{
	"student_grade_semesters": `SELECT COALESCE(json_agg(json_build_object('id', value, 'name', label) ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='student_grade_semesters'`,
	"admin_student_programs":  `SELECT COALESCE(json_agg(value ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='student_programs'`,
	"admin_labs":              `SELECT COALESCE(json_agg(value ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='labs'`,
	"admin_user_tabs":         `SELECT COALESCE(json_agg(json_build_object('id', value, 'label', label, 'icon', icon) ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='user_tabs'`,
	"admin_import_file_types": `SELECT COALESCE(json_agg(value ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='import_file_types'`,
	"admin_import_lecturer_preview": `
		SELECT COALESCE(json_agg(json_build_object(
			'name', l.name, 'email', l.email, 'password', l.password, 'nidn', l.nidn,
			'courses', COALESCE((SELECT json_agg(course_name ORDER BY sort_order) FROM import_lecturer_preview_courses WHERE lecturer_email=l.email), '[]'::json)
		)), '[]'::json)
		FROM lecturers l WHERE l.email='imported@university.ac.id'`,
	"admin_dashboard": `
		SELECT json_build_object(
			'stats', json_build_array(
				json_build_object('label','Total Mahasiswa','value',(SELECT count(*)::text FROM students),'color','blue'),
				json_build_object('label','Total Dosen','value',(SELECT count(*)::text FROM lecturers),'color','purple'),
				json_build_object('label','Total Asisten Lab','value',(SELECT count(*)::text FROM lab_assistants),'color','green'),
				json_build_object('label','Total Kursus','value',(SELECT count(*)::text FROM courses),'color','orange'),
				json_build_object('label','Tugas Aktif','value',(SELECT count(*)::text FROM assignments WHERE status='pending'),'color','yellow')
			),
			'recentActivities', COALESCE((SELECT json_agg(json_build_object('action',action,'detail',detail,'time',activity_time,'icon',icon) ORDER BY sort_order) FROM admin_activities), '[]'::json),
			'topCourses', COALESCE((SELECT json_agg(json_build_object('name',name,'students',students,'completion',COALESCE(LEAST(100,attendance_present*100/NULLIF(attendance_total,0)),0),'instructor',instructor) ORDER BY students DESC) FROM courses LIMIT 4), '[]'::json)
		)`,
	"lecturer_dashboard": `
		SELECT json_build_object(
			'courses', COALESCE((SELECT json_agg(json_build_object('id',id,'code',class_code,'name',name,'class',class_code,'students',students,'schedule',CASE WHEN start_time = '' AND end_time = '' THEN day ELSE day || ', ' || start_time || ' - ' || end_time END,'room',room,'averageGrade',COALESCE((SELECT average_grade FROM lecturer_grade_courses lg WHERE lg.course_code=courses.class_code OR lg.course_name=courses.name ORDER BY lg.id LIMIT 1),0)) ORDER BY id) FROM courses LIMIT 3), '[]'::json),
			'upcomingClasses', COALESCE((SELECT json_agg(json_build_object('id',cs.id,'course',c.name,'class',c.class_code,'time',cs.session_date || ', ' || cs.session_time,'room',cs.room,'topic',cs.topic) ORDER BY cs.id) FROM course_sessions cs JOIN courses c ON c.id=cs.course_id LIMIT 2), '[]'::json),
			'recentActivities', COALESCE((SELECT json_agg(json_build_object('id',id,'type',icon,'message',detail,'time',activity_time) ORDER BY sort_order) FROM admin_activities LIMIT 3), '[]'::json)
		)`,
	"lecturer_courses":     `SELECT COALESCE(json_agg(json_build_object('id',id,'code',class_code,'name',name,'class',class_code,'semester',NULL,'academicYear',academic_year,'students',students,'schedule',CASE WHEN start_time = '' AND end_time = '' THEN day ELSE day || ', ' || start_time || ' - ' || end_time END,'room',room,'sks',credits,'description',NULLIF(description,''),'materials',(SELECT count(*) FROM course_materials WHERE course_id=courses.id),'assignments',(SELECT count(*) FROM session_assignments WHERE course_id=courses.id)) ORDER BY id), '[]'::json) FROM courses`,
	"lecturer_attendance":  attendanceQuery("lecturer"),
	"assistant_attendance": attendanceQuery("assistant"),
	"admin_attendance": `
		SELECT json_build_object(
			'courses', COALESCE((SELECT json_agg(json_build_object('id',id,'name',name,'code',class_code,'class',class_code,'instructor',instructor,'assistant',assistant,'totalSessions',sessions,'completedSessions',attendance_present,'totalStudents',students) ORDER BY id) FROM courses), '[]'::json),
			'sessions', COALESCE((SELECT json_object_agg(course_id, sessions_json) FROM (SELECT 1 AS course_id, json_agg(json_build_object('id',id,'courseId',1,'sessionNumber',session_number,'date',session_date,'time',session_time,'topic',topic,'present',present,'absent',absent,'excused',excused,'totalStudents',total_students,'status',status,'assistantStatus',assistant_status,'assistantCheckInTime',assistant_check_in_time) ORDER BY id) sessions_json FROM attendance_sessions WHERE role_scope='admin') s), '{}'::json),
			'studentAttendanceData', COALESCE((SELECT json_object_agg(session_id, records_json) FROM (SELECT session_id, json_agg(json_build_object('id',id,'nim',nim,'name',name,'status',status,'checkInTime',check_in_time) ORDER BY id) records_json FROM attendance_records WHERE session_id IN (SELECT id FROM attendance_sessions WHERE role_scope='admin') GROUP BY session_id) r), '{}'::json)
		)`,
	"lecturer_grades": lecturerGradesQuery(false),
	"admin_grades":    lecturerGradesQuery(true),
	"assistant_dashboard": `
		SELECT json_build_object(
			'practicalSessions', COALESCE((SELECT json_agg(json_build_object('id',c.id,'course',c.name,'class',c.class_code,'lab',c.room,'schedule',CASE WHEN c.start_time = '' AND c.end_time = '' THEN c.day ELSE c.day || ', ' || c.start_time || ' - ' || c.end_time END,'students',COALESCE(cardinality(cls.students),c.students,0),'attendance',c.attendance_present,'topic',NULL) ORDER BY c.id) FROM courses c LEFT JOIN classes cls ON cls.code = c.class_code WHERE EXISTS (SELECT 1 FROM lab_assistants la WHERE la.name = c.assistant)), '[]'::json),
			'pendingTasks', COALESCE((SELECT json_agg(json_build_object('id',id,'task',task,'submitted',submitted,'total',total,'deadline',deadline) ORDER BY id) FROM assistant_tasks), '[]'::json),
			'upcomingPracticals', COALESCE((SELECT json_agg(json_build_object('id',c.id,'course',c.name,'class',c.class_code,'time',CASE WHEN c.start_time = '' AND c.end_time = '' THEN c.day ELSE c.day || ', ' || c.start_time || ' - ' || c.end_time END,'lab',c.room,'topic',NULL) ORDER BY c.id) FROM courses c WHERE EXISTS (SELECT 1 FROM lab_assistants la WHERE la.name = c.assistant)), '[]'::json),
			'recentActivities', COALESCE((SELECT json_agg(json_build_object('id',id,'message',detail,'time',activity_time,'type',icon) ORDER BY sort_order) FROM admin_activities LIMIT 2), '[]'::json)
		)`,
	"assistant_practicals": `
		SELECT json_build_object(
			'courses', COALESCE((SELECT json_agg(json_build_object('id',c.id,'courseCode',c.class_code,'name',c.name,'lecturer',c.instructor,'class',c.class_code,'lab',c.room,'schedule',json_build_object('day',c.day,'startTime',c.start_time,'endTime',c.end_time),'semester',NULL,'academicYear',c.academic_year,'students',COALESCE(cardinality(cls.students),c.students,0),'attendance',json_build_object('present',c.attendance_present,'total',c.attendance_total,'percentage',COALESCE(c.attendance_present*100/NULLIF(c.attendance_total,0),0)),'color',c.color) ORDER BY c.id) FROM courses c LEFT JOIN classes cls ON cls.code = c.class_code WHERE EXISTS (SELECT 1 FROM lab_assistants la WHERE la.name = c.assistant)), '[]'::json),
			'sessions', COALESCE((SELECT json_agg(json_build_object('id',cs.id,'courseId',cs.course_id,'sessionNumber',cs.session_number,'title',cs.topic,'date',cs.session_date,'time',cs.session_time,'description',cs.description,'room',cs.room,'reportStatus',(SELECT ar.status FROM assistant_reports ar WHERE ar.course_code=c.class_code AND ar.week=cs.session_number ORDER BY ar.id DESC LIMIT 1),'reportId',(SELECT ar.id FROM assistant_reports ar WHERE ar.course_code=c.class_code AND ar.week=cs.session_number ORDER BY ar.id DESC LIMIT 1)) ORDER BY cs.course_id, cs.session_number) FROM course_sessions cs JOIN courses c ON c.id=cs.course_id WHERE EXISTS (SELECT 1 FROM lab_assistants la WHERE la.name = c.assistant)), '[]'::json),
			'materials', COALESCE((SELECT json_agg(json_build_object('id',m.id,'courseId',m.course_id,'sessionId',m.session_id,'name',m.title,'type',m.material_type,'size',m.size,'uploadDate',m.upload_date) ORDER BY m.id) FROM course_materials m WHERE m.session_id IS NOT NULL), '[]'::json),
			'assignments', COALESCE((SELECT json_agg(json_build_object('id',sa.id,'courseId',sa.course_id,'sessionId',sa.session_id,'title',sa.title,'description',sa.description,'deadline',sa.due_date,'submittedCount',sa.submitted_count,'totalStudents',sa.total_students,'sessionNumber',cs.session_number) ORDER BY sa.id) FROM session_assignments sa LEFT JOIN course_sessions cs ON cs.id=sa.session_id), '[]'::json),
			'submissions', COALESCE((SELECT json_agg(json_build_object('id',id,'nim',nim,'name',name,'submittedAt',submitted_at,'fileName',file_name,'fileSize',file_size,'score',score,'feedback',CASE WHEN score IS NULL THEN '' ELSE 'Bagus!' END) ORDER BY id) FROM assistant_reports), '[]'::json),
			'attendanceRecordsBySession', COALESCE((
				SELECT json_object_agg(course_session_id, records)
				FROM (
					SELECT cs.id course_session_id, json_agg(json_build_object('id',ar.id,'nim',ar.nim,'name',ar.name,'status',ar.status,'time',COALESCE(NULLIF(ar.attendance_time,''),'-')) ORDER BY ar.id) records
					FROM course_sessions cs
					JOIN courses c ON c.id=cs.course_id
					JOIN attendance_sessions ats ON ats.role_scope='assistant' AND ats.course_code=c.class_code AND ats.session_number=cs.session_number
					JOIN attendance_records ar ON ar.session_id=ats.id
					GROUP BY cs.id
				) s
			), '{}'::json),
			'attendanceRecords', COALESCE((SELECT json_agg(json_build_object('id',id,'nim',nim,'name',name,'status',status,'time',COALESCE(NULLIF(attendance_time,''),'-')) ORDER BY id) FROM attendance_records WHERE session_id=(SELECT id FROM attendance_sessions WHERE role_scope='assistant' ORDER BY id LIMIT 1)), '[]'::json)
		)`,
	"assistant_reports":      `SELECT json_build_object('reports', COALESCE((SELECT json_agg(json_build_object('id',ar.id,'nim',ar.nim,'name',ar.name,'courseCode',ar.course_code,'courseName',ar.course_name,'class',ar.class_name,'week',ar.week,'topic',ar.topic,'submittedAt',ar.submitted_at,'status',ar.status,'score',ar.score,'fileName',ar.file_name,'fileSize',ar.file_size,'lecturer',c.instructor,'assistant',c.assistant) ORDER BY ar.id) FROM assistant_reports ar LEFT JOIN courses c ON c.class_code=ar.course_code), '[]'::json), 'reportSummary', COALESCE((SELECT json_agg(json_build_object('courseCode',course_code,'courseName',course_name,'class',class_name,'totalReports',total_reports,'reviewed',reviewed,'pending',pending,'approved',approved,'needsRevision',needs_revision) ORDER BY course_code) FROM (SELECT course_code,course_name,class_name,count(*) total_reports,count(*) FILTER (WHERE status IN ('Disetujui','Ditolak')) reviewed,count(*) FILTER (WHERE status='Menunggu Review') pending,count(*) FILTER (WHERE status='Disetujui') approved,count(*) FILTER (WHERE status='Ditolak') needs_revision FROM assistant_reports GROUP BY course_code,course_name,class_name) s), '[]'::json))`,
	"student_course_detail":  `SELECT json_build_object('materials', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'week',week,'status',status) ORDER BY id) FROM course_materials WHERE course_id=1 AND week IS NOT NULL), '[]'::json), 'courseSessions', COALESCE((SELECT json_agg(json_build_object('id',id,'sessionName',title,'topic',topic,'date',session_date,'time',session_time,'type',session_type,'conferenceLink',conference_link) ORDER BY id) FROM course_sessions WHERE course_id=1), '[]'::json), 'assignments', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'dueDate',due_date,'status',status,'score',score) ORDER BY id) FROM session_assignments WHERE course_id=1 AND session_id IS NULL), '[]'::json))`,
	"student_session_detail": `SELECT json_build_object('materials', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'type',material_type,'size',size) ORDER BY id) FROM course_materials WHERE session_id=1), '[]'::json), 'assignments', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'deadline',due_date,'status',status) ORDER BY id) FROM session_assignments WHERE session_id=1), '[]'::json))`,
	"student_materials":      `SELECT json_build_object('courses', (SELECT json_agg(x ORDER BY ord) FROM (SELECT 0 ord, json_build_object('id','all','name','Semua Kursus') x UNION ALL SELECT id, json_build_object('id',id,'name',name,'color',color) FROM courses) q), 'materials', COALESCE((SELECT json_agg(json_build_object('id',m.id,'title',m.title,'type',lower(m.material_type),'courseId',m.course_id,'courseName',c.name,'size',m.size,'duration',NULLIF(m.duration,''),'downloads',m.downloads,'uploadDate',m.upload_date,'courseColor',c.color) ORDER BY m.id) FROM course_materials m JOIN courses c ON c.id=m.course_id WHERE m.session_id IS NULL), '[]'::json))`,
	"admin_reports": `SELECT json_build_object(
		'stats', json_build_array(
			json_build_object('label','Total Mahasiswa Aktif','value',(SELECT count(*)::text FROM students WHERE status='Aktif')),
			json_build_object('label','Total Mahasiswa','value',(SELECT count(*)::text FROM students)),
			json_build_object('label','Total Laporan','value',(SELECT count(*)::text FROM assistant_reports)),
			json_build_object('label','Laporan Approved','value',(SELECT count(*)::text FROM assistant_reports WHERE status='approved'))
		),
		'coursePerformance', COALESCE((
			SELECT json_agg(json_build_object(
				'course', course_name,
				'students', total_reports,
				'avgScore', COALESCE((SELECT round(avg(score)) FROM assistant_reports ar WHERE ar.course_code=s.course_code AND score IS NOT NULL),0),
				'completion', COALESCE(reviewed*100/NULLIF(total_reports,0),0),
				'passing', COALESCE(approved*100/NULLIF(total_reports,0),0)
			) ORDER BY id)
			FROM assistant_report_summary s
		), '[]'::json),
		'topStudents', COALESCE((SELECT json_agg(json_build_object('name',name,'program',program,'gpa',NULL,'completedCourses',cardinality(courses)) ORDER BY id) FROM students LIMIT 3), '[]'::json),
		'monthlyRegistrations', COALESCE((SELECT json_agg(json_build_object('month',to_char(month,'Mon YYYY'),'students',students) ORDER BY month) FROM (SELECT date_trunc('month', join_date) month, count(*) students FROM students GROUP BY 1) m), '[]'::json)
	)`,
	"admin_assignments":         `SELECT COALESCE(json_agg(json_build_object('id',id,'title',title,'course',course,'instructor',COALESCE(NULLIF(instructor,''),assistant),'dueDate',due_date,'totalStudents',total_students,'submitted',submitted,'graded',graded,'pending',pending,'status',CASE WHEN status='graded' THEN 'Selesai' ELSE 'Aktif' END,'type',assignment_type) ORDER BY id), '[]'::json) FROM assignments`,
	"admin_class_management":    `SELECT COALESCE(json_agg(json_build_object('id',c.id,'name',c.name,'code',c.code,'academicYear',c.academic_year,'semester','Genap','capacity',c.capacity,'students',COALESCE((SELECT json_agg(json_build_object('id',s.id,'name',s.name,'nim',s.student_id,'email',s.email) ORDER BY s.id) FROM students s WHERE s.id=ANY(c.students)), '[]'::json),'courses',COALESCE((SELECT json_agg(name ORDER BY id) FROM courses WHERE class_code=c.code), '[]'::json)) ORDER BY c.id), '[]'::json) FROM classes c`,
	"admin_class_students":      `SELECT COALESCE(json_agg(json_build_object('id',id,'nim',student_id,'name',name,'email',email,'status',status) ORDER BY id), '[]'::json) FROM students`,
	"admin_course_form_options": `SELECT json_build_object('dosenList', COALESCE((SELECT json_agg(name ORDER BY id) FROM lecturers), '[]'::json), 'asistenList', COALESCE((SELECT json_agg(name ORDER BY id) FROM lab_assistants), '[]'::json), 'kelasList', COALESCE((SELECT json_agg(json_build_object('code',code,'name',name) ORDER BY id) FROM classes), '[]'::json))`,
}

func attendanceQuery(scope string) string {
	roomExpr := "'room', room"
	if scope == "assistant" {
		roomExpr = "'lab', lab"
	}
	filter := `role_scope='` + scope + `'`
	if scope == "assistant" {
		filter += ` AND EXISTS (SELECT 1 FROM courses c WHERE c.class_code = attendance_sessions.course_code AND EXISTS (SELECT 1 FROM lab_assistants la WHERE la.name = c.assistant))`
	}
	return `SELECT json_build_object('attendanceSessions', COALESCE((SELECT json_agg(json_build_object('id',id,'courseCode',course_code,'courseName',course_name,'class',class_name,'date',session_date,'time',session_time,` + roomExpr + `,'topic',topic,'totalStudents',total_students,'present',present,'absent',absent,'sick',sick,'permit',permit) ORDER BY id) FROM attendance_sessions WHERE ` + filter + `), '[]'::json), 'studentAttendanceBySession', COALESCE((SELECT json_object_agg(session_id, records_json) FROM (SELECT session_id, json_agg(json_build_object('id',id,'nim',nim,'name',name,'status',status,'time',NULLIF(attendance_time,'')) ORDER BY id) records_json FROM attendance_records WHERE session_id IN (SELECT id FROM attendance_sessions WHERE ` + filter + `) GROUP BY session_id) r), '{}'::json), 'studentAttendance', COALESCE((SELECT json_agg(json_build_object('id',id,'nim',nim,'name',name,'status',status,'time',NULLIF(attendance_time,'')) ORDER BY id) FROM attendance_records WHERE session_id=(SELECT id FROM attendance_sessions WHERE ` + filter + ` ORDER BY id LIMIT 1)), '[]'::json))`
}

func lecturerGradesQuery(admin bool) string {
	courseClassKey := "'class'"
	finalExamKey := "'ujianAkhir'"
	finalGradeKey := "'nilaiAkhir'"
	letterKey := "'grade'"
	statusPart := ""
	if admin {
		courseClassKey = "'className'"
		finalExamKey = "'finalExam'"
		finalGradeKey = "'finalGrade'"
		letterKey = "'letterGrade'"
		statusPart = `,'status', status`
	}
	studentKey := "studentGrades"
	if admin {
		studentKey = "studentGradesData"
		return `SELECT json_build_object('courseGrades', COALESCE((SELECT json_agg(json_build_object('id',id,'courseName',course_name,'courseCode',course_code,` + courseClassKey + `,class_name,'semester',semester,'academicYear',academic_year,'totalStudents',total_students,'averageGrade',average_grade,'highestGrade',highest_grade,'lowestGrade',lowest_grade,'passRate',pass_rate) ORDER BY id) FROM lecturer_grade_courses), '[]'::json), '` + studentKey + `', COALESCE((SELECT json_object_agg(grade_course_id, records) FROM (SELECT grade_course_id, json_agg(json_build_object('id',id,'nim',nim,'name',name,'tugas1',tugas1,'tugas2',tugas2,'tugas3',tugas3,` + finalExamKey + `,ujian_akhir,` + finalGradeKey + `,nilai_akhir,` + letterKey + `,grade` + statusPart + `) ORDER BY id) records FROM lecturer_student_grades GROUP BY grade_course_id) s), '{}'::json))`
	}
	return `SELECT json_build_object('courseGrades', COALESCE((SELECT json_agg(json_build_object('id',id,'courseCode',course_code,'courseName',course_name,` + courseClassKey + `,class_name,'semester',semester,'academicYear',academic_year,'totalStudents',total_students,'averageGrade',average_grade,'highestGrade',highest_grade,'lowestGrade',lowest_grade) ORDER BY id) FROM lecturer_grade_courses), '[]'::json), '` + studentKey + `', COALESCE((SELECT json_agg(json_build_object('id',id,'nim',nim,'name',name,'tugas1',tugas1,'tugas2',tugas2,'tugas3',tugas3,` + finalExamKey + `,ujian_akhir,` + finalGradeKey + `,nilai_akhir,` + letterKey + `,grade) ORDER BY id) FROM lecturer_student_grades WHERE grade_course_id=1), '[]'::json))`
}
