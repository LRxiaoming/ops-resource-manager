-- Create admin user with password 'admin123'
INSERT INTO users (username, password, role, created_at)
VALUES ('admin', '$2y$10$K8G8XKz9JxU0J1V0P8zJHOJxU0J1V0P8zJHOJxU0J1V0P8zJHOJxU0J', 'admin', NOW());