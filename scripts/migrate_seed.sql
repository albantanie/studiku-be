-- Full database migration + seed for StudiKu.
-- Run: go run main.go migrate seed

BEGIN;


-- >>> 001_init.up.sql
CREATE TABLE IF NOT EXISTS academic_years (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  semester VARCHAR(20) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'Mendatang',
  total_courses INT NOT NULL DEFAULT 0,
  total_students INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS courses (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  instructor VARCHAR(255) NOT NULL,
  assistant VARCHAR(255) NOT NULL,
  study_program VARCHAR(255) NOT NULL DEFAULT 'Teknik Informatika',
  academic_year VARCHAR(100) NOT NULL DEFAULT 'Genap 2024/2025',
  class_code VARCHAR(50) NOT NULL DEFAULT '',
  status VARCHAR(20) NOT NULL DEFAULT 'Aktif',
  day VARCHAR(50) NOT NULL DEFAULT '',
  start_time VARCHAR(10) NOT NULL DEFAULT '',
  end_time VARCHAR(10) NOT NULL DEFAULT '',
  room VARCHAR(100) NOT NULL DEFAULT '',
  sessions INT NOT NULL DEFAULT 14 CHECK (sessions > 0),
  credits INT NOT NULL DEFAULT 3 CHECK (credits > 0),
  students INT NOT NULL DEFAULT 0 CHECK (students >= 0),
  attendance_present INT NOT NULL DEFAULT 0,
  attendance_total INT NOT NULL DEFAULT 12,
  color VARCHAR(50) NOT NULL DEFAULT 'bg-blue-500',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS assignments (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  course VARCHAR(255) NOT NULL,
  assistant VARCHAR(255) NOT NULL,
  due_date VARCHAR(50) NOT NULL,
  due_time VARCHAR(10) NOT NULL DEFAULT '23:59',
  status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','submitted','graded')),
  course_color VARCHAR(50) NOT NULL DEFAULT 'bg-blue-500',
  submitted_date VARCHAR(50),
  score INT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS grades (
  id SERIAL PRIMARY KEY,
  course_name VARCHAR(255) NOT NULL,
  code VARCHAR(50) NOT NULL,
  semester VARCHAR(50) NOT NULL,
  credits INT NOT NULL CHECK (credits > 0),
  grade VARCHAR(5) NOT NULL,
  score INT NOT NULL CHECK (score >= 0 AND score <= 100),
  color VARCHAR(50) NOT NULL DEFAULT 'bg-blue-500',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS students (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  default_password VARCHAR(255) NOT NULL,
  student_id VARCHAR(50) NOT NULL UNIQUE,
  program VARCHAR(255) NOT NULL,
  semester INT NOT NULL CHECK (semester > 0),
  courses TEXT[] NOT NULL DEFAULT '{}',
  status VARCHAR(50) NOT NULL DEFAULT 'Aktif',
  join_date DATE NOT NULL DEFAULT CURRENT_DATE,
  is_password_changed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lecturers (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL DEFAULT 'password',
  default_password VARCHAR(255) NOT NULL DEFAULT 'password',
  nidn VARCHAR(50) NOT NULL UNIQUE,
  courses TEXT[] NOT NULL DEFAULT '{}',
  is_password_changed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS lab_assistants (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  phone VARCHAR(50) NOT NULL DEFAULT '',
  student_id VARCHAR(50) NOT NULL UNIQUE,
  role VARCHAR(30) NOT NULL DEFAULT 'aslab',
  lab VARCHAR(255) NOT NULL,
  supervisor VARCHAR(255) NOT NULL DEFAULT '',
  semester INT NOT NULL DEFAULT 1,
  gpa NUMERIC(3,2) NOT NULL DEFAULT 0,
  assigned_courses INT NOT NULL DEFAULT 0,
  weekly_hours INT NOT NULL DEFAULT 0,
  status VARCHAR(50) NOT NULL DEFAULT 'Aktif',
  join_date DATE NOT NULL DEFAULT CURRENT_DATE,
  password VARCHAR(255) NOT NULL DEFAULT 'password',
  default_password VARCHAR(255) NOT NULL DEFAULT 'password',
  is_password_changed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE lab_assistants ADD COLUMN IF NOT EXISTS role VARCHAR(30) NOT NULL DEFAULT 'aslab';

CREATE TABLE IF NOT EXISTS classes (
  id SERIAL PRIMARY KEY,
  code VARCHAR(50) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  academic_year VARCHAR(50) NOT NULL,
  assistant VARCHAR(255) NOT NULL DEFAULT '',
  schedule VARCHAR(255) NOT NULL DEFAULT '',
  room VARCHAR(100) NOT NULL DEFAULT '',
  total_students INT NOT NULL DEFAULT 0,
  capacity INT NOT NULL CHECK (capacity > 0),
  students INT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- <<< 001_init.up.sql

-- >>> 002_seed.up.sql
INSERT INTO academic_years (id,name,start_date,end_date,semester,status,total_courses,total_students) VALUES
(1,'2024/2025 Genap','2025-01-15','2025-06-30','Genap','Aktif',48,1234),
(2,'2024/2025 Ganjil','2024-08-15','2024-12-31','Ganjil','Selesai',45,1189),
(3,'2025/2026 Ganjil','2025-08-15','2025-12-31','Ganjil','Mendatang',0,0),
(4,'2023/2024 Genap','2024-01-15','2024-06-30','Genap','Selesai',42,1098)
ON CONFLICT (id) DO NOTHING;
SELECT setval('academic_years_id_seq', (SELECT COALESCE(MAX(id), 1) FROM academic_years));

INSERT INTO courses (id,name,instructor,assistant,study_program,academic_year,class_code,status,day,start_time,end_time,room,sessions,credits,students,attendance_present,attendance_total,color) VALUES
(1,'Algoritma & Struktur Data','Dr. Ahmad Rahman','Rama Dhani','Teknik Informatika','2025/2026','CS301','Aktif','Senin & Kamis','08:00','10:00','Lab 301',14,4,35,10,12,'bg-blue-500'),
(2,'Basis Data Lanjutan','Prof. Siti Nurhaliza','Dina Amelia','Teknik Informatika','2025/2026','CS302','Aktif','Selasa','10:00','13:00','Ruang 402',14,3,38,12,12,'bg-blue-500'),
(3,'Pemrograman Web','Ir. Budi Hartono','Rama Dhani','Teknik Informatika','Genap 2024/2025','TI-201','Aktif','Rabu & Jumat','13:00','15:00','Lab Komputer 1',14,3,40,11,12,'bg-blue-500'),
(4,'Jaringan Komputer','Dr. Rina Kusuma','Fahmi Akbar','Teknik Informatika','Genap 2024/2025','CS304','Aktif','Kamis','15:00','18:00','Lab 304',14,3,32,9,12,'bg-blue-500'),
(5,'Rekayasa Perangkat Lunak','Prof. Agus Setiawan','Rama Dhani','Teknik Informatika','Genap 2024/2025','TI-305','Aktif','Selasa','08:00','10:00','Ruang 402',12,3,28,12,12,'bg-blue-500')
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name,instructor=EXCLUDED.instructor,assistant=EXCLUDED.assistant,study_program=EXCLUDED.study_program,academic_year=EXCLUDED.academic_year,class_code=EXCLUDED.class_code,status=EXCLUDED.status,day=EXCLUDED.day,start_time=EXCLUDED.start_time,end_time=EXCLUDED.end_time,room=EXCLUDED.room,sessions=EXCLUDED.sessions,credits=EXCLUDED.credits,students=EXCLUDED.students,attendance_present=EXCLUDED.attendance_present,attendance_total=EXCLUDED.attendance_total,color=EXCLUDED.color;
SELECT setval('courses_id_seq', (SELECT COALESCE(MAX(id), 1) FROM courses));

INSERT INTO assignments (id,title,course,assistant,due_date,due_time,status,course_color,submitted_date,score) VALUES
(1,'Project Akhir Basis Data','Basis Data Lanjutan','Dr. Ahmad Fauzi, M.Kom','5 Jan 2026','23:59','pending','bg-blue-500',NULL,NULL),
(2,'Quiz Sorting Algorithms','Algoritma & Struktur Data','Rina Susanti, S.Kom','7 Jan 2026','23:59','pending','bg-blue-500',NULL,NULL),
(3,'Tugas Routing Protocol','Jaringan Komputer','Budi Prasetyo, M.T','10 Jan 2026','23:59','pending','bg-blue-500',NULL,NULL),
(4,'Implementasi REST API','Pemrograman Web','Sarah Anggraini, S.Kom','12 Jan 2026','23:59','pending','bg-blue-500',NULL,NULL),
(5,'Essay Arsitektur Microservices','Pemrograman Web','Sarah Anggraini, S.Kom','15 Jan 2026','23:59','pending','bg-blue-500',NULL,NULL),
(6,'Implementasi Binary Search Tree','Algoritma & Struktur Data','Rina Susanti, S.Kom','28 Des 2025','23:59','submitted','bg-blue-500','27 Des 2025',NULL),
(7,'Laporan Praktikum Subnetting','Jaringan Komputer','Budi Prasetyo, M.T','20 Des 2025','23:59','graded','bg-blue-500',NULL,92)
ON CONFLICT (id) DO NOTHING;
SELECT setval('assignments_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assignments));

INSERT INTO grades (id,course_name,code,semester,credits,grade,score,color) VALUES
(1,'Algoritma & Struktur Data','CS301','2025/2026',4,'A',88,'bg-blue-500'),
(2,'Basis Data Lanjutan','CS302','2025/2026',3,'A-',85,'bg-blue-500'),
(3,'Pemrograman Web','CS303','2025/2026',3,'B+',78,'bg-blue-500'),
(4,'Jaringan Komputer','CS304','2025/2026',3,'A',90,'bg-blue-500'),
(5,'Interaksi Manusia Komputer','CS201','2024/2025',3,'A',92,'bg-blue-500'),
(6,'Sistem Operasi','CS202','2024/2025',4,'B+',82,'bg-blue-500')
ON CONFLICT (id) DO NOTHING;
SELECT setval('grades_id_seq', (SELECT COALESCE(MAX(id), 1) FROM grades));

INSERT INTO students (id,name,email,password,default_password,student_id,program,semester,courses,status,join_date,is_password_changed) VALUES
(1,'Ahmad Rizki','ahmad.rizki@student.ac.id','password','password','TI2021001','Teknik Informatika',6,ARRAY['Pemrograman Web','Basis Data','Algoritma'],'Aktif','2021-08-15',true),
(2,'Siti Nurhaliza','siti.nurhaliza@student.ac.id','password','password','TI2021002','Teknik Informatika',6,ARRAY['Pemrograman Web','Jaringan Komputer'],'Aktif','2021-08-15',false),
(3,'Budi Santoso','budi.santoso@student.ac.id','password','password','TI2022001','Teknik Informatika',4,ARRAY['Analisis Sistem','Manajemen Proyek'],'Aktif','2022-08-20',false),
(4,'Dewi Lestari','dewi.lestari@student.ac.id','password','password','TI2021003','Teknik Informatika',6,ARRAY['Desain Grafis','Tipografi'],'Cuti','2021-08-15',true),
(5,'Eko Prasetyo','eko.prasetyo@student.ac.id','password','password','TI2023001','Teknik Informatika',2,ARRAY['Dasar Pemrograman','Matematika Diskrit'],'Aktif','2023-08-25',false)
ON CONFLICT (id) DO NOTHING;
SELECT setval('students_id_seq', (SELECT COALESCE(MAX(id), 1) FROM students));

INSERT INTO lecturers (id,name,email,password,default_password,nidn,courses,is_password_changed) VALUES
(1,'Dr. Budi Santoso, M.Kom','budi.santoso@university.ac.id','password','password','0123456789',ARRAY['Pemrograman Web','Basis Data','Rekayasa Perangkat Lunak'],false),
(2,'Dr. Siti Aminah, M.T','siti.aminah@university.ac.id','password','password','0123456790',ARRAY['Algoritma dan Struktur Data','Matematika Diskrit'],false),
(3,'Ir. Ahmad Fauzi, M.Sc','ahmad.fauzi@university.ac.id','password','password','0123456791',ARRAY['Jaringan Komputer','Sistem Operasi','Keamanan Jaringan'],false),
(4,'Dr. Dewi Lestari, M.Kom','dewi.lestari@university.ac.id','password','password','0123456792',ARRAY['Kecerdasan Buatan','Machine Learning'],false)
ON CONFLICT (id) DO NOTHING;
SELECT setval('lecturers_id_seq', (SELECT COALESCE(MAX(id), 1) FROM lecturers));

INSERT INTO lab_assistants (id,name,email,phone,student_id,role,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date,password,default_password,is_password_changed) VALUES
(1,'Rama Dhani','rama.dhani@student.ac.id','081234567890','TI2021001','aslab','Laboratorium Pemrograman','Dr. Budi Santoso',7,3.85,2,12,'Aktif','2024-09-01','password','password',false),
(2,'Dina Amelia','dina.amelia@student.ac.id','082345678901','TI2021002','aslab','Laboratorium Database','Prof. Siti Aminah',7,3.92,1,8,'Aktif','2024-09-01','password','password',false),
(3,'Fahmi Akbar','fahmi.akbar@student.ac.id','083456789012','TI2021003','aslab','Laboratorium Jaringan','Dr. Ahmad Wijaya',6,3.78,2,10,'Aktif','2024-09-01','password','password',false),
(4,'Sari Laboran','laboran@app.com','081200000002','DEMO-LABORAN-001','laboran','Laboratorium Pemrograman','Kepala Lab Informatika',1,0,0,40,'Aktif','2024-09-01','password','password',false),
(5,'Kepala Lab Informatika','kalab@app.com','081200000003','DEMO-KALAB-001','kalab','Kepala Laboratorium','-',1,0,0,40,'Aktif','2024-09-01','password','password',false)
ON CONFLICT (id) DO NOTHING;
SELECT setval('lab_assistants_id_seq', (SELECT COALESCE(MAX(id), 1) FROM lab_assistants));

INSERT INTO classes (id,code,name,academic_year,assistant,schedule,room,total_students,capacity,students) VALUES
(1,'CS301','Algoritma & Struktur Data','2025/2026','Rama Dhani','Senin & Kamis, 08:00 - 10:00','Lab 301',35,40,ARRAY[1,2,3,4,5]),
(2,'CS302','Basis Data Lanjutan','2025/2026','Dina Amelia','Selasa, 10:00 - 13:00','Ruang 402',38,40,ARRAY[6,7]),
(3,'TI-201','Pemrograman Web','Genap 2024/2025','Rama Dhani','Rabu & Jumat, 13:00 - 15:00','Lab Komputer 1',40,40,ARRAY[]::int[]),
(4,'CS304','Jaringan Komputer','Genap 2024/2025','Fahmi Akbar','Kamis, 15:00 - 18:00','Lab 304',32,40,ARRAY[]::int[]),
(5,'TI-305','Rekayasa Perangkat Lunak','Genap 2024/2025','Rama Dhani','Selasa, 08:00 - 10:00','Ruang 402',28,35,ARRAY[]::int[])
ON CONFLICT (id) DO UPDATE SET code=EXCLUDED.code,name=EXCLUDED.name,academic_year=EXCLUDED.academic_year,assistant=EXCLUDED.assistant,schedule=EXCLUDED.schedule,room=EXCLUDED.room,total_students=EXCLUDED.total_students,capacity=EXCLUDED.capacity,students=EXCLUDED.students;
SELECT setval('classes_id_seq', (SELECT COALESCE(MAX(id), 1) FROM classes));

-- <<< 002_seed.up.sql

-- >>> 003_page_data_legacy_noop.up.sql
-- Deprecated no-op migration. Page data lives in normalized domain tables from migration 004.

-- <<< 003_page_data_legacy_noop.up.sql

-- >>> 004_normalized_page_data.up.sql
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
  sort_order INT NOT NULL DEFAULT 0,
  UNIQUE(course_id, session_number)
);

CREATE TABLE IF NOT EXISTS course_materials (
  id SERIAL PRIMARY KEY,
  course_id INT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  session_id INT REFERENCES course_sessions(id) ON DELETE SET NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  material_type VARCHAR(50) NOT NULL,
  size VARCHAR(50) NOT NULL DEFAULT '',
  duration VARCHAR(50) NOT NULL DEFAULT '',
  downloads INT NOT NULL DEFAULT 0,
  upload_date VARCHAR(50) NOT NULL DEFAULT '',
  week INT,
  status VARCHAR(50) NOT NULL DEFAULT 'available'
);

ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS file_path VARCHAR(500) NOT NULL DEFAULT '';
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS created_by INT;
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMP;
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS approved_by INT;
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP;
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS rejected_by INT;
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMP;
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS rejection_note TEXT NOT NULL DEFAULT '';

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

-- Jawaban tugas mahasiswa. Dibuat di sini karena trigger sinkronisasi nilai di
-- bawah menempel pada tabel ini, jadi harus ada sebelum trigger dibuat.
CREATE TABLE IF NOT EXISTS student_submissions (
  id SERIAL PRIMARY KEY,
  assignment_id INT NOT NULL REFERENCES session_assignments(id) ON DELETE CASCADE,
  student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  answer_text TEXT NOT NULL DEFAULT '',
  file_url TEXT NOT NULL DEFAULT '',
  file_name VARCHAR(255) NOT NULL DEFAULT '',
  file_size VARCHAR(50) NOT NULL DEFAULT '',
  submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  score INT,
  feedback TEXT NOT NULL DEFAULT '',
  UNIQUE(assignment_id, student_id)
);

CREATE TABLE IF NOT EXISTS session_assessments (
  id SERIAL PRIMARY KEY,
  session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
  assessment_type VARCHAR(20) NOT NULL CHECK (assessment_type IN ('pretest','posttest')),
  title VARCHAR(255) NOT NULL DEFAULT '',
  score INT,
  max_score INT NOT NULL DEFAULT 100,
  status VARCHAR(50) NOT NULL DEFAULT 'not_started',
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(session_id, assessment_type)
);

CREATE TABLE IF NOT EXISTS session_assessment_results (
  id SERIAL PRIMARY KEY,
  session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
  student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  assessment_type VARCHAR(20) NOT NULL CHECK (assessment_type IN ('pretest','posttest')),
  score INT,
  max_score INT NOT NULL DEFAULT 100,
  status VARCHAR(50) NOT NULL DEFAULT 'not_started',
  note TEXT NOT NULL DEFAULT '',
  submitted_at TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(session_id, student_id, assessment_type)
);

-- Bank soal pretest dan post-test, dibedakan oleh assessment_type.
-- Dibuat di sini supaya migrate_seed berdiri sendiri, tidak bergantung pada
-- pembuatan skema oleh kode Go yang hanya jalan saat perintah rest.
CREATE TABLE IF NOT EXISTS session_pretest_questions (
  id SERIAL PRIMARY KEY,
  session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
  assessment_type VARCHAR(20) NOT NULL DEFAULT 'pretest' CHECK (assessment_type IN ('pretest','posttest')),
  question TEXT NOT NULL,
  options JSONB NOT NULL DEFAULT '[]'::jsonb,
  correct_option INT NOT NULL DEFAULT 0,
  points INT NOT NULL DEFAULT 10,
  explanation TEXT NOT NULL DEFAULT '',
  sort_order INT NOT NULL DEFAULT 0,
  created_by INT REFERENCES lab_assistants(id) ON DELETE SET NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS session_pretest_submissions (
  id SERIAL PRIMARY KEY,
  session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
  student_id INT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
  assessment_type VARCHAR(20) NOT NULL DEFAULT 'pretest' CHECK (assessment_type IN ('pretest','posttest')),
  answers JSONB NOT NULL DEFAULT '[]'::jsonb,
  score INT NOT NULL DEFAULT 0,
  max_score INT NOT NULL DEFAULT 100,
  status VARCHAR(50) NOT NULL DEFAULT 'not_started',
  submitted_at TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Keunikan (session,student,type) dipasang lewat CREATE UNIQUE INDEX di bawah,
-- sama seperti skema Go, supaya DB lama dan baru sama-sama punya satu index saja.

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
  course_id INT REFERENCES courses(id) ON DELETE SET NULL,
  created_by INT,
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
  file_size VARCHAR(50) NOT NULL,
  file_path VARCHAR(500) NOT NULL DEFAULT '',
  submitted_to_laboran_at TIMESTAMP,
  laboran_approved_by INT,
  laboran_approved_at TIMESTAMP,
  submitted_to_kalab_at TIMESTAMP,
  kalab_approved_by INT,
  kalab_approved_at TIMESTAMP,
  rejected_by INT,
  rejected_at TIMESTAMP,
  rejection_note TEXT NOT NULL DEFAULT '',
  returned_to_role VARCHAR(30) NOT NULL DEFAULT ''
);

ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS course_id INT REFERENCES courses(id) ON DELETE SET NULL;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS created_by INT;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS file_path VARCHAR(500) NOT NULL DEFAULT '';
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS submitted_to_laboran_at TIMESTAMP;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS laboran_approved_by INT;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS laboran_approved_at TIMESTAMP;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS submitted_to_kalab_at TIMESTAMP;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS kalab_approved_by INT;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS kalab_approved_at TIMESTAMP;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS rejected_by INT;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMP;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS rejection_note TEXT NOT NULL DEFAULT '';
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS returned_to_role VARCHAR(30) NOT NULL DEFAULT '';

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

CREATE TABLE IF NOT EXISTS report_workflows (
  course_id INT PRIMARY KEY,
  status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','SUBMITTED','APPROVED','REJECTED')),
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS institution_settings (
  id SERIAL PRIMARY KEY,
  university_name VARCHAR(255) NOT NULL,
  faculty_name VARCHAR(255) NOT NULL DEFAULT '',
  study_program_name VARCHAR(255) NOT NULL DEFAULT '',
  laboratory_name VARCHAR(255) NOT NULL DEFAULT '',
  campus_a_address TEXT NOT NULL DEFAULT '',
  campus_b_address TEXT NOT NULL DEFAULT '',
  website VARCHAR(255) NOT NULL DEFAULT '',
  email VARCHAR(255) NOT NULL DEFAULT '',
  phone VARCHAR(255) NOT NULL DEFAULT '',
  logo_path VARCHAR(500) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS course_assessment_weights (
  id SERIAL PRIMARY KEY,
  course_id INT REFERENCES courses(id) ON DELETE CASCADE,
  attendance_weight NUMERIC(5,2) NOT NULL DEFAULT 10,
  pretest_weight NUMERIC(5,2) NOT NULL DEFAULT 15,
  assignment_weight NUMERIC(5,2) NOT NULL DEFAULT 20,
  practicum_weight NUMERIC(5,2) NOT NULL DEFAULT 20,
  posttest_weight NUMERIC(5,2) NOT NULL DEFAULT 35,
  passing_grade NUMERIC(5,2) NOT NULL DEFAULT 55,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(course_id)
);

CREATE TABLE IF NOT EXISTS grade_scales (
  id SERIAL PRIMARY KEY,
  min_score NUMERIC(5,2) NOT NULL,
  max_score NUMERIC(5,2) NOT NULL,
  grade VARCHAR(5) NOT NULL,
  is_passed BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS report_signers (
  id SERIAL PRIMARY KEY,
  role VARCHAR(80) NOT NULL,
  name VARCHAR(255) NOT NULL,
  identifier_type VARCHAR(50) NOT NULL DEFAULT 'NIDN',
  identifier_number VARCHAR(80) NOT NULL DEFAULT '',
  signature_path VARCHAR(500) NOT NULL DEFAULT '',
  is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS student_activity_logs (
  id SERIAL PRIMARY KEY,
  student_id INT REFERENCES students(id) ON DELETE SET NULL,
  course_id INT REFERENCES courses(id) ON DELETE SET NULL,
  session_id INT REFERENCES course_sessions(id) ON DELETE SET NULL,
  activity_type VARCHAR(80) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  ip_address VARCHAR(80) NOT NULL DEFAULT '',
  user_agent TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
('user_tabs','Mahasiswa','students','graduation',1),('user_tabs','Lainnya','lecturers','users',2),('user_tabs','Asisten Laboratorium','assistants','userCog',3),
('import_file_types','text/csv','text/csv','',1),('import_file_types','application/vnd.ms-excel','application/vnd.ms-excel','',2),('import_file_types','application/vnd.openxmlformats-officedocument.spreadsheetml.sheet','application/vnd.openxmlformats-officedocument.spreadsheetml.sheet','',3)
ON CONFLICT (option_group,value) DO UPDATE SET label=EXCLUDED.label, icon=EXCLUDED.icon, sort_order=EXCLUDED.sort_order;

INSERT INTO lecturers (name,email,password,default_password,nidn,courses,is_password_changed) VALUES
('Dr. Imported Lecturer, M.Kom','imported@university.ac.id','password','password','9999999999',ARRAY['Imported Course 1','Imported Course 2'],false)
ON CONFLICT (email) DO UPDATE SET name=EXCLUDED.name,password=EXCLUDED.password,default_password=EXCLUDED.default_password,nidn=EXCLUDED.nidn,courses=EXCLUDED.courses;

INSERT INTO import_lecturer_preview_courses (lecturer_email, course_name, sort_order) VALUES
('imported@university.ac.id','Imported Course 1',1),('imported@university.ac.id','Imported Course 2',2)
ON CONFLICT DO NOTHING;

INSERT INTO course_sessions (course_id,session_number,title,topic,session_date,session_time,session_type,conference_link,room,description,sort_order)
SELECT
  c.id,
  n,
  'Sesi ' || n,
  CASE
    WHEN n = 1 THEN 'Pengenalan ' || c.name
    WHEN n = 2 THEN 'Konfigurasi ' || c.name
    WHEN n = 3 THEN 'Praktikum ' || c.name
    ELSE 'Pertemuan ' || n || ' - ' || c.name
  END,
  to_char(DATE '2026-01-08' + ((n - 1) * INTERVAL '7 days'), 'YYYY-MM-DD'),
  CASE WHEN c.start_time = '' AND c.end_time = '' THEN c.day ELSE c.start_time || ' - ' || c.end_time END,
  'offline',
  NULL,
  c.room,
  'Sesi otomatis dari konfigurasi jumlah sesi kursus admin.',
  n
FROM courses c
CROSS JOIN generate_series(1, c.sessions) AS gs(n)
ON CONFLICT (course_id,session_number) DO UPDATE SET title=EXCLUDED.title, topic=EXCLUDED.topic, session_date=EXCLUDED.session_date, session_time=EXCLUDED.session_time, session_type=EXCLUDED.session_type, conference_link=EXCLUDED.conference_link, room=EXCLUDED.room, description=EXCLUDED.description, sort_order=EXCLUDED.sort_order;
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
(4,1,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),'Latihan - Analisis Kompleksitas','','15 Jan 2026, 23:59','pending',NULL,0,35),
(5,1,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=2),'Quiz - Pengenalan Algoritma','','17 Jan 2026, 23:59','submitted',NULL,30,35),
(6,1,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=3),'Praktikum 1: Implementasi Algoritma Dasar','Buat program sederhana.','15 Jan 2026, 23:59','pending',NULL,32,35)
ON CONFLICT (id) DO UPDATE SET course_id=EXCLUDED.course_id, session_id=EXCLUDED.session_id, title=EXCLUDED.title, description=EXCLUDED.description, due_date=EXCLUDED.due_date, status=EXCLUDED.status, score=EXCLUDED.score, submitted_count=EXCLUDED.submitted_count, total_students=EXCLUDED.total_students;
SELECT setval('session_assignments_id_seq', (SELECT COALESCE(MAX(id), 1) FROM session_assignments));

INSERT INTO session_assessments (session_id,assessment_type,title,score,max_score,status,note)
SELECT
  cs.id,
  assessment.assessment_type,
  CASE
    WHEN assessment.assessment_type = 'pretest' THEN 'Pretest - ' || cs.topic
    ELSE 'Post-test - ' || cs.topic
  END,
  CASE
    WHEN cs.session_number = 1 AND assessment.assessment_type = 'pretest' THEN 70 + (cs.course_id % 10)
    WHEN cs.session_number = 1 AND assessment.assessment_type = 'posttest' THEN 82 + (cs.course_id % 10)
    WHEN cs.session_number = 2 AND assessment.assessment_type = 'pretest' THEN 68 + (cs.course_id % 10)
    ELSE NULL
  END,
  100,
  CASE
    WHEN cs.session_number = 1 THEN 'completed'
    WHEN cs.session_number = 2 AND assessment.assessment_type = 'pretest' THEN 'completed'
    ELSE 'not_started'
  END,
  ''
FROM course_sessions cs
CROSS JOIN (VALUES ('pretest'), ('posttest')) AS assessment(assessment_type)
WHERE cs.session_number <= 3
ON CONFLICT (session_id,assessment_type) DO UPDATE SET title=EXCLUDED.title,score=EXCLUDED.score,max_score=EXCLUDED.max_score,status=EXCLUDED.status,note=EXCLUDED.note,updated_at=now();

INSERT INTO session_assessment_results (session_id,student_id,assessment_type,score,max_score,status,note,submitted_at)
SELECT
  cs.id,
  st.id,
  assessment.assessment_type,
  CASE
    WHEN cs.session_number = 1 AND assessment.assessment_type = 'pretest' THEN 62 + (st.id % 18)
    WHEN cs.session_number = 1 AND assessment.assessment_type = 'posttest' THEN 76 + (st.id % 16)
    WHEN cs.session_number = 2 AND assessment.assessment_type = 'pretest' THEN 58 + (st.id % 20)
    WHEN cs.session_number = 2 AND assessment.assessment_type = 'posttest' THEN 72 + (st.id % 18)
    WHEN cs.session_number = 3 AND assessment.assessment_type = 'pretest' THEN 60 + (st.id % 18)
    ELSE NULL
  END,
  100,
  CASE
    WHEN cs.session_number <= 2 THEN 'completed'
    WHEN cs.session_number = 3 AND assessment.assessment_type = 'pretest' THEN 'completed'
    ELSE 'not_started'
  END,
  CASE WHEN assessment.assessment_type = 'pretest' THEN 'Nilai awal mahasiswa' ELSE 'Nilai akhir mahasiswa' END,
  CASE
    WHEN cs.session_number <= 2 THEN CURRENT_TIMESTAMP
    WHEN cs.session_number = 3 AND assessment.assessment_type = 'pretest' THEN CURRENT_TIMESTAMP
    ELSE NULL
  END
FROM course_sessions cs
JOIN courses c ON c.id=cs.course_id
JOIN classes cls ON cls.code=c.class_code
JOIN students st ON st.id=ANY(cls.students)
CROSS JOIN (VALUES ('pretest'), ('posttest')) AS assessment(assessment_type)
WHERE cs.session_number <= 3
ON CONFLICT (session_id,student_id,assessment_type) DO UPDATE SET
  score=EXCLUDED.score,
  max_score=EXCLUDED.max_score,
  status=EXCLUDED.status,
  note=EXCLUDED.note,
  submitted_at=EXCLUDED.submitted_at,
  updated_at=now();

-- Bank soal dipakai bersama oleh pretest dan post-test lewat kolom assessment_type.
ALTER TABLE session_pretest_questions ADD COLUMN IF NOT EXISTS assessment_type VARCHAR(20) NOT NULL DEFAULT 'pretest';
ALTER TABLE session_pretest_submissions ADD COLUMN IF NOT EXISTS assessment_type VARCHAR(20) NOT NULL DEFAULT 'pretest';
ALTER TABLE session_pretest_submissions DROP CONSTRAINT IF EXISTS session_pretest_submissions_session_id_student_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS session_pretest_submissions_unique_idx ON session_pretest_submissions (session_id, student_id, assessment_type);
CREATE INDEX IF NOT EXISTS session_pretest_questions_lookup_idx ON session_pretest_questions (session_id, assessment_type, sort_order);

INSERT INTO session_pretest_questions (id,session_id,assessment_type,question,options,correct_option,points,explanation,sort_order,created_by) VALUES
(1,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),'pretest','Apa tujuan pretest dalam praktikum?','["Mengukur pemahaman awal","Menilai akhir pembelajaran","Mengganti presensi","Membuat laporan akhir"]'::jsonb,0,10,'Pretest dipakai untuk mengukur pemahaman awal mahasiswa.',1,1),
(2,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),'pretest','Apa keluaran utama dari pretest?','["Nilai awal","Nilai tugas","Nilai presensi","Nilai laporan"]'::jsonb,0,10,'Pretest menghasilkan nilai awal sebelum pembelajaran dimulai.',2,1),
(3,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),'pretest','Kapan pretest dikerjakan?','["Sebelum materi","Setelah post-test","Saat upload tugas","Setelah laporan"]'::jsonb,0,10,'Pretest dikerjakan sebelum mahasiswa mempelajari materi sesi.',3,1),
(4,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),'posttest','Setelah praktikum, apa fungsi post-test?','["Mengukur pemahaman akhir","Mengukur pemahaman awal","Mengganti presensi","Menentukan jadwal"]'::jsonb,0,10,'Post-test mengukur pemahaman akhir setelah materi dipelajari.',1,1),
(5,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),'posttest','Nilai pretest dan post-test dipakai untuk menghitung apa?','["N-Gain","Presensi","Jumlah SKS","Kuota kelas"]'::jsonb,0,10,'Selisih ternormalisasi pretest dan post-test menghasilkan N-Gain.',2,1),
(6,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),'posttest','Nilai N-Gain 0,75 termasuk kategori apa?','["Tinggi","Sedang","Rendah","Tidak valid"]'::jsonb,0,10,'Menurut Hake, N-Gain >= 0,70 termasuk kategori tinggi.',3,1)
ON CONFLICT (id) DO UPDATE SET assessment_type=EXCLUDED.assessment_type,question=EXCLUDED.question,options=EXCLUDED.options,correct_option=EXCLUDED.correct_option,points=EXCLUDED.points,explanation=EXCLUDED.explanation,sort_order=EXCLUDED.sort_order,created_by=EXCLUDED.created_by,updated_at=now();
SELECT setval('session_pretest_questions_id_seq', (SELECT COALESCE(MAX(id), 1) FROM session_pretest_questions));

-- Semua mahasiswa CS301 mengerjakan pretest dan post-test sesi 1. Skor selaras
-- dengan session_assessment_results agar tampilan quiz dan N-Gain konsisten.
INSERT INTO session_pretest_submissions (id,session_id,student_id,assessment_type,answers,score,max_score,status,submitted_at) VALUES
(1,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),1,'pretest','[{"questionId":1,"answerIndex":0},{"questionId":2,"answerIndex":0},{"questionId":3,"answerIndex":1}]'::jsonb,63,100,'completed',CURRENT_TIMESTAMP),
(2,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),2,'pretest','[{"questionId":1,"answerIndex":0},{"questionId":2,"answerIndex":0},{"questionId":3,"answerIndex":1}]'::jsonb,64,100,'completed',CURRENT_TIMESTAMP),
(3,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),3,'pretest','[{"questionId":1,"answerIndex":0},{"questionId":2,"answerIndex":0},{"questionId":3,"answerIndex":1}]'::jsonb,65,100,'completed',CURRENT_TIMESTAMP),
(4,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),4,'pretest','[{"questionId":1,"answerIndex":0},{"questionId":2,"answerIndex":0},{"questionId":3,"answerIndex":1}]'::jsonb,66,100,'completed',CURRENT_TIMESTAMP),
(5,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),5,'pretest','[{"questionId":1,"answerIndex":0},{"questionId":2,"answerIndex":0},{"questionId":3,"answerIndex":1}]'::jsonb,67,100,'completed',CURRENT_TIMESTAMP),
(6,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),1,'posttest','[{"questionId":4,"answerIndex":0},{"questionId":5,"answerIndex":0},{"questionId":6,"answerIndex":0}]'::jsonb,77,100,'completed',CURRENT_TIMESTAMP),
(7,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),2,'posttest','[{"questionId":4,"answerIndex":0},{"questionId":5,"answerIndex":0},{"questionId":6,"answerIndex":0}]'::jsonb,78,100,'completed',CURRENT_TIMESTAMP),
(8,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),3,'posttest','[{"questionId":4,"answerIndex":0},{"questionId":5,"answerIndex":0},{"questionId":6,"answerIndex":0}]'::jsonb,79,100,'completed',CURRENT_TIMESTAMP),
(9,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),4,'posttest','[{"questionId":4,"answerIndex":0},{"questionId":5,"answerIndex":0},{"questionId":6,"answerIndex":0}]'::jsonb,80,100,'completed',CURRENT_TIMESTAMP),
(10,(SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),5,'posttest','[{"questionId":4,"answerIndex":0},{"questionId":5,"answerIndex":0},{"questionId":6,"answerIndex":0}]'::jsonb,81,100,'completed',CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET session_id=EXCLUDED.session_id,student_id=EXCLUDED.student_id,assessment_type=EXCLUDED.assessment_type,answers=EXCLUDED.answers,score=EXCLUDED.score,max_score=EXCLUDED.max_score,status=EXCLUDED.status,submitted_at=EXCLUDED.submitted_at,updated_at=now();
SELECT setval('session_pretest_submissions_id_seq', (SELECT COALESCE(MAX(id), 1) FROM session_pretest_submissions));

INSERT INTO attendance_sessions (id,role_scope,course_code,course_name,class_name,session_number,session_date,session_time,room,lab,topic,total_students,present,absent,sick,permit,excused,status,assistant_status,assistant_check_in_time) VALUES
(1,'lecturer','CS301','Algoritma & Struktur Data','CS301',1,'2026-01-08','08:00 - 10:00','Lab 301','','Pengenalan Algoritma',35,32,1,1,1,0,'Selesai','',''),
(2,'lecturer','CS302','Basis Data Lanjutan','CS302',1,'2026-01-08','10:00 - 13:00','Ruang 402','','Normalisasi Database',38,35,2,1,0,0,'Selesai','',''),
(3,'lecturer','TI-201','Pemrograman Web','TI-201',1,'2026-01-08','13:00 - 15:00','Lab Komputer 1','','REST API',40,30,9,0,1,0,'Selesai','',''),
(4,'assistant','CS301','Algoritma & Struktur Data','CS301',1,'2026-01-08','08:00 - 10:00','','Lab 301','Pengenalan Algoritma',35,32,1,1,1,0,'Selesai','Hadir','07:55'),
(5,'assistant','CS302','Basis Data Lanjutan','CS302',1,'2026-01-08','10:00 - 13:00','','Ruang 402','Normalisasi Database',38,35,2,1,0,0,'Selesai','Hadir','09:55'),
(6,'assistant','CS304','Jaringan Komputer','CS304',1,'2026-01-08','15:00 - 18:00','','Lab 304','Routing Dasar',32,30,1,0,1,0,'Selesai','Hadir','14:55'),
(7,'admin','CS301','Algoritma & Struktur Data','CS301',1,'2026-01-08','08:00 - 10:00','Lab 301','','Pengenalan Algoritma',35,33,1,0,0,1,'Selesai','Hadir','07:55')
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
(1,'CS301','Algoritma & Struktur Data','Dr. Ahmad Rahman','CS301','Lab 301','Senin & Kamis','08:00','10:00','Ganjil','2025/2026',35,12,14,'bg-blue-500'),
(2,'CS302','Basis Data Lanjutan','Prof. Siti Nurhaliza','CS302','Ruang 402','Selasa','10:00','13:00','Ganjil','2025/2026',38,13,14,'bg-blue-500'),
(3,'CS304','Jaringan Komputer','Dr. Rina Kusuma','CS304','Lab 304','Kamis','15:00','18:00','Ganjil','2025/2026',32,12,14,'bg-blue-500')
ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name,lecturer=EXCLUDED.lecturer,class_name=EXCLUDED.class_name,lab=EXCLUDED.lab,day=EXCLUDED.day,start_time=EXCLUDED.start_time,end_time=EXCLUDED.end_time,semester=EXCLUDED.semester,academic_year=EXCLUDED.academic_year,students=EXCLUDED.students,attendance_present=EXCLUDED.attendance_present,attendance_total=EXCLUDED.attendance_total,color=EXCLUDED.color;
SELECT setval('assistant_practical_courses_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assistant_practical_courses));

INSERT INTO assistant_tasks (id,task,submitted,total,deadline) VALUES
(1,'Verifikasi laporan praktikum Pemrograman Dasar',28,35,'2 hari lagi'),(2,'Input nilai praktikum Struktur Data',35,38,'3 hari lagi')
ON CONFLICT (id) DO UPDATE SET task=EXCLUDED.task,submitted=EXCLUDED.submitted,total=EXCLUDED.total,deadline=EXCLUDED.deadline;
SELECT setval('assistant_tasks_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assistant_tasks));

INSERT INTO assistant_reports (id,nim,name,course_code,course_name,class_name,week,topic,submitted_at,status,score,file_name,file_size) VALUES
(1,'TI2021001','Rama Dhani','CS301','Algoritma & Struktur Data','CS301',1,'Pengenalan Algoritma','2026-01-08 10:15','Menunggu Review',NULL,'Laporan Sesi 1 - Algoritma & Struktur Data.pdf','2.4 MB'),
(2,'TI2021002','Dina Amelia','CS302','Basis Data Lanjutan','CS302',1,'Normalisasi Database','2026-01-08 13:15','Disetujui',90,'Laporan Sesi 1 - Basis Data Lanjutan.pdf','1.8 MB'),
(3,'TI2021003','Fahmi Akbar','CS304','Jaringan Komputer','CS304',1,'Routing Dasar','2026-01-08 18:15','Ditolak',NULL,'Laporan Sesi 1 - Jaringan Komputer.pdf','2.1 MB')
ON CONFLICT (id) DO UPDATE SET nim=EXCLUDED.nim,name=EXCLUDED.name,course_code=EXCLUDED.course_code,course_name=EXCLUDED.course_name,class_name=EXCLUDED.class_name,week=EXCLUDED.week,topic=EXCLUDED.topic,submitted_at=EXCLUDED.submitted_at,status=EXCLUDED.status,score=EXCLUDED.score,file_name=EXCLUDED.file_name,file_size=EXCLUDED.file_size;
SELECT setval('assistant_reports_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assistant_reports));

INSERT INTO assistant_report_summary (id,course_code,course_name,class_name,total_reports,reviewed,pending,approved,needs_revision) VALUES
(1,'CS301','Algoritma & Struktur Data','CS301',1,0,1,0,0),
(2,'CS302','Basis Data Lanjutan','CS302',1,1,0,1,0),
(3,'CS304','Jaringan Komputer','CS304',1,1,0,0,1)
ON CONFLICT (id) DO UPDATE SET total_reports=EXCLUDED.total_reports,reviewed=EXCLUDED.reviewed,pending=EXCLUDED.pending,approved=EXCLUDED.approved,needs_revision=EXCLUDED.needs_revision;
SELECT setval('assistant_report_summary_id_seq', (SELECT COALESCE(MAX(id), 1) FROM assistant_report_summary));

INSERT INTO institution_settings (id,university_name,faculty_name,study_program_name,laboratory_name,campus_a_address,campus_b_address,website,email,phone,logo_path) VALUES
(1,'Universitas Muhammadiyah Jakarta','Fakultas Teknik','Teknik Informatika','Laboratorium Teknik Informatika','JL. K. H. Ahmad Dahlan Cirendeu Ciputat Tangerang Selatan','Jl. Cempaka Putih Tengah XXVII, Jakarta Pusat 10510','umj.ac.id','info@umj.ac.id','+6221-7492862/7401894','')
ON CONFLICT (id) DO UPDATE SET university_name=EXCLUDED.university_name,faculty_name=EXCLUDED.faculty_name,study_program_name=EXCLUDED.study_program_name,laboratory_name=EXCLUDED.laboratory_name,campus_a_address=EXCLUDED.campus_a_address,campus_b_address=EXCLUDED.campus_b_address,website=EXCLUDED.website,email=EXCLUDED.email,phone=EXCLUDED.phone,logo_path=EXCLUDED.logo_path,updated_at=CURRENT_TIMESTAMP;

INSERT INTO course_assessment_weights (course_id,attendance_weight,pretest_weight,assignment_weight,practicum_weight,posttest_weight,passing_grade)
SELECT id,10,15,20,20,35,55 FROM courses
ON CONFLICT (course_id) DO NOTHING;

INSERT INTO grade_scales (id,min_score,max_score,grade,is_passed) VALUES
(1,85,100,'A',TRUE),
(2,80,84.99,'A-',TRUE),
(3,75,79.99,'B+',TRUE),
(4,70,74.99,'B',TRUE),
(5,65,69.99,'B-',TRUE),
(6,60,64.99,'C+',TRUE),
(7,55,59.99,'C',TRUE),
(8,45,54.99,'D',FALSE),
(9,0,44.99,'E',FALSE)
ON CONFLICT (id) DO UPDATE SET min_score=EXCLUDED.min_score,max_score=EXCLUDED.max_score,grade=EXCLUDED.grade,is_passed=EXCLUDED.is_passed;

DELETE FROM report_signers
WHERE lower(name) IN ('popy meilina, m.kom','poppy melina, m.kom','sitti nurbaya ambo, mmsi');

INSERT INTO report_signers (id,role,name,identifier_type,identifier_number,signature_path,is_active)
SELECT 2,'head_of_laboratory',name,'ID',student_id,'',TRUE
FROM lab_assistants
WHERE role='kalab' AND status='Aktif'
ORDER BY id
LIMIT 1
ON CONFLICT (id) DO UPDATE SET role=EXCLUDED.role,name=EXCLUDED.name,identifier_type=EXCLUDED.identifier_type,identifier_number=EXCLUDED.identifier_number,signature_path=EXCLUDED.signature_path,is_active=EXCLUDED.is_active;

INSERT INTO student_activity_logs (student_id,course_id,session_id,activity_type,description,created_at)
SELECT st.id, c.id, cs.id, activity.activity_type, activity.description, CURRENT_TIMESTAMP - (cs.session_number || ' days')::interval
FROM course_sessions cs
JOIN courses c ON c.id=cs.course_id
JOIN classes cls ON cls.code=c.class_code
JOIN students st ON st.id=ANY(cls.students)
CROSS JOIN (VALUES
  ('Mengisi Presensi','Presensi sesi praktikum'),
  ('Mengerjakan Pretest','Submit pretest sesi'),
  ('Membuka Materi','Membuka materi praktikum'),
  ('Mengerjakan Post-test','Submit post-test sesi')
) AS activity(activity_type,description)
WHERE cs.session_number <= 2
ON CONFLICT DO NOTHING;

INSERT INTO admin_activities (id,action,detail,activity_time,icon,sort_order) VALUES
(1,'Mahasiswa baru terdaftar','Ahmad Fauzi - Teknik Informatika','5 menit lalu','user',1),
(2,'Tugas baru dibuat','Tugas Algoritma & Pemrograman','15 menit lalu','file',2),
(3,'Kursus diperbarui','Database Management - Materi Week 5','1 jam lalu','book',3),
(4,'Nilai diinput','UTS Pemrograman Web - 45 mahasiswa','2 jam lalu','award',4)
ON CONFLICT (id) DO UPDATE SET action=EXCLUDED.action,detail=EXCLUDED.detail,activity_time=EXCLUDED.activity_time,icon=EXCLUDED.icon,sort_order=EXCLUDED.sort_order;
SELECT setval('admin_activities_id_seq', (SELECT COALESCE(MAX(id), 1) FROM admin_activities));

-- <<< 004_normalized_page_data.up.sql

-- >>> 006_action_endpoints_schema.up.sql
ALTER TABLE courses ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS instructor VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS total_students INT NOT NULL DEFAULT 0;
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS submitted INT NOT NULL DEFAULT 0;
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS graded INT NOT NULL DEFAULT 0;
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS pending INT NOT NULL DEFAULT 0;
ALTER TABLE assignments ADD COLUMN IF NOT EXISTS assignment_type VARCHAR(50) NOT NULL DEFAULT 'Tugas';
UPDATE assignments a SET instructor = COALESCE(c.instructor, a.assistant), total_students = COALESCE(c.students, 0), submitted = CASE WHEN a.status IN ('submitted','graded') THEN COALESCE(c.students,0) ELSE 0 END, graded = CASE WHEN a.status='graded' THEN COALESCE(c.students,0) ELSE 0 END, pending = CASE WHEN a.status='pending' THEN COALESCE(c.students,0) ELSE 0 END, assignment_type = CASE WHEN lower(a.title) LIKE '%quiz%' THEN 'Quiz' WHEN lower(a.title) LIKE '%project%' THEN 'Project' ELSE 'Tugas' END FROM courses c WHERE c.name = a.course;

-- <<< 006_action_endpoints_schema.up.sql

-- >>> 007_admins_and_demo_login_users.up.sql
CREATE TABLE IF NOT EXISTS admins (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL DEFAULT 'password',
  default_password VARCHAR(255) NOT NULL DEFAULT 'password',
  is_password_changed BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO admins (name,email,password,default_password,is_password_changed) VALUES
('Administrator','admin@app.com','password','password',false)
ON CONFLICT (email) DO UPDATE SET
  name = EXCLUDED.name,
  password = EXCLUDED.password,
  default_password = EXCLUDED.default_password,
  is_password_changed = EXCLUDED.is_password_changed,
  updated_at = CURRENT_TIMESTAMP;

INSERT INTO students (name,email,password,default_password,student_id,program,semester,courses,status,join_date,is_password_changed) VALUES
('Budi Santoso','mahasiswa@app.com','password','password','DEMO-MHS-001','Teknik Informatika',6,ARRAY['Algoritma & Struktur Data','Basis Data Lanjutan','Pemrograman Web'],'Aktif',CURRENT_DATE,false)
ON CONFLICT (email) DO UPDATE SET
  name = EXCLUDED.name,
  password = EXCLUDED.password,
  default_password = EXCLUDED.default_password,
  student_id = EXCLUDED.student_id,
  program = EXCLUDED.program,
  semester = EXCLUDED.semester,
  courses = EXCLUDED.courses,
  status = EXCLUDED.status,
  is_password_changed = EXCLUDED.is_password_changed,
  updated_at = CURRENT_TIMESTAMP;

INSERT INTO lecturers (name,email,password,default_password,nidn,courses,is_password_changed) VALUES
('Prof. Dr. Ahmad Wijaya','dosen@app.com','password','password','DEMO-DOSEN-001',ARRAY['Algoritma & Struktur Data','Basis Data Lanjutan','Pemrograman Web'],false)
ON CONFLICT (email) DO UPDATE SET
  name = EXCLUDED.name,
  password = EXCLUDED.password,
  default_password = EXCLUDED.default_password,
  nidn = EXCLUDED.nidn,
  courses = EXCLUDED.courses,
  is_password_changed = EXCLUDED.is_password_changed,
  updated_at = CURRENT_TIMESTAMP;

INSERT INTO lab_assistants (name,email,phone,student_id,role,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date,password,default_password,is_password_changed) VALUES
('Andi Pratama','asslab@app.com','081200000001','DEMO-ASLAB-001','aslab','Laboratorium Pemrograman','Prof. Dr. Ahmad Wijaya',7,3.85,3,12,'Aktif',CURRENT_DATE,'password','password',false),
('Sari Laboran','laboran@app.com','081200000002','DEMO-LABORAN-001','laboran','Laboratorium Pemrograman','Kepala Lab Informatika',1,0,0,40,'Aktif',CURRENT_DATE,'password','password',false),
('Kepala Lab Informatika','kalab@app.com','081200000003','DEMO-KALAB-001','kalab','Kepala Laboratorium','-',1,0,0,40,'Aktif',CURRENT_DATE,'password','password',false)
ON CONFLICT (email) DO UPDATE SET
  name = EXCLUDED.name,
  phone = EXCLUDED.phone,
  student_id = EXCLUDED.student_id,
  role = EXCLUDED.role,
  lab = EXCLUDED.lab,
  supervisor = EXCLUDED.supervisor,
  semester = EXCLUDED.semester,
  gpa = EXCLUDED.gpa,
  assigned_courses = EXCLUDED.assigned_courses,
  weekly_hours = EXCLUDED.weekly_hours,
  status = EXCLUDED.status,
  password = EXCLUDED.password,
  default_password = EXCLUDED.default_password,
  is_password_changed = EXCLUDED.is_password_changed,
  updated_at = CURRENT_TIMESTAMP;

-- <<< 007_admins_and_demo_login_users.up.sql

COMMIT;

-- >>> 008_features.up.sql
ALTER TABLE course_sessions ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;
UPDATE course_sessions SET sort_order = session_number WHERE sort_order = 0;
ALTER TABLE assistant_reports ADD COLUMN IF NOT EXISTS rejection_note TEXT NOT NULL DEFAULT '';
ALTER TABLE course_materials ADD COLUMN IF NOT EXISTS file_url TEXT NOT NULL DEFAULT '';
ALTER TABLE session_assignments ADD COLUMN IF NOT EXISTS file_url TEXT NOT NULL DEFAULT '';
CREATE TABLE IF NOT EXISTS assistant_session_attendance (
  id SERIAL PRIMARY KEY,
  session_id INT NOT NULL REFERENCES course_sessions(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL DEFAULT 'Hadir',
  check_in_time VARCHAR(20) NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(session_id)
);
-- <<< 008_features.up.sql

-- >>> 009_sync_grades.up.sql
CREATE OR REPLACE FUNCTION sync_submission_grade_to_report()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE assistant_reports
  SET score = NEW.score, feedback = NEW.feedback
  WHERE nim = (SELECT student_id FROM students WHERE id = NEW.student_id LIMIT 1)
    AND week = (SELECT session_number FROM course_sessions WHERE id = (SELECT session_id FROM session_assignments WHERE id = NEW.assignment_id LIMIT 1) LIMIT 1)
    AND course_code = (SELECT class_code FROM courses WHERE id = (SELECT course_id FROM session_assignments WHERE id = NEW.assignment_id LIMIT 1) LIMIT 1);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS sync_submission_grade_trigger ON student_submissions;

CREATE TRIGGER sync_submission_grade_trigger
AFTER UPDATE ON student_submissions
FOR EACH ROW
WHEN (NEW.score IS DISTINCT FROM OLD.score OR NEW.feedback IS DISTINCT FROM OLD.feedback)
EXECUTE FUNCTION sync_submission_grade_to_report();
-- <<< 009_sync_grades.up.sql

-- >>> 010_demo_ready_data.up.sql
-- Melengkapi data agar seluruh layar demo terisi: pengumpulan tugas, presensi
-- asisten, dan status alur laporan. Semua idempotent lewat ON CONFLICT.

-- Pengumpulan tugas mahasiswa untuk tugas yang berstatus submitted.
-- Sebagian sudah dinilai, sebagian belum, supaya aslab punya bahan penilaian.
INSERT INTO student_submissions (assignment_id,student_id,answer_text,file_url,file_name,file_size,submitted_at,score,feedback) VALUES
(1,1,'Implementasi bubble sort dengan optimasi flag swap.','','bubble_sort_ahmad.pdf','1.2 MB','2026-01-18 09:20',88,'Bagus, sudah ada optimasi.'),
(1,2,'Implementasi bubble sort lengkap dengan analisis kompleksitas O(n^2).','','bubble_sort_siti.pdf','1.4 MB','2026-01-18 10:05',95,'Sangat lengkap.'),
(1,3,'Implementasi bubble sort dasar.','','bubble_sort_budi.pdf','0.9 MB','2026-01-19 08:40',NULL,''),
(2,1,'Analisis quick sort partisi Lomuto.','','quick_sort_ahmad.pdf','1.1 MB','2026-01-25 14:10',85,'Rapi.'),
(2,2,'Analisis quick sort partisi Hoare dan Lomuto beserta perbandingan.','','quick_sort_siti.pdf','1.5 MB','2026-01-25 15:00',88,'Perbandingannya bagus.'),
(5,2,'Jawaban quiz pengenalan algoritma.','','quiz_siti.pdf','0.6 MB','2026-01-17 11:00',80,'Tuntas.'),
(5,1,'Jawaban quiz pengenalan algoritma.','','quiz_ahmad.pdf','0.5 MB','2026-01-17 11:30',NULL,'')
ON CONFLICT (assignment_id,student_id) DO UPDATE SET answer_text=EXCLUDED.answer_text,file_name=EXCLUDED.file_name,file_size=EXCLUDED.file_size,submitted_at=EXCLUDED.submitted_at,score=EXCLUDED.score,feedback=EXCLUDED.feedback;
SELECT setval('student_submissions_id_seq', (SELECT COALESCE(MAX(id), 1) FROM student_submissions));

-- Presensi asisten lab per sesi praktikum (kursus 1).
INSERT INTO assistant_session_attendance (session_id,status,check_in_time) VALUES
((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1),'Hadir','07:55'),
((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=2),'Hadir','07:58'),
((SELECT id FROM course_sessions WHERE course_id=1 AND session_number=3),'Hadir','08:02')
ON CONFLICT (session_id) DO UPDATE SET status=EXCLUDED.status, check_in_time=EXCLUDED.check_in_time, updated_at=now();

-- Status alur laporan per kursus (DRAFT/SUBMITTED/APPROVED/REJECTED).
INSERT INTO report_workflows (course_id,status,updated_at) VALUES
(1,'SUBMITTED',now()),
(2,'APPROVED',now()),
(4,'REJECTED',now())
ON CONFLICT (course_id) DO UPDATE SET status=EXCLUDED.status, updated_at=now();
-- <<< 010_demo_ready_data.up.sql

-- >>> 011_demo_account_linkage.up.sql
-- Menyambungkan akun demo (mahasiswa/dosen/aslab@app.com) ke kelas CS301 yang
-- datanya paling lengkap, supaya login demo langsung melihat data terisi.

-- CS301 diampu dosen dan aslab demo agar dashboard keduanya berisi.
UPDATE courses
SET instructor = COALESCE((SELECT name FROM lecturers WHERE email='dosen@app.com'), instructor),
    assistant  = COALESCE((SELECT name FROM lab_assistants WHERE email='asslab@app.com'), assistant)
WHERE class_code = 'CS301';

-- Mahasiswa demo dimasukkan sebagai anggota kelas CS301.
UPDATE classes
SET students = ARRAY(SELECT DISTINCT unnest(students || ARRAY[(SELECT id FROM students WHERE email='mahasiswa@app.com')]))
WHERE code = 'CS301' AND (SELECT id FROM students WHERE email='mahasiswa@app.com') IS NOT NULL;

-- Nilai pretest dan post-test mahasiswa demo pada sesi 1 CS301.
INSERT INTO session_assessment_results (session_id,student_id,assessment_type,score,max_score,status,submitted_at,updated_at)
SELECT s.id, m.id, t.atype, t.score, 100, 'completed', now(), now()
FROM (SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1) s
CROSS JOIN (SELECT id FROM students WHERE email='mahasiswa@app.com') m
CROSS JOIN (VALUES ('pretest',68),('posttest',85)) AS t(atype,score)
ON CONFLICT (session_id,student_id,assessment_type) DO UPDATE SET score=EXCLUDED.score, status='completed', updated_at=now();

INSERT INTO session_pretest_submissions (session_id,student_id,assessment_type,answers,score,max_score,status,submitted_at)
SELECT s.id, m.id, t.atype, '[]'::jsonb, t.score, 100, 'completed', now()
FROM (SELECT id FROM course_sessions WHERE course_id=1 AND session_number=1) s
CROSS JOIN (SELECT id FROM students WHERE email='mahasiswa@app.com') m
CROSS JOIN (VALUES ('pretest',68),('posttest',85)) AS t(atype,score)
ON CONFLICT (session_id,student_id,assessment_type) DO UPDATE SET score=EXCLUDED.score, status='completed', updated_at=now();

-- Pengumpulan tugas mahasiswa demo: satu sudah dinilai, satu belum.
INSERT INTO student_submissions (assignment_id,student_id,answer_text,file_name,file_size,submitted_at,score,feedback)
SELECT a.aid, m.id, 'Jawaban tugas demo mahasiswa.', 'tugas_demo.pdf','1.0 MB','2026-01-18 09:00', a.sc, a.fb
FROM (SELECT id FROM students WHERE email='mahasiswa@app.com') m
CROSS JOIN (VALUES (1,90::int,'Bagus, lengkap.'),(2,NULL::int,'')) AS a(aid,sc,fb)
WHERE EXISTS (SELECT 1 FROM session_assignments sa WHERE sa.id=a.aid)
ON CONFLICT (assignment_id,student_id) DO NOTHING;

-- Presensi mahasiswa demo pada sesi asisten CS301 agar presensi & progress terisi.
INSERT INTO attendance_records (session_id,nim,name,status,attendance_time,check_in_time)
SELECT (SELECT id FROM attendance_sessions WHERE role_scope='assistant' AND course_code='CS301' ORDER BY id LIMIT 1),
       m.student_id, m.name, 'Hadir', '08:00', '08:00'
FROM (SELECT student_id, name FROM students WHERE email='mahasiswa@app.com') m
WHERE (SELECT id FROM attendance_sessions WHERE role_scope='assistant' AND course_code='CS301' ORDER BY id LIMIT 1) IS NOT NULL
ON CONFLICT DO NOTHING;
-- <<< 011_demo_account_linkage.up.sql

-- >>> 012_blank_user_accounts.up.sql
-- Akun kosong (tanpa data) untuk diisi sendiri oleh pengguna saat demo.
-- Semua kata sandinya "password". Admin tidak dibuat di sini (sudah ada akun admin).

INSERT INTO students (name,email,password,default_password,student_id,program,semester,courses,status,join_date,is_password_changed)
VALUES ('Akun Mahasiswa (Kosong)','akun-mahasiswa@app.com','password','password','USER-MHS-01','Teknik Informatika',1,ARRAY[]::text[],'Aktif',CURRENT_DATE,FALSE)
ON CONFLICT (email) DO NOTHING;

INSERT INTO lab_assistants (name,email,phone,student_id,role,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date,password,default_password,is_password_changed) VALUES
('Akun Aslab (Kosong)','akun-aslab@app.com','','USER-ASLAB-01','aslab','','',1,0,0,0,'Aktif',CURRENT_DATE,'password','password',FALSE),
('Akun Laboran (Kosong)','akun-laboran@app.com','','USER-LABORAN-01','laboran','','',1,0,0,0,'Aktif',CURRENT_DATE,'password','password',FALSE),
('Akun Kalab (Kosong)','akun-kalab@app.com','','USER-KALAB-01','kalab','','',1,0,0,0,'Aktif',CURRENT_DATE,'password','password',FALSE)
ON CONFLICT (email) DO NOTHING;
-- <<< 012_blank_user_accounts.up.sql

COMMIT;
