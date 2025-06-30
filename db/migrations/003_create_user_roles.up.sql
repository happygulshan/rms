-- CREATE TYPE role AS ENUM('admin', 'subadmin', 'user')
CREATE TABLE IF NOT EXISTS user_roles  (
    user_id UUID REFERENCES users(id),
    -- name role NOT NULL
    role_id UUID REFERENCES roles(id),
    PRIMARY KEY (user_id, role_id)
);
