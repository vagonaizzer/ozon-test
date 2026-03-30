BEGIN;

CREATE TABLE IF NOT EXISTS posts (
    id               VARCHAR(26)  PRIMARY KEY,         
    author_id        VARCHAR(26)  NOT NULL,             
    title            VARCHAR(255) NOT NULL,
    content          TEXT         NOT NULL,
    comments_enabled BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts (created_at ASC);

CREATE TABLE IF NOT EXISTS comments (
    id         BIGSERIAL    PRIMARY KEY,
    post_id    VARCHAR(26)  NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id  VARCHAR(26)  NOT NULL,
    parent_id  BIGINT       REFERENCES comments(id) ON DELETE CASCADE,
    content    VARCHAR(2000) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comments_post_id_root
    ON comments (post_id, id ASC)
    WHERE parent_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_comments_parent_id
    ON comments (parent_id, id ASC)
    WHERE parent_id IS NOT NULL;

COMMIT;
