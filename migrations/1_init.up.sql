CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    "group" TEXT NOT NULL,
    song TEXT NOT NULL,
    release_date DATE,
    lyrics TEXT,
    link TEXT,

    CONSTRAINT unique_group_song UNIQUE ("group", song)
);