CREATE TABLE IF NOT EXISTS ui_options (
  id SERIAL PRIMARY KEY,
  option_group VARCHAR(100) NOT NULL,
  label VARCHAR(255) NOT NULL,
  value VARCHAR(255) NOT NULL,
  icon VARCHAR(50) NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0,
  UNIQUE(option_group, value)
);

CREATE TABLE IF NOT EXISTS import_lecturer_preview_courses (
  id SERIAL PRIMARY KEY,
  lecturer_email VARCHAR(255) NOT NULL,
  course_name VARCHAR(255) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS course_sessions (
  id SERIAL PRIMARY KEY,
  course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  session_number INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  topic VARCHAR(255) NOT NULL,
  session_date VARCHAR(50) NOT NULL,
  session_time VARCHAR(50) NOT NULL,
  session_type VARCHAR(20) NOT NULL DEFAULT 'offline',
  conference_link TEXT,
  room VARCHAR(100) NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  UNIQUE(course_id, session_number)
);

CREATE TABLE IF NOT EXISTS course_materials (
  id SERIAL PRIMARY KEY,
  course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  session_id INT REFERENCES course_sessions(id) ON DELETE SET NULL,
  title VARCHAR(255) NOT NULL,
  material_type VARCHAR(50) NOT NULL,
  size VARCHAR(50) NOT NULL DEFAULT '',
  duration VARCHAR(50) NOT NULL DEFAULT '',
  downloads INT NOT NULL DEFAULT 0,
  upload_date VARCHAR(50) NOT NULL DEFAULT '',
  week INT,
  status VARCHAR(50) NOT NULL DEFAULT 'available'
);

CREATE TABLE IF NOT EXISTS session_assignments (
  id SERIAL PRIMARY KEY,
  course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  session_id INT REFERENCES course_sessions(id) ON DELETE SET NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  due_date VARCHAR(80) NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'pending',
  score INT,
  submitted_count INT NOT NULL DEFAULT 0,
  total_students INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS attendance_sessions (
  id SERIAL PRIMARY KEY,
  role_scope VARCHAR(20) NOT NULL CHECK (role_scope IN ('admin','lecturer','assistant')),
  course_code VARCHAR(50) NOT NULL,
  course_name VARCHAR(255) NOT NULL,
  class_name VARCHAR(50) NOT NULL,
  session_number INT NOT NULL DEFAULT 1,
  session_date VARCHAR(50) NOT NULL,
  session_time VARCHAR(50) NOT NULL,
  room VARCHAR(100) NOT NULL DEFAULT '',
  lab VARCHAR(100) NOT NULL DEFAULT '',
  topic VARCHAR(255) NOT NULL,
  total_students INT NOT NULL DEFAULT 0,
  present INT NOT NULL DEFAULT 0,
  absent INT NOT NULL DEFAULT 0,
  sick INT NOT NULL DEFAULT 0,
  permit INT NOT NULL DEFAULT 0,
  excused INT NOT NULL DEFAULT 0,
  status VARCHAR(50) NOT NULL DEFAULT 'Selesai',
  assistant_status VARCHAR(50) NOT NULL DEFAULT '',
  assistant_check_in_time VARCHAR(20) NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS attendance_records (
  id SERIAL PRIMARY KEY,
  session_id INT NOT NULL REFERENCES attendance_sessions(id) ON DELETE CASCADE,
  nim VARCHAR(50) NOT NULL,
  name VARCHAR(255) NOT NULL,
  status VARCHAR(50) NOT NULL,
  attendance_time VARCHAR(20) NOT NULL DEFAULT '',
  check_in_time VARCHAR(20) NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS lecturer_grade_courses (
  id SERIAL PRIMARY KEY,
  course_code VARCHAR(50) NOT NULL,
  course_name VARCHAR(255) NOT NULL,
  class_name VARCHAR(50) NOT NULL,
  semester VARCHAR(50) NOT NULL,
  academic_year VARCHAR(50) NOT NULL,
  total_students INT NOT NULL,
  average_grade NUMERIC(5,2) NOT NULL,
  highest_grade INT NOT NULL,
  lowest_grade INT NOT NULL,
  pass_rate NUMERIC(5,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS lecturer_student_grades (
  id SERIAL PRIMARY KEY,
  grade_course_id INT NOT NULL REFERENCES lecturer_grade_courses(id) ON DELETE CASCADE,
  nim VARCHAR(50) NOT NULL,
  name VARCHAR(255) NOT NULL,
  tugas1 INT NOT NULL DEFAULT 0,
  tugas2 INT NOT NULL DEFAULT 0,
  tugas3 INT NOT NULL DEFAULT 0,
  ujian_akhir INT NOT NULL DEFAULT 0,
  nilai_akhir NUMERIC(5,2) NOT NULL DEFAULT 0,
  grade VARCHAR(5) NOT NULL DEFAULT 'E',
  status VARCHAR(50) NOT NULL DEFAULT 'Lulus'
);

CREATE TABLE IF NOT EXISTS assistant_practical_courses (
  id SERIAL PRIMARY KEY,
  course_code VARCHAR(50) NOT NULL,
  name VARCHAR(255) NOT NULL,
  lecturer VARCHAR(255) NOT NULL,
  class_name VARCHAR(50) NOT NULL,
  lab VARCHAR(100) NOT NULL,
  day VARCHAR(50) NOT NULL,
  start_time VARCHAR(20) NOT NULL,
  end_time VARCHAR(20) NOT NULL,
  semester VARCHAR(50) NOT NULL,
  academic_year VARCHAR(50) NOT NULL,
  students INT NOT NULL DEFAULT 0,
  attendance_present INT NOT NULL DEFAULT 0,
  attendance_total INT NOT NULL DEFAULT 0,
  color VARCHAR(50) NOT NULL DEFAULT 'bg-blue-500'
);

CREATE TABLE IF NOT EXISTS assistant_tasks (
  id SERIAL PRIMARY KEY,
  task VARCHAR(255) NOT NULL,
  submitted INT NOT NULL DEFAULT 0,
  total INT NOT NULL DEFAULT 0,
  deadline VARCHAR(80) NOT NULL
);

CREATE TABLE IF NOT EXISTS assistant_reports (
  id SERIAL PRIMARY KEY,
  nim VARCHAR(50) NOT NULL,
  name VARCHAR(255) NOT NULL,
  course_code VARCHAR(50) NOT NULL,
  course_name VARCHAR(255) NOT NULL,
  class_name VARCHAR(50) NOT NULL,
  week INT NOT NULL,
  topic VARCHAR(255) NOT NULL,
  submitted_at VARCHAR(80) NOT NULL,
  status VARCHAR(80) NOT NULL,
  score INT,
  file_name VARCHAR(255) NOT NULL,
  file_size VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS assistant_report_summary (
  id SERIAL PRIMARY KEY,
  course_code VARCHAR(50) NOT NULL,
  course_name VARCHAR(255) NOT NULL,
  class_name VARCHAR(50) NOT NULL,
  total_reports INT NOT NULL DEFAULT 0,
  reviewed INT NOT NULL DEFAULT 0,
  pending INT NOT NULL DEFAULT 0,
  approved INT NOT NULL DEFAULT 0,
  needs_revision INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS admin_activities (
  id SERIAL PRIMARY KEY,
  action VARCHAR(255) NOT NULL,
  detail VARCHAR(255) NOT NULL,
  activity_time VARCHAR(80) NOT NULL,
  icon VARCHAR(50) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);

INSERT INTO ui_options (option_group,label,value,icon,sort_order) VALUES
('student_grade_semesters','2025/2026','2025/2026','',1),('student_grade_semesters','2024/2025','2024/2025','',2),('student_grade_semesters','2023/2024','2023/2024','',3),
('student_programs','Teknik Informatika','Teknik Informatika','',1),('student_programs','Sistem Informasi','Sistem Informasi','',2),('student_programs','Desain Komunikasi Visual','Desain Komunikasi Visual','',3),('student_programs','Teknik Komputer','Teknik Komputer','',4),('student_programs','Manajemen Informatika','Manajemen Informatika','',5),
('labs','Semua Lab','Semua Lab','',1),('labs','Laboratorium Pemrograman','Laboratorium Pemrograman','',2),('labs','Laboratorium Database','Laboratorium Database','',3),('labs','Laboratorium Jaringan','Laboratorium Jaringan','',4),('labs','Laboratorium AI & Machine Learning','Laboratorium AI & Machine Learning','',5),('labs','Laboratorium Desain','Laboratorium Desain','',6),
('user_tabs','Mahasiswa','students','graduation',1),('user_tabs','Dosen','lecturers','users',2),('user_tabs','Asisten Laboratorium','assistants','userCog',3),
('import_file_types','text/csv','text/csv','',1),('import_file_types','application/vnd.ms-excel','application/vnd.ms-excel','',2),('import_file_types','application/vnd.openxmlformats-officedocument.spreadsheetml.sheet','application/vnd.openxmlformats-officedocument.spreadsheetml.sheet','',3)
ON CONFLICT (option_group,value) DO UPDATE SET label=EXCLUDED.label, icon=EXCLUDED.icon, sort_order=EXCLUDED.sort_order;

INSERT INTO lecturers (name,email,password,default_password,nidn,courses,is_password_changed) VALUES
('Dr. Imported Lecturer, M.Kom','imported@university.ac.id','password','password','9999999999',ARRAY['Imported Course 1','Imported Course 2'],false)
ON CONFLICT (email) DO UPDATE SET name=EXCLUDED.name,password=EXCLUDED.password,default_password=EXCLUDED.default_password,nidn=EXCLUDED.nidn,courses=EXCLUDED.courses;

INSERT INTO import_lecturer_preview_courses (lecturer_email, course_name, sort_order) VALUES
('imported@university.ac.id','Imported Course 1',1),('imported@university.ac.id','Imported Course 2',2)
ON CONFLICT DO NOTHING;

INSERT INTO course_sessions (id,course_id,session_number,title,topic,session_date,session_time,session_type,conference_link,room,description) VALUES
(1,1,1,'Sesi 1','Pengenalan Algoritma','8 Jan 2026','08:00 - 10:00','offline',NULL,'Lab 301','Pertemuan pertama membahas konsep dasar algoritma.'),
(2,1,2,'Sesi 2','Kompleksitas Algoritma','11 Jan 2026','08:00 - 10:00','online','https://meet.google.com/abc-defg-hij','Lab 301','Analisis kompleksitas waktu dan ruang.'),
(3,1,3,'Sesi 3','Sorting - Bubble & Selection','15 Jan 2026','08:00 - 10:00','offline',NULL,'Lab 301','Implementasi sorting dasar.')
ON CONFLICT (course_id,session_number) DO UPDATE SET title=EXCLUDED.title, topic=EXCLUDED.topic, session_date=EXCLUDED.session_date, session_time=EXCLUDED.session_time, session_type=EXCLUDED.session_type, conference_link=EXCLUDED.conference_link, room=EXCLUDED.room, description=EXCLUDED.description;
SELECT setval('course_sessions_id_seq', (SELECT COALESCE(MAX(id), 1) FROM course_sessions));

INSERT INTO course_materials (id,course_id,session_id,title,material_type,size,duration,downloads,upload_date,week,status) VALUES
(1,1,NULL,'Pengenalan Algoritma','pdf','2.4 MB','',145,'15 Des 2025',1,'completed'),
(2,1,NULL,'Kompleksitas Algoritma','video','45.2 MB','28 menit',98,'18 Des 2025',2,'completed'),
(3,2,NULL,'Slide Materi: Normalisasi Database','pdf','5.1 MB','',167,'10 Des 2025',NULL,'available'),
(4,3,NULL,'Video: RESTful API Best Practices','video','67.5 MB','42 menit',189,'20 Des 2025',NULL,'available'),
(5,1,1,'Slide Presentasi - Pengenalan Algoritma.pdf','PDF','2.5 MB','',0,'8 Jan 2026',NULL,'available'),
(6,1,1,'Video Tutorial - Kompleksitas Big O.mp4','Video','45 MB','',0,'8 Jan 2026',NULL,'available')
ON CONFLICT (id) DO UPDATE SET course_id=EXCLUDED.course_id, session_id=EXCLUDED.session_id, title=EXCLUDED.title, material_type=EXCLUDED.material_type, size=EXCLUDED.size, duration=EXCLUDED.duration, downloads=EXCLUDED.downloads, upload_date=EXCLUDED.upload_date, week=EXCLUDED.week, status=EXCLUDED.status;
SELECT setval('course_materials_id_seq', (SELECT COALESCE(MAX(id), 1) FROM course_materials));

INSERT INTO session_assignments (id,course_id,session_id,title,description,due_date,status,score,submitted_count,total_students) VALUES
(1,1,NULL,'Implementasi Bubble Sort','','20 Jan 2026','submitted',95,32,35),
(2,1,NULL,'Analisis Quick Sort','','27 Jan 2026','submitted',88,30,35),
(3,1,NULL,'Binary Search Tree','','10 Feb 2026','pending',NULL,0,35),
(4,1,1,'Latihan - Analisis Kompleksitas','','15 Jan 2026, 23:59','pending',NULL,0,35),
(5,1,1,'Quiz - Pengenalan Algoritma','','17 Jan 2026, 23:59','submitted',NULL,30,35),
(6,1,1,'Praktikum 1: Implementasi Algoritma Dasar','Buat program sederhana.','15 Jan 2026, 23:59','pending',NULL,32,35)
ON CONFLICT (id) DO UPDATE SET course_id=EXCLUDED.course_id, session_id=EXCLUDED.session_id, title=EXCLUDED.title, description=EXCLUDED.description, due_date=EXCLUDED.due_date, status=EXCLUDED.status, score=EXCLUDED.score, submitted_count=EXCLUDED.submitted_count, total_students=EXCLUDED.total_students;
SELECT setval('session_assignments_id_seq', (SELECT COALESCE(MAX(id), 1) FROM session_assignments));

INSERT INTO attendance_sessions (id,role_scope,course_code,course_name,class_name,session_number,session_date,session_time,room,lab,topic,total_students,present,absent,sick,permit,excused,status,assistant_status,assistant_check_in_time) VALUES
(1,'lecturer','TIF101','Pemrograman Dasar','TIF-A',1,'2025-01-13','08:00 - 10:00','Lab 301','','Perulangan dan Fungsi',35,32,1,1,1,0,'Selesai','',''),
(2,'lecturer','TIF102','Struktur Data','TIF-B',1,'2025-01-08','08:00 - 10:00','Lab 302','','Binary Search Tree',38,35,2,1,0,0,'Selesai','',''),
(3,'lecturer','TIF201','Basis Data','TIF-C',1,'2025-01-10','08:00 - 10:00','Lab 303','','Query SQL Lanjutan',32,30,1,0,1,0,'Selesai','',''),
(4,'assistant','TIF101','Pemrograman Dasar','TIF-A',1,'2025-01-13','14:00 - 16:00','','Lab 301','Perulangan dan Fungsi',35,32,1,1,1,0,'Selesai','',''),
(5,'assistant','TIF102','Struktur Data','TIF-B',1,'2025-01-08','14:00 - 16:00','','Lab 302','Binary Search Tree',38,35,2,1,0,0,'Selesai','',''),
(6,'assistant','TIF201','Basis Data','TIF-C',1,'2025-01-10','14:00 - 16:00','','Lab 303','Query SQL Lanjutan',32,30,1,0,1,0,'Selesai','',''),
(7,'admin','TIF101','Pemrograman Dasar','A1',1,'2025-01-06','08:00 - 10:00','Lab 301','','Pengenalan Pemrograman',35,33,1,0,0,1,'Selesai','Hadir','07:55')
ON CONFLICT (id) DO UPDATE SET role_scope=EXCLUDED.role_scope, course_code=EXCLUDED.course_code, course_name=EXCLUDED.course_name, class_name=EXCLUDED.class_name, session_number=EXCLUDED.session_number, session_date=EXCLUDED.session_date, session_time=EXCLUDED.session_time, room=EXCLUDED.room, lab=EXCLUDED.lab, topic=EXCLUDED.topic, total_students=EXCLUDED.total_students, present=EXCLUDED.present, absent=EXCLUDED.absent, sick=EXCLUDED.sick, permit=EXCLUDED.permit, excused=EXCLUDED.excused, status=EXCLUDED.status, assistant_status=EXCLUDED.assistant_status, assistant_check_in_time=EXCLUDED.assistant_check_in_time;
SELECT setval('attendance_sessions_id_seq', (SELECT COALESCE(MAX(id), 1) FROM attendance_sessions));

INSERT INTO attendance_records (session_id,nim,name,status,attendance_time,check_in_time) VALUES
(1,'210101001','Ahmad Fauzi','Hadir','08:05','08:05'),(1,'210101002','Siti Nurhaliza','Hadir','08:03','08:03'),(1,'210101003','Budi Santoso','Hadir','08:07','08:07'),(1,'210101004','Dewi Lestari','Izin','',''),(1,'210101005','Eko Prasetyo','Alpa','',''),
(4,'210101001','Ahmad Fauzi','Hadir','08:05','08:05'),(4,'210101002','Siti Nurhaliza','Hadir','08:03','08:03'),(4,'210101003','Budi Santoso','Hadir','08:07','08:07'),(4,'210101004','Dewi Lestari','Izin','',''),(4,'210101005','Eko Prasetyo','Alpa','',''),
(7,'210101001','Ahmad Fauzi','Hadir','','08:05'),(7,'210101002','Siti Nurhaliza','Izin','','')
ON CONFLICT DO NOTHING;

INSERT INTO lecturer_grade_courses (id,course_code,course_name,class_name,semester,academic_year,total_students,average_grade,highest_grade,lowest_grade,pass_rate) VALUES
(1,'TIF101','Pemrograman Dasar','TIF-A','Ganjil','2024/2025',35,78.5,95,65,94.3),
(2,'TIF102','Struktur Data','TIF-B','Ganjil','2024/2025',38,82.3,98,70,97.4),
(3,'TIF201','Basis Data','TIF-C','Ganjil','2024/2025',32,75.8,92,60,90.0)
ON CONFLICT (id) DO UPDATE SET average_grade=EXCLUDED.average_grade, highest_grade=EXCLUDED.highest_grade, lowest_grade=EXCLUDED.lowest_grade, pass_rate=EXCLUDED.pass_rate;
SELECT setval('lecturer_grade_courses_id_seq', (SELECT COALESCE(MAX(id), 1) FROM lecturer_grade_courses));

INSERT INTO lecturer_student_grades (id,grade_course_id,nim,name,tugas1,tugas2,tugas3,ujian_akhir,nilai_akhir,grade,status) VALUES
(1,1,'210101001','Ahmad Fauzi',85,80,82,82,81.7,'A','Lulus'),
(2,1,'210101002','Siti Nurhaliza',90,88,92,95,92.3,'A','Lulus'),
(3,1,'210101003','Budi Santoso',75,78,72,75,74.5,'B','Lulus')
ON CONFLICT (id) DO UPDATE SET grade_course_id=EXCLUDED.grade_course_id,nim=EXCLUDED.nim,name=EXCLUDED.name,tugas1=EXCLUDED.tugas1,tugas2=EXCLUDED.tugas2,tugas3=EXCLUDED.tugas3,ujian_akhir=EXCLUDED.ujian_akhir,nilai_akhir=EXCLUDED.nilai_akhir,grade=EXCLUDED.grade,status=EXCLUDED.status;
SELECT setval('lecturer_student_grades_id_seq', (SELECT COALESCE(MAX(id), 1) FROM lecturer_student_grades));

INSERT INTO assistant_practical_courses (id,course_code,name,lecturer,class_name,lab,day,start_time,end_time,semester,academic_year,students,attendance_present,attendance_total,color) VALUES
(1,'TIF101','Pemrograman Dasar','Dr. Ahmad Rahman','TIF-A','Lab 301','Senin','14:00','16:00','Ganjil','2025/2026',35,12,14,'bg-blue-500'),
(2,'TIF102','Struktur Data','Prof. Siti Nurhaliza','TIF-B','Lab 302','Rabu','14:00','16:00','Ganjil','2025/2026',38,13,14,'bg-blue-500'),
(3,'TIF201','Basis Data','Dr. Budi Santoso','TIF-C','Lab 303','Jumat','14:00','16:00','Ganjil','2025/2026',32,30,35,'bg-blue-500')
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name,lecturer=EXCLUDED.lecturer,class_name=EXCLUDED.class_name,lab=EXCLUDED.lab,day=EXCLUDED.day,start_time=EXCLUDED.start_time,end_time=EXCLUDED.end_time,semester=EXCLUDED.semester,academic_year=EXCLUDED.academic_year,students=EXCLUDED.students,attendance_present=EXCLUDED.attendance_present,attendance_total=EXCLUDED.attendance_total,color=EXCLUDED.color;
SELECT setval('assistant_practical_courses_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assistant_practical_courses));

INSERT INTO assistant_tasks (id,task,submitted,total,deadline) VALUES
(1,'Verifikasi laporan praktikum Pemrograman Dasar',28,35,'2 hari lagi'),(2,'Input nilai praktikum Struktur Data',35,38,'3 hari lagi')
ON CONFLICT (id) DO UPDATE SET task=EXCLUDED.task,submitted=EXCLUDED.submitted,total=EXCLUDED.total,deadline=EXCLUDED.deadline;
SELECT setval('assistant_tasks_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assistant_tasks));

INSERT INTO assistant_reports (id,nim,name,course_code,course_name,class_name,week,topic,submitted_at,status,score,file_name,file_size) VALUES
(1,'210101001','Ahmad Fauzi','TIF101','Pemrograman Dasar','TIF-A',5,'Array dan Matrix','2025-01-13 15:45','Menunggu Review',NULL,'Laporan_Array_Ahmad.pdf','2.4 MB'),
(2,'210101002','Siti Nurhaliza','TIF101','Pemrograman Dasar','TIF-A',5,'Array dan Matrix','2025-01-13 15:30','Disetujui',90,'Laporan_Array_Siti.pdf','1.8 MB')
ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status,score=EXCLUDED.score;
SELECT setval('assistant_reports_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assistant_reports));

INSERT INTO assistant_report_summary (id,course_code,course_name,class_name,total_reports,reviewed,pending,approved,needs_revision) VALUES
(1,'TIF101','Pemrograman Dasar','TIF-A',35,28,7,25,3),
(2,'TIF102','Struktur Data','TIF-B',38,32,6,30,2)
ON CONFLICT (id) DO UPDATE SET total_reports=EXCLUDED.total_reports,reviewed=EXCLUDED.reviewed,pending=EXCLUDED.pending,approved=EXCLUDED.approved,needs_revision=EXCLUDED.needs_revision;
SELECT setval('assistant_report_summary_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assistant_report_summary));

INSERT INTO admin_activities (id,action,detail,activity_time,icon,sort_order) VALUES
(1,'Mahasiswa baru terdaftar','Ahmad Fauzi - Teknik Informatika','5 menit lalu','user',1),
(2,'Tugas baru dibuat','Tugas Algoritma & Pemrograman','15 menit lalu','file',2),
(3,'Kursus diperbarui','Database Management - Materi Week 5','1 jam lalu','book',3),
(4,'Nilai diinput','UTS Pemrograman Web - 45 mahasiswa','2 jam lalu','award',4)
ON CONFLICT (id) DO UPDATE SET action=EXCLUDED.action,detail=EXCLUDED.detail,activity_time=EXCLUDED.activity_time,icon=EXCLUDED.icon,sort_order=EXCLUDED.sort_order;
SELECT setval('admin_activities_id_seq', (SELECT COALESCE(MAX(id), 1) FROM admin_activities));
