CREATE TABLE IF NOT EXISTS urls(
    id TEXT PRIMARY KEY NOT NULL,
    url TEXT NOT NULL,
    email TEXT,
    username TEXT,
    hits INTEGER    
);