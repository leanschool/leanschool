-- Create test data for the timetable_plans table
INSERT INTO timetable_plans (id, school_year_id) VALUES
('plan1', '2023-2024'),
('plan2', '2024-2025');

-- Create test data for the time_slot_definitions table
INSERT INTO time_slot_definitions (id, plan_id, start_time, end_time, day_of_week) VALUES
('slot1', 'plan1', '08:00:00', '09:00:00', 'Monday'),
('slot2', 'plan1', '09:00:00', '10:00:00', 'Monday'),
('slot3', 'plan2', '08:00:00', '09:00:00', 'Tuesday');

-- Create test data for the grade_requirements table
INSERT INTO grade_requirements (id, plan_id, grade_level, subject_id, hours_per_week) VALUES
('req1', 'plan1', '10', 'math', 5),
('req2', 'plan1', '10', 'physics', 3),
('req3', 'plan2', '11', 'chemistry', 4);

-- Create test data for the class_constraints table
INSERT INTO class_constraints (id, plan_id, constraint_type, value) VALUES
('con1', 'plan1', 'room', 'room101'),
('con2', 'plan1', 'teacher', 'teacher1'),
('con3', 'plan2', 'room', 'room102');

-- Create test data for the timetable_entries table
INSERT INTO timetable_entries (id, plan_id, time_slot_id, subject_id, teacher_id, room_id, class_id) VALUES
('entry1', 'plan1', 'slot1', 'math', 'teacher1', 'room101', 'class10A'),
('entry2', 'plan1', 'slot2', 'physics', 'teacher2', 'room102', 'class10B'),
('entry3', 'plan2', 'slot3', 'chemistry', 'teacher3', 'room103', 'class11A');

-- Create test data for the conflicts table
INSERT INTO conflicts (id, plan_id, conflict_type, description) VALUES
('conf1', 'plan1', 'room', 'Room conflict'),
('conf2', 'plan1', 'teacher', 'Teacher conflict'),
('conf3', 'plan2', 'room', 'Room conflict');

-- Create test data for the conflict_entries table
INSERT INTO conflict_entries (conflict_id, entry_id) VALUES
('conf1', 'entry1'),
('conf2', 'entry2'),
('conf3', 'entry3');

-- -- Create test data for the user_registry table
-- INSERT INTO user_registry (id, user_sub, email, first_name, last_name) VALUES
-- ('user1', 'sub1', 'user1@example.com', 'John', 'Doe'),
-- ('user2', 'sub2', 'user2@example.com', 'Jane', 'Smith'),
-- ('user3', 'sub3', 'user3@example.com', 'Alice', 'Johnson');

-- -- Create test data for the registration_workflow table
-- INSERT INTO registration_workflow (id, user_id, status, created_at, updated_at) VALUES
-- ('reg1', 'user1', 'pending', '2023-01-01 10:00:00', '2023-01-01 10:00:00'),
-- ('reg2', 'user2', 'approved', '2023-01-02 11:00:00', '2023-01-03 12:00:00'),
-- ('reg3', 'user3', 'rejected', '2023-01-03 12:00:00', '2023-01-04 13:00:00');

-- -- Create test data for the role_mappings table
-- INSERT INTO role_mappings (keycloak_role, local_role) VALUES
-- ('admin', 'admin'),
-- ('teacher', 'teacher'),
-- ('student', 'student');

-- Create test data for the locations table
INSERT INTO locations (id, lon, lat) VALUES
('loc1', 12.34567890, 45.67890123),
('loc2', 23.45678901, 56.78901234),
('loc3', 34.56789012, 67.89012345);

-- Create test data for the buildings table
INSERT INTO buildings (id, name, location_id) VALUES
('build1', 'Main Building', 'loc1'),
('build2', 'Science Building', 'loc2'),
('build3', 'Art Building', 'loc3');

-- Create test data for the rooms table
INSERT INTO rooms (id, name, building_id) VALUES
('room1', 'Room 101', 'build1'),
('room2', 'Room 102', 'build2'),
('room3', 'Room 103', 'build3');

-- Create test data for the postal_codes table
INSERT INTO postal_codes (number, city) VALUES
(12345, 'New York'),
(23456, 'Los Angeles'),
(34567, 'Chicago');

-- Create test data for the cities table
INSERT INTO cities (id, name) VALUES
('city1', 'New York'),
('city2', 'Los Angeles'),
('city3', 'Chicago');

-- Create test data for the city_postal_codes table
INSERT INTO city_postal_codes (city_id, postal_code_no) VALUES
('city1', 12345),
('city2', 23456),
('city3', 34567);

-- Create test data for the addresses table
INSERT INTO addresses (id, street, postal_code_no, city_id) VALUES
('addr1', '123 Main St', 12345, 'city1'),
('addr2', '456 Elm St', 23456, 'city2'),
('addr3', '789 Oak St', 34567, 'city3');

-- Create test data for the school_years table
INSERT INTO school_years (id, from_dt, to_dt) VALUES
('year1', '2023-09-01', '2024-06-30'),
('year2', '2024-09-01', '2025-06-30'),
('year3', '2025-09-01', '2026-06-30');

-- Create test data for the curricula table
INSERT INTO curricula (id, name, school_year_id) VALUES
('curr1', 'General Curriculum', 'year1'),
('curr2', 'Science Curriculum', 'year2'),
('curr3', 'Art Curriculum', 'year3');

-- Create test data for the subjects table
INSERT INTO subjects (id, name, curriculum_id) VALUES
('subj1', 'Mathematics', 'curr1'),
('subj2', 'Physics', 'curr2'),
('subj3', 'Art', 'curr3');

-- Create test data for the accounts table
INSERT INTO accounts (id, name, valid_from, valid_to) VALUES
('acc1', 'Account 1', '2023-01-01', '2023-12-31'),
('acc2', 'Account 2', '2024-01-01', '2024-12-31'),
('acc3', 'Account 3', '2025-01-01', '2025-12-31');

-- Create test data for the persons table
INSERT INTO persons (id, name, sub) VALUES
('person1', 'John Doe', 'sub1'),
('person2', 'Jane Smith', 'sub2'),
('person3', 'Alice Johnson', 'sub3');

-- Create test data for the receipts table
INSERT INTO receipts (id, receipt_owner_id, issue_date, total_amount) VALUES
('receipt1', 'person1', '2023-01-01', 100.00),
('receipt2', 'person2', '2023-02-01', 150.00),
('receipt3', 'person3', '2023-03-01', 200.00);

-- Create test data for the receipt_items table
INSERT INTO receipt_items (id, receipt_id, description, amount) VALUES
('item1', 'receipt1', 'Item 1', 50.00),
('item2', 'receipt1', 'Item 2', 50.00),
('item3', 'receipt2', 'Item 3', 75.00),
('item4', 'receipt2', 'Item 4', 75.00),
('item5', 'receipt3', 'Item 5', 100.00),
('item6', 'receipt3', 'Item 6', 100.00);