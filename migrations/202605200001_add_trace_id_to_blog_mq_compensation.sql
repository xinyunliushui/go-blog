-- +goose Up
ALTER TABLE blog_mq_compensation
    ADD COLUMN trace_id CHAR(36) NULL COMMENT '关联请求 traceId' AFTER blog_id;

CREATE INDEX idx_blog_mq_compensation_trace_id ON blog_mq_compensation (trace_id);

-- +goose Down
DROP INDEX idx_blog_mq_compensation_trace_id ON blog_mq_compensation;

ALTER TABLE blog_mq_compensation DROP COLUMN trace_id;
