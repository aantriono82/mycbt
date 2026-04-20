CREATE TABLE password_reset_tokens (
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (token)
);
CREATE INDEX idx_password_reset_email ON password_reset_tokens(email);
