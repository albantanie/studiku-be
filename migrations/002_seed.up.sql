INSERT INTO academic_years (id,name,start_date,end_date,semester,status,total_courses,total_students) VALUES
(1,'2024/2025 Genap','2025-01-15','2025-06-30','Genap','Aktif',48,1234),
(2,'2024/2025 Ganjil','2024-08-15','2024-12-31','Ganjil','Selesai',45,1189),
(3,'2025/2026 Ganjil','2025-08-15','2025-12-31','Ganjil','Mendatang',0,0),
(4,'2023/2024 Genap','2024-01-15','2024-06-30','Genap','Selesai',42,1098)
ON CONFLICT (id) DO NOTHING;
SELECT setval('academic_years_id_seq', (SELECT COALESCE(MAX(id), 1) FROM academic_years));

INSERT INTO courses (id,name,instructor,assistant,study_program,academic_year,class_code,status,day,start_time,end_time,room,sessions,credits,students,attendance_present,attendance_total,color) VALUES
(1,'Algoritma & Struktur Data','Dr. Ahmad Rahman','Andi Prasetyo, S.Kom','Teknik Informatika','2025/2026','CS301','Aktif','Senin & Kamis','08:00','10:00','Lab 301',14,4,35,10,12,'bg-blue-500'),
(2,'Basis Data Lanjutan','Prof. Siti Nurhaliza','Budi Santoso, M.Kom','Teknik Informatika','2025/2026','CS302','Aktif','Selasa','10:00','13:00','Ruang 402',14,3,142,12,12,'bg-blue-500'),
(3,'Pemrograman Web','Ir. Budi Hartono','Dewi Lestari, S.Kom','Teknik Informatika','Genap 2024/2025','TI-201','Aktif','Rabu & Jumat','13:00','15:00','Lab Komputer 1',14,3,156,11,12,'bg-blue-500'),
(4,'Jaringan Komputer','Dr. Rina Kusuma','Eko Wijaya, M.T','Teknik Informatika','Genap 2024/2025','CS304','Aktif','Kamis','15:00','18:00','Lab 304',14,3,138,9,12,'bg-blue-500'),
(5,'Rekayasa Perangkat Lunak','Prof. Agus Setiawan','Fina Marlina, S.Kom','Teknik Informatika','Genap 2024/2025','TI-305','Aktif','Selasa','08:00','10:00','Ruang 402',12,3,124,12,12,'bg-blue-500')
ON CONFLICT (id) DO NOTHING;
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

INSERT INTO lab_assistants (id,name,email,phone,student_id,lab,supervisor,semester,gpa,assigned_courses,weekly_hours,status,join_date,password,default_password,is_password_changed) VALUES
(1,'Rama Dhani','rama.dhani@student.ac.id','081234567890','TI2021001','Laboratorium Pemrograman','Dr. Budi Santoso',7,3.85,2,12,'Aktif','2024-09-01','password','password',false),
(2,'Dina Amelia','dina.amelia@student.ac.id','082345678901','TI2021002','Laboratorium Database','Prof. Siti Aminah',7,3.92,1,8,'Aktif','2024-09-01','password','password',false),
(3,'Fahmi Akbar','fahmi.akbar@student.ac.id','083456789012','TI2021003','Laboratorium Jaringan','Dr. Ahmad Wijaya',6,3.78,2,10,'Aktif','2024-09-01','password','password',false)
ON CONFLICT (id) DO NOTHING;
SELECT setval('lab_assistants_id_seq', (SELECT COALESCE(MAX(id), 1) FROM lab_assistants));

INSERT INTO classes (id,code,name,academic_year,assistant,schedule,room,total_students,capacity,students) VALUES
(1,'A1','Pemrograman Dasar','2024/2025','Ahmad Fauzi','Senin, 08:00 - 10:00','Lab 301',35,40,ARRAY[1,2,3,4,5]),
(2,'A2','Struktur Data','2024/2025','Siti Nurhaliza','Selasa, 10:00 - 12:00','Lab 302',38,40,ARRAY[6,7]),
(3,'B1','Basis Data','2024/2025','Budi Setiawan','Rabu, 13:00 - 15:00','Lab 303',32,40,ARRAY[]::int[]),
(4,'B2','Pemrograman Web','2024/2025','Dewi Kartika','Kamis, 08:00 - 10:00','Lab 304',40,40,ARRAY[]::int[]),
(5,'C1','Kecerdasan Buatan','2024/2025','Eko Prasetyo','Jumat, 10:00 - 12:00','Lab 305',28,35,ARRAY[]::int[])
ON CONFLICT (id) DO NOTHING;
SELECT setval('classes_id_seq', (SELECT COALESCE(MAX(id), 1) FROM classes));
