package database

import (
	"log"
)

func RunMigrations() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL CHECK (role IN ('student','admin','lecturer','assistant')),
			nidn VARCHAR(50) DEFAULT '',
			nim VARCHAR(50) DEFAULT '',
			study_program VARCHAR(255) DEFAULT '',
			semester INT DEFAULT 1,
			phone VARCHAR(50) DEFAULT '',
			status VARCHAR(20) DEFAULT 'Aktif',
			is_password_changed BOOLEAN DEFAULT false,
			join_date DATE DEFAULT CURRENT_DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS academic_years (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			semester VARCHAR(20) NOT NULL,
			status VARCHAR(20) DEFAULT 'Mendatang',
			total_courses INT DEFAULT 0,
			total_students INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS courses (
			id SERIAL PRIMARY KEY,
			code VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			credits INT DEFAULT 3,
			study_program VARCHAR(255) DEFAULT '',
			academic_year_id INT REFERENCES academic_years(id),
			class_code VARCHAR(50) DEFAULT '',
			lecturer_id INT REFERENCES users(id),
			assistant_id INT REFERENCES users(id),
			day VARCHAR(20) DEFAULT '',
			start_time VARCHAR(10) DEFAULT '',
			end_time VARCHAR(10) DEFAULT '',
			room VARCHAR(100) DEFAULT '',
			status VARCHAR(20) DEFAULT 'Aktif',
			sessions INT DEFAULT 14,
			color VARCHAR(50) DEFAULT 'bg-blue-500',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS course_enrollments (
			id SERIAL PRIMARY KEY,
			student_id INT REFERENCES users(id),
			course_id INT REFERENCES courses(id),
			academic_year_id INT REFERENCES academic_years(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id SERIAL PRIMARY KEY,
			course_id INT REFERENCES courses(id),
			session_number INT NOT NULL,
			topic VARCHAR(255) DEFAULT '',
			date DATE,
			start_time VARCHAR(10) DEFAULT '',
			end_time VARCHAR(10) DEFAULT '',
			type VARCHAR(20) DEFAULT 'offline' CHECK (type IN ('online','offline')),
			conference_link TEXT DEFAULT '',
			room VARCHAR(100) DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS materials (
			id SERIAL PRIMARY KEY,
			course_id INT REFERENCES courses(id),
			session_id INT REFERENCES sessions(id),
			title VARCHAR(255) NOT NULL,
			type VARCHAR(20) NOT NULL DEFAULT 'pdf',
			file_url TEXT DEFAULT '',
			file_size VARCHAR(20) DEFAULT '',
			duration VARCHAR(50) DEFAULT '',
			downloads INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS assignments (
			id SERIAL PRIMARY KEY,
			course_id INT REFERENCES courses(id),
			session_id INT REFERENCES sessions(id),
			title VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			deadline_date DATE,
			deadline_time VARCHAR(10) DEFAULT '23:59',
			assistant_id INT REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS assignment_submissions (
			id SERIAL PRIMARY KEY,
			assignment_id INT REFERENCES assignments(id),
			student_id INT REFERENCES users(id),
			answer_text TEXT DEFAULT '',
			file_url TEXT DEFAULT '',
			file_name VARCHAR(255) DEFAULT '',
			file_size VARCHAR(20) DEFAULT '',
			submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status VARCHAR(20) DEFAULT 'submitted',
			score INT,
			feedback TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS attendance (
			id SERIAL PRIMARY KEY,
			session_id INT REFERENCES sessions(id),
			course_id INT REFERENCES courses(id),
			student_id INT REFERENCES users(id),
			status VARCHAR(20) NOT NULL DEFAULT 'Hadir',
			check_in_time VARCHAR(10) DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS grades (
			id SERIAL PRIMARY KEY,
			course_id INT REFERENCES courses(id),
			student_id INT REFERENCES users(id),
			academic_year_id INT REFERENCES academic_years(id),
			tugas1 DECIMAL(5,2) DEFAULT 0,
			tugas2 DECIMAL(5,2) DEFAULT 0,
			tugas3 DECIMAL(5,2) DEFAULT 0,
			ujian_akhir DECIMAL(5,2) DEFAULT 0,
			nilai_akhir DECIMAL(5,2) DEFAULT 0,
			grade VARCHAR(5) DEFAULT '',
			status VARCHAR(20) DEFAULT 'Tidak Lulus',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS lab_reports (
			id SERIAL PRIMARY KEY,
			course_id INT REFERENCES courses(id),
			student_id INT REFERENCES users(id),
			session_id INT REFERENCES sessions(id),
			week INT DEFAULT 1,
			topic VARCHAR(255) DEFAULT '',
			file_name VARCHAR(255) DEFAULT '',
			file_size VARCHAR(20) DEFAULT '',
			file_url TEXT DEFAULT '',
			submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status VARCHAR(50) DEFAULT 'Menunggu Review',
			score INT,
			feedback TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS activities (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id),
			action VARCHAR(255) NOT NULL,
			detail TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS lab_assistants (
			id SERIAL PRIMARY KEY,
			user_id INT UNIQUE REFERENCES users(id),
			phone VARCHAR(50) DEFAULT '',
			lab VARCHAR(255) DEFAULT '',
			supervisor_id INT REFERENCES users(id),
			weekly_hours INT DEFAULT 12,
			assigned_courses INT DEFAULT 0,
			gpa DECIMAL(3,2) DEFAULT 0,
			join_date DATE DEFAULT CURRENT_DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatal("Migration error:", err)
		}
	}

	log.Println("Database migrations completed")
}
