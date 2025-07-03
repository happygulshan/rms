INSERT INTO roles (name, priority) VALUES
		('admin', 3),
		('subadmin', 2),
		('user', 1)
		ON CONFLICT (name) DO NOTHING;