CREATE ROLE user1 LOGIN SUPERUSER PASSWORD 'password';
CREATE DATABASE test_postgres_db;
GRANT all privileges ON DATABASE test_postgres_db TO user1;