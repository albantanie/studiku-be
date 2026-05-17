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

	var payload []byte
	if err := r.db.QueryRow(query).Scan(&payload); err != nil {
		return nil, err
	}
	return json.RawMessage(payload), nil
}

var pageDataQueries = map[string]string{
	"student_grade_semesters": `SELECT COALESCE(json_agg(json_build_object('id', value, 'name', label) ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='student_grade_semesters'`,
	"admin_student_programs": `SELECT COALESCE(json_agg(value ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='student_programs'`,
	"admin_labs": `SELECT COALESCE(json_agg(value ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='labs'`,
	"admin_user_tabs": `SELECT COALESCE(json_agg(json_build_object('id', value, 'label', label, 'icon', icon) ORDER BY sort_order), '[]'::json) FROM ui_options WHERE option_group='user_tabs'`,
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
				json_build_object('label','Total Mahasiswa','value',to_char((SELECT count(*) FROM students),'FM9,999'),'color','blue','change','+12%'),
				json_build_object('label','Total Kursus','value',(SELECT count(*)::text FROM courses),'color','purple','change','+5%'),
				json_build_object('label','Tugas Aktif','value',(SELECT count(*)::text FROM assignments),'color','green','change','+8%'),
				json_build_object('label','Tingkat Kelulusan','value','87%','color','orange','change','+3%')
			),
			'recentActivities', COALESCE((SELECT json_agg(json_build_object('action',action,'detail',detail,'time',activity_time,'icon',icon) ORDER BY sort_order) FROM admin_activities), '[]'::json),
			'topCourses', COALESCE((SELECT json_agg(json_build_object('name',name,'students',students,'completion',COALESCE(LEAST(100,attendance_present*100/NULLIF(attendance_total,0)),0),'instructor',instructor) ORDER BY students DESC) FROM courses LIMIT 4), '[]'::json)
		)`,
	"lecturer_dashboard": `
		SELECT json_build_object(
			'courses', COALESCE((SELECT json_agg(json_build_object('id',id,'code',class_code,'name',name,'class',class_code,'students',students,'schedule',CASE WHEN start_time = '' AND end_time = '' THEN day ELSE day || ', ' || start_time || ' - ' || end_time END,'room',room,'averageGrade',78.5) ORDER BY id) FROM courses LIMIT 3), '[]'::json),
			'upcomingClasses', COALESCE((SELECT json_agg(json_build_object('id',cs.id,'course',c.name,'class',c.class_code,'time',cs.session_date || ', ' || cs.session_time,'room',cs.room,'topic',cs.topic) ORDER BY cs.id) FROM course_sessions cs JOIN courses c ON c.id=cs.course_id LIMIT 2), '[]'::json),
			'recentActivities', COALESCE((SELECT json_agg(json_build_object('id',id,'type',icon,'message',detail,'time',activity_time) ORDER BY sort_order) FROM admin_activities LIMIT 3), '[]'::json)
		)`,
	"lecturer_courses": `SELECT COALESCE(json_agg(json_build_object('id',id,'code',class_code,'name',name,'class',class_code,'semester','Ganjil','academicYear',academic_year,'students',students,'schedule',CASE WHEN start_time = '' AND end_time = '' THEN day ELSE day || ', ' || start_time || ' - ' || end_time END,'room',room,'sks',credits,'description',COALESCE(NULLIF(description,''),'Mata kuliah ' || lower(name)),'materials',(SELECT count(*) FROM course_materials WHERE course_id=courses.id),'assignments',(SELECT count(*) FROM session_assignments WHERE course_id=courses.id)) ORDER BY id), '[]'::json) FROM courses`,
	"lecturer_attendance": attendanceQuery("lecturer"),
	"assistant_attendance": attendanceQuery("assistant"),
	"admin_attendance": `
		SELECT json_build_object(
			'courses', COALESCE((SELECT json_agg(json_build_object('id',id,'name',name,'code',class_code,'class',class_code,'instructor',instructor,'assistant',assistant,'totalSessions',sessions,'completedSessions',attendance_present,'totalStudents',students) ORDER BY id) FROM courses), '[]'::json),
			'sessions', COALESCE((SELECT json_object_agg(course_id, sessions_json) FROM (SELECT 1 AS course_id, json_agg(json_build_object('id',id,'courseId',1,'sessionNumber',session_number,'date',session_date,'time',session_time,'topic',topic,'present',present,'absent',absent,'excused',excused,'totalStudents',total_students,'status',status,'assistantStatus',assistant_status,'assistantCheckInTime',assistant_check_in_time) ORDER BY id) sessions_json FROM attendance_sessions WHERE role_scope='admin') s), '{}'::json),
			'studentAttendanceData', COALESCE((SELECT json_object_agg(session_id, records_json) FROM (SELECT session_id, json_agg(json_build_object('id',id,'nim',nim,'name',name,'status',status,'checkInTime',check_in_time) ORDER BY id) records_json FROM attendance_records WHERE session_id IN (SELECT id FROM attendance_sessions WHERE role_scope='admin') GROUP BY session_id) r), '{}'::json)
		)`,
	"lecturer_grades": lecturerGradesQuery(false),
	"admin_grades": lecturerGradesQuery(true),
	"assistant_dashboard": `
		SELECT json_build_object(
			'practicalSessions', COALESCE((SELECT json_agg(json_build_object('id',id,'course',name,'class',class_name,'lab',lab,'schedule',CASE WHEN start_time = '' AND end_time = '' THEN day ELSE day || ', ' || start_time || ' - ' || end_time END,'students',students,'attendance',attendance_present,'topic','Praktikum ' || name) ORDER BY id) FROM assistant_practical_courses), '[]'::json),
			'pendingTasks', COALESCE((SELECT json_agg(json_build_object('id',id,'task',task,'submitted',submitted,'total',total,'deadline',deadline) ORDER BY id) FROM assistant_tasks), '[]'::json),
			'upcomingPracticals', COALESCE((SELECT json_agg(json_build_object('id',id,'course',name,'class',class_name,'time',CASE WHEN start_time = '' AND end_time = '' THEN day ELSE day || ', ' || start_time || ' - ' || end_time END,'lab',lab,'topic','Praktikum ' || name) ORDER BY id) FROM assistant_practical_courses LIMIT 2), '[]'::json),
			'recentActivities', COALESCE((SELECT json_agg(json_build_object('id',id,'message',detail,'time',activity_time,'type',icon) ORDER BY sort_order) FROM admin_activities LIMIT 2), '[]'::json)
		)`,
	"assistant_practicals": `
		SELECT json_build_object(
			'courses', COALESCE((SELECT json_agg(json_build_object('id',id,'courseCode',course_code,'name',name,'lecturer',lecturer,'class',class_name,'lab',lab,'schedule',json_build_object('day',day,'startTime',start_time,'endTime',end_time),'semester',semester,'academicYear',academic_year,'students',students,'attendance',json_build_object('present',attendance_present,'total',attendance_total,'percentage',COALESCE(attendance_present*100/NULLIF(attendance_total,0),0)),'color',color) ORDER BY id) FROM assistant_practical_courses), '[]'::json),
			'sessions', COALESCE((SELECT json_agg(json_build_object('id',id,'sessionNumber',session_number,'title',topic,'date',session_date,'time',session_time,'description',description,'room',room) ORDER BY id) FROM course_sessions), '[]'::json),
			'materials', COALESCE((SELECT json_agg(json_build_object('id',id,'name',title,'type',material_type,'size',size,'uploadDate',upload_date) ORDER BY id) FROM course_materials WHERE session_id IS NOT NULL LIMIT 1), '[]'::json),
			'assignments', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'description',description,'deadline',due_date,'submittedCount',submitted_count,'totalStudents',total_students,'sessionNumber',1) ORDER BY id) FROM session_assignments WHERE id=6), '[]'::json),
			'submissions', COALESCE((SELECT json_agg(json_build_object('id',id,'nim',nim,'name',name,'submittedAt',submitted_at,'fileName',file_name,'fileSize',file_size,'score',score,'feedback',CASE WHEN score IS NULL THEN '' ELSE 'Bagus!' END) ORDER BY id) FROM assistant_reports LIMIT 1), '[]'::json),
			'attendanceRecords', COALESCE((SELECT json_agg(json_build_object('id',id,'nim',nim,'name',name,'status',status,'time',COALESCE(NULLIF(attendance_time,''),'-')) ORDER BY id) FROM attendance_records WHERE session_id=1 LIMIT 2), '[]'::json)
		)`,
	"assistant_reports": `SELECT json_build_object('reports', COALESCE((SELECT json_agg(json_build_object('id',id,'nim',nim,'name',name,'courseCode',course_code,'courseName',course_name,'class',class_name,'week',week,'topic',topic,'submittedAt',submitted_at,'status',status,'score',score,'fileName',file_name,'fileSize',file_size) ORDER BY id) FROM assistant_reports), '[]'::json), 'reportSummary', COALESCE((SELECT json_agg(json_build_object('courseCode',course_code,'courseName',course_name,'class',class_name,'totalReports',total_reports,'reviewed',reviewed,'pending',pending,'approved',approved,'needsRevision',needs_revision) ORDER BY id) FROM assistant_report_summary), '[]'::json))`,
	"student_course_detail": `SELECT json_build_object('materials', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'week',week,'status',status) ORDER BY id) FROM course_materials WHERE course_id=1 AND week IS NOT NULL), '[]'::json), 'courseSessions', COALESCE((SELECT json_agg(json_build_object('id',id,'sessionName',title,'topic',topic,'date',session_date,'time',session_time,'type',session_type,'conferenceLink',conference_link) ORDER BY id) FROM course_sessions WHERE course_id=1), '[]'::json), 'assignments', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'dueDate',due_date,'status',status,'score',score) ORDER BY id) FROM session_assignments WHERE course_id=1 AND session_id IS NULL), '[]'::json))`,
	"student_session_detail": `SELECT json_build_object('materials', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'type',material_type,'size',size) ORDER BY id) FROM course_materials WHERE session_id=1), '[]'::json), 'assignments', COALESCE((SELECT json_agg(json_build_object('id',id,'title',title,'deadline',due_date,'status',status) ORDER BY id) FROM session_assignments WHERE session_id=1), '[]'::json))`,
	"student_materials": `SELECT json_build_object('courses', (SELECT json_agg(x ORDER BY ord) FROM (SELECT 0 ord, json_build_object('id','all','name','Semua Kursus') x UNION ALL SELECT id, json_build_object('id',id,'name',name,'color',color) FROM courses) q), 'materials', COALESCE((SELECT json_agg(json_build_object('id',m.id,'title',m.title,'type',lower(m.material_type),'courseId',m.course_id,'courseName',c.name,'size',m.size,'duration',NULLIF(m.duration,''),'downloads',m.downloads,'uploadDate',m.upload_date,'courseColor',c.color) ORDER BY m.id) FROM course_materials m JOIN courses c ON c.id=m.course_id WHERE m.session_id IS NULL), '[]'::json))`,
	"admin_reports": `SELECT json_build_object('stats', json_build_array(json_build_object('label','Total Mahasiswa Aktif','value',(SELECT count(*)::text FROM students WHERE status='Aktif'),'change','+12%','trend','up'), json_build_object('label','Rata-rata IPK','value','3.45','change','+0.15','trend','up'), json_build_object('label','Tingkat Kelulusan','value','87%','change','+3%','trend','up'), json_build_object('label','Mahasiswa Baru','value',(SELECT count(*)::text FROM students),'change','+8%','trend','up')), 'coursePerformance', COALESCE((SELECT json_agg(json_build_object('course',name,'students',students,'avgScore',82,'completion',COALESCE(attendance_present*100/NULLIF(attendance_total,0),0),'passing',89) ORDER BY id) FROM courses LIMIT 3), '[]'::json), 'topStudents', COALESCE((SELECT json_agg(json_build_object('name',name,'program',program,'gpa',3.75,'completedCourses',array_length(courses,1)) ORDER BY id) FROM students LIMIT 3), '[]'::json), 'monthlyRegistrations', json_build_array(json_build_object('month','Sep 2024','students',198),json_build_object('month','Oct 2024','students',156),json_build_object('month','Nov 2024','students',142),json_build_object('month','Dec 2024','students',178),json_build_object('month','Jan 2025','students',245)))`,
	"admin_assignments": `SELECT COALESCE(json_agg(json_build_object('id',id,'title',title,'course',course,'instructor',COALESCE(NULLIF(instructor,''),assistant),'dueDate',due_date,'totalStudents',total_students,'submitted',submitted,'graded',graded,'pending',pending,'status',CASE WHEN status='graded' THEN 'Selesai' ELSE 'Aktif' END,'type',assignment_type) ORDER BY id), '[]'::json) FROM assignments`,
	"admin_class_management": `SELECT COALESCE(json_agg(json_build_object('id',c.id,'name',c.name,'code',c.code,'academicYear',c.academic_year,'semester','Genap','capacity',c.capacity,'students',COALESCE((SELECT json_agg(json_build_object('id',s.id,'name',s.name,'nim',s.student_id,'email',s.email) ORDER BY s.id) FROM students s WHERE s.id=ANY(c.students)), '[]'::json),'courses',COALESCE((SELECT json_agg(name ORDER BY id) FROM courses WHERE class_code=c.code), '[]'::json)) ORDER BY c.id), '[]'::json) FROM classes c`,
	"admin_class_students": `SELECT COALESCE(json_agg(json_build_object('id',id,'nim',student_id,'name',name,'email',email,'status',status) ORDER BY id), '[]'::json) FROM students`,
	"admin_course_form_options": `SELECT json_build_object('dosenList', COALESCE((SELECT json_agg(name ORDER BY id) FROM lecturers), '[]'::json), 'asistenList', COALESCE((SELECT json_agg(name ORDER BY id) FROM lab_assistants), '[]'::json), 'kelasList', COALESCE((SELECT json_agg(json_build_object('code',code,'name',name) ORDER BY id) FROM classes), '[]'::json))`,
}

func attendanceQuery(scope string) string {
	roomExpr := "'room', room"
	if scope == "assistant" {
		roomExpr = "'lab', lab"
	}
	return `SELECT json_build_object('attendanceSessions', COALESCE((SELECT json_agg(json_build_object('id',id,'courseCode',course_code,'courseName',course_name,'class',class_name,'date',session_date,'time',session_time,` + roomExpr + `,'topic',topic,'totalStudents',total_students,'present',present,'absent',absent,'sick',sick,'permit',permit) ORDER BY id) FROM attendance_sessions WHERE role_scope='` + scope + `'), '[]'::json), 'studentAttendance', COALESCE((SELECT json_agg(json_build_object('id',id,'nim',nim,'name',name,'status',status,'time',NULLIF(attendance_time,'')) ORDER BY id) FROM attendance_records WHERE session_id=(SELECT id FROM attendance_sessions WHERE role_scope='` + scope + `' ORDER BY id LIMIT 1)), '[]'::json))`
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
