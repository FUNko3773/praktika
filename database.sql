CREATE DATABASE IF NOT EXISTS hospital;
USE hospital;

CREATE TABLE IF NOT EXISTS Doctors (
    id INT PRIMARY KEY AUTO_INCREMENT,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL UNIQUE,
    address VARCHAR(255) NOT NULL,
    graduation_year INT NOT NULL
);

CREATE TABLE IF NOT EXISTS Qualifications (
    doctor_id INT PRIMARY KEY,
    position VARCHAR(255) NOT NULL,
    FOREIGN KEY (doctor_id) REFERENCES Doctors(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Patients (
    id INT PRIMARY KEY AUTO_INCREMENT,
    full_name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    pensioner_discount DECIMAL(5,2) DEFAULT 0.00
);

CREATE TABLE IF NOT EXISTS MedicalServices (
    id INT PRIMARY KEY AUTO_INCREMENT,
    service_name VARCHAR(255) NOT NULL,
    cost INT NOT NULL
);

CREATE TABLE IF NOT EXISTS Appeals (
    id INT PRIMARY KEY AUTO_INCREMENT,
    patient_id INT NOT NULL,
    diagnosis VARCHAR(255) NOT NULL,
    appeal_date DATE NOT NULL,
    FOREIGN KEY (patient_id) REFERENCES Patients(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Appointments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    appeal_id INT NOT NULL,
    doctor_id INT NOT NULL,
    diagnosis VARCHAR(255) NOT NULL,
    appointment_date DATE NOT NULL,
    FOREIGN KEY (appeal_id) REFERENCES Appeals(id) ON DELETE CASCADE,
    FOREIGN KEY (doctor_id) REFERENCES Doctors(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS AppointmentServices (
    appointment_id INT,
    service_id INT,
    PRIMARY KEY (appointment_id, service_id),
    FOREIGN KEY (appointment_id) REFERENCES Appointments(id) ON DELETE CASCADE,
    FOREIGN KEY (service_id) REFERENCES MedicalServices(id) ON DELETE CASCADE
);

INSERT INTO Doctors (full_name, phone, address, graduation_year) VALUES
('Иванов Иван Иванович', '+7-999-111-22-33', 'ул. Ленина, д. 1, кв. 5', 2010),
('Петрова Анна Сергеевна', '+7-999-222-33-44', 'ул. Пушкина, д. 10, кв. 15', 2015),
('Соколова Ирина Дмитриевна', '+7-999-111-11-11', 'ул. Ленина, д. 25, кв. 4', 2015),
('Волков Сергей Александрович', '+7-999-222-22-22', 'ул. Мира, д. 12, кв. 7', 2013),
('Морозова Наталья Петровна', '+7-999-333-33-33', 'ул. Пушкина, д. 8, кв. 15', 2016),
('Новиков Алексей Иванович', '+7-999-444-44-44', 'ул. Советская, д. 3, кв. 9', 2019),
('Федорова Ольга Сергеевна', '+7-999-555-55-55', 'ул. Лесная, д. 7, кв. 12', 2017),
('Попов Дмитрий Владимирович', '+7-999-666-66-66', 'ул. Речная, д. 5, кв. 6', 2014),
('Лебедева Анна Михайловна', '+7-999-777-77-77', 'ул. Парковая, д. 10, кв. 3', 2021),
('Белова Екатерина Андреевна', '+7-999-888-88-88', 'ул. Садовая, д. 2, кв. 8', 2018),
('Громов Игорь Петрович', '+7-999-999-99-99', 'ул. Северная, д. 15, кв. 1', 2011),
('Крылова Татьяна Сергеевна', '+7-999-000-00-00', 'ул. Южная, д. 9, кв. 5', 2020);

INSERT INTO Qualifications (doctor_id, position) VALUES
(1, 'Терапевт'),
(2, 'Хирург'),
(3, 'Невролог'),
(4, 'Хирург'),
(5, 'Терапевт'),
(6, 'Дерматолог'),
(7, 'Эндокринолог'),
(8, 'Ортопед'),
(9, 'Гинеколог'),
(10, 'Уролог'),
(11, 'Отоларинголог'),
(12, 'Онколог');

INSERT INTO Patients (full_name, address, pensioner_discount) VALUES
('Козлова Мария Ивановна', 'ул. Мира, д. 8, кв. 3', 0.10),
('Смирнов Александр Петрович', 'ул. Советская, д. 15, кв. 7', 0.00),
('Васильева Елена Дмитриевна', 'ул. Победы, д. 3, кв. 12', 0.10),
('Иванов Петр Сергеевич', 'ул. Лесная, д. 5, кв. 2', 0.00),
('Сидорова Анна Ивановна', 'ул. Садовая, д. 10, кв. 5', 0.10),
('Алексеева Ольга Владимировна', 'ул. Парковая, д. 5, кв. 8', 0.10),
('Денисов Виктор Павлович', 'ул. Центральная, д. 12, кв. 3', 0.00),
('Тихомирова Татьяна Николаевна', 'ул. Молодежная, д. 7, кв. 15', 0.10),
('Новиков Илья Михайлович', 'ул. Новая, д. 1, кв. 1', 0.00),
('Морозова Екатерина Сергеевна', 'ул. Весенняя, д. 8, кв. 4', 0.10);

INSERT INTO MedicalServices (service_name, cost) VALUES
('Приём терапевта', 1500),
('УЗИ брюшной полости', 2500),
('Анализ крови', 800),
('Консультация хирурга', 2000),
('Проверка зрения', 1200),
('МРТ головного мозга', 5000),
('ЭКГ', 1200),
('Рентген грудной клетки', 1800),
('Анализ мочи', 600),
('Приём кардиолога', 2000),
('Массаж', 1500),
('УЗИ сердца', 3500),
('Хирургическая операция', 15000),
('Стоматологический осмотр', 1000),
('Неврологическая консультация', 1800);

INSERT INTO Appeals (patient_id, diagnosis, appeal_date) VALUES
(1, 'Грипп', '2026-01-15'),
(2, 'Боль в животе', '2026-01-20'),
(3, 'Ухудшение зрения', '2026-01-25');

INSERT INTO Appointments (appeal_id, doctor_id, diagnosis, appointment_date) VALUES
(1, 1, 'Грипп - повторный приём', '2026-02-10'),
(1, 2, 'Консультация хирурга по поводу гриппа', '2026-02-15'),
(2, 2, 'Послеоперационный осмотр', '2026-02-01'),
(2, 1, 'Контрольный осмотр терапевта', '2026-02-05'),
(3, 3, 'Проверка зрения - контроль', '2026-02-20'),
(3, 1, 'Общий осмотр', '2026-02-25'),
(1, 3, 'Консультация окулиста', '2026-03-01'),
(2, 3, 'Проверка зрения перед операцией', '2026-03-05'),
(3, 2, 'Консультация хирурга', '2026-03-10'),
(1, 1, 'Повторный приём терапевта', '2026-03-15');

INSERT INTO AppointmentServices (appointment_id, service_id) VALUES
(1, 1),
(1, 3),
(2, 4),
(2, 2),
(3, 5),
(3, 1),
(4, 3),
(4, 1),
(5, 6),
(5, 5),
(6, 1),
(7, 5),
(8, 2),
(9, 4),
(10, 1);