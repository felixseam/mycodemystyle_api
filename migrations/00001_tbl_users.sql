-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_users
(
    id            SERIAL PRIMARY KEY,
    user_uuid   UUID         NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    first_name    VARCHAR(50),
    last_name     VARCHAR(50),
    nickname      VARCHAR(50)  DEFAULT '',
    user_name     VARCHAR(50)  UNIQUE NOT NULL,
    password      VARCHAR(100) NOT NULL,          -- bcrypt hash, never plain text
    email         VARCHAR(50),
    phone         VARCHAR(20),
    address       VARCHAR(100),
    role_id       SMALLINT     NOT NULL,
    gender_id     SMALLINT     DEFAULT 1,
    lang          VARCHAR(10)  DEFAULT 'km',
    last_access   TIMESTAMP,
    last_login    TIMESTAMP,
    login_session VARCHAR(350) DEFAULT gen_random_uuid()::VARCHAR,
    session_share VARCHAR(350),
    remark        VARCHAR(250),
    profile_img   VARCHAR(250),

    status_id     SMALLINT     NOT NULL DEFAULT 1,
    "order"       INT          DEFAULT 0,
    created_by    INT,
    created_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by    INT,
    updated_at    TIMESTAMP,
    deleted_by    INT,
    deleted_at    TIMESTAMP
);

-- Root admin (replace the hash, see below)
INSERT INTO tbl_users (
    first_name, last_name, nickname, user_name, password,
    email, address, role_id, gender_id, lang, remark, status_id, "order"
) VALUES (
    'Admin', 'AD', 'Admin', 'ADMIN', '123456',
    'admin@gmail.com', 'Phnom Penh', 1, 1, 'km', 'Company root user', 1, 0
)
ON CONFLICT (user_name) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS tbl_members;