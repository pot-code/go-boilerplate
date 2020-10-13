CREATE DATABASE IF NOT EXISTS go_boilerplate;
USE go_boilerplate;
CREATE TABLE user
(
    id          VARCHAR(32)                         NOT NULL
        PRIMARY KEY,
    username    VARCHAR(32)                         NULL,
    email       VARCHAR(255)                        NULL,
    password    VARCHAR(64)                         NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NULL,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP,
    login_retry INT                                 NOT NULL,
    CONSTRAINT uc_email
        UNIQUE (email),
    CONSTRAINT uc_name
        UNIQUE (username)
);

CREATE TABLE lesson_time_spent
(
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id    VARCHAR(32),
    vocabulary SMALLINT,
    grammar    SMALLINT,
    listening  SMALLINT,
    writing    SMALLINT,
    ts         DATE,
    CONSTRAINT fk_lesson_time_spent FOREIGN KEY (user_id) REFERENCES user (id)
);
CREATE TABLE lesson
(
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    `index`    SMALLINT,
    `name`     VARCHAR(128),
    created_at DATETIME,
    updated_at DATETIME
);
CREATE TABLE lesson_progress
(
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id    VARCHAR(32),
    lesson_id  BIGINT,
    progress   DECIMAL(5, 4),
    created_at DATETIME,
    updated_at DATETIME,
    CONSTRAINT fk_lesson_progress_user FOREIGN KEY (user_id) REFERENCES `user` (id),
    CONSTRAINT fk_lesson_progress_lesson FOREIGN KEY (lesson_id) REFERENCES lesson (id)
);
