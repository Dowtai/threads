CREATE TABLE posts(
    id UUID PRIMARY KEY,
    author TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    comments_allowed BOOLEAN NOT NULL DEFAULT TRUE
);

create TABLE comments(
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    author TEXT NOT NULL,
    text VARCHAR(2000) NOT NULL
);