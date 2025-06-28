
CREATE TYPE place AS ENUM ('home', 'work');
CREATE TABLE IF NOT EXISTS addresses  (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    label place NOT NULL DEFAULT 'home'
);
