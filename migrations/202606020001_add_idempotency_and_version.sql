-- +goose Up
ALTER TABLE users ADD COLUMN version INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '乐观锁版本号';
ALTER TABLE users ADD COLUMN request_id VARCHAR(64) NULL COMMENT '创建幂等请求ID';
CREATE UNIQUE INDEX idx_users_request_id ON users (request_id);

ALTER TABLE roles ADD COLUMN version INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '乐观锁版本号';
ALTER TABLE roles ADD COLUMN request_id VARCHAR(64) NULL COMMENT '创建幂等请求ID';
CREATE UNIQUE INDEX idx_roles_request_id ON roles (request_id);

ALTER TABLE menus ADD COLUMN version INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '乐观锁版本号';
ALTER TABLE menus ADD COLUMN request_id VARCHAR(64) NULL COMMENT '创建幂等请求ID';
CREATE UNIQUE INDEX idx_menus_request_id ON menus (request_id);

ALTER TABLE blogs ADD COLUMN version INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '乐观锁版本号';
ALTER TABLE blogs ADD COLUMN request_id VARCHAR(64) NULL COMMENT '创建幂等请求ID';
CREATE UNIQUE INDEX idx_blogs_request_id ON blogs (request_id);

-- +goose Down
DROP INDEX idx_blogs_request_id ON blogs;
ALTER TABLE blogs DROP COLUMN request_id;
ALTER TABLE blogs DROP COLUMN version;

DROP INDEX idx_menus_request_id ON menus;
ALTER TABLE menus DROP COLUMN request_id;
ALTER TABLE menus DROP COLUMN version;

DROP INDEX idx_roles_request_id ON roles;
ALTER TABLE roles DROP COLUMN request_id;
ALTER TABLE roles DROP COLUMN version;

DROP INDEX idx_users_request_id ON users;
ALTER TABLE users DROP COLUMN request_id;
ALTER TABLE users DROP COLUMN version;
