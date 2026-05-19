-- Keep database schema, wipe all data, seed only 4 demo login accounts.
-- Run via: go run main.go init demo

BEGIN;

DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN (
    SELECT tablename
    FROM pg_tables
    WHERE schemaname = 'public'
  ) LOOP
    EXECUTE 'TRUNCATE TABLE public.' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE';
  END LOOP;
END $$;

INSERT INTO students (
  name,
  email,
  password,
  default_password,
  student_id,
  program,
  semester,
  courses,
  status,
  join_date,
  is_password_changed
) VALUES (
  'Demo Mahasiswa',
  'mahasiswa@app.com',
  'password',
  'password',
  'DEMO-STU-001',
  'Teknik Informatika',
  1,
  ARRAY[]::text[],
  'Aktif',
  CURRENT_DATE,
  FALSE
);

INSERT INTO lecturers (
  name,
  email,
  password,
  default_password,
  nidn,
  courses,
  is_password_changed
) VALUES (
  'Demo Dosen',
  'dosen@app.com',
  'password',
  'password',
  '9999999001',
  ARRAY[]::text[],
  FALSE
);

INSERT INTO lab_assistants (
  name,
  email,
  phone,
  student_id,
  lab,
  supervisor,
  semester,
  gpa,
  assigned_courses,
  weekly_hours,
  status,
  join_date,
  password,
  default_password,
  is_password_changed
) VALUES (
  'Demo Aslab',
  'asslab@app.com',
  '080000000000',
  'DEMO-ASL-001',
  'Laboratorium Pemrograman',
  'Demo Dosen',
  7,
  3.50,
  0,
  0,
  'Aktif',
  CURRENT_DATE,
  'password',
  'password',
  FALSE
);

INSERT INTO admins (
  name,
  email,
  password,
  default_password,
  is_password_changed
) VALUES (
  'Demo Admin',
  'admin@app.com',
  'password',
  'password',
  FALSE
);

COMMIT;
