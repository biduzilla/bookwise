-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS books (
    id bigserial PRIMARY KEY,
    title text NOT NULL,
    author text NOT NULL,
    pages integer NOT NULL,
    description text NOT NULL,
    user_id BIGINT NOT NULL,

    version integer NOT NULL DEFAULT 1,
    deleted bool NOT NULL DEFAULT false,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    created_by BIGINT, 
    updated_at timestamp(0) with time zone,
    updated_by BIGINT,
    
    CONSTRAINT fk_books_user FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_books_pages_positive CHECK (pages > 0),
    
    CONSTRAINT unique_title_per_user UNIQUE (user_id, title, deleted)
);

CREATE INDEX IF NOT EXISTS idx_books_user_id ON books(user_id);
CREATE INDEX IF NOT EXISTS idx_books_title ON books(title);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_deleted ON books(deleted) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_books_user_title ON books(user_id, title);

CREATE INDEX IF NOT EXISTS idx_books_title_tsvector 
    ON books USING GIN (to_tsvector('simple', title));
    
CREATE INDEX IF NOT EXISTS idx_books_author_tsvector 
    ON books USING GIN (to_tsvector('simple', author));

CREATE INDEX IF NOT EXISTS idx_books_search 
    ON books(user_id, deleted) 
    INCLUDE (title, author);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS books;
-- +goose StatementEnd