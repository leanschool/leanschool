-- Create the timetable database if it doesn't exist.
-- This script is mounted into the postgres container's docker-entrypoint-initdb.d.
SELECT 'CREATE DATABASE timetable'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'timetable')\gexec
