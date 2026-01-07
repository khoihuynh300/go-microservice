CREATE TYPE address_type_enum AS ENUM ('home', 'work', 'other');

CREATE TABLE user_addresses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    address_type address_type_enum NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    ward VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);
CREATE UNIQUE INDEX idx_user_addresses_default ON user_addresses(user_id) WHERE is_default = TRUE;
