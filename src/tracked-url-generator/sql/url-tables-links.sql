CREATE TABLE IF NOT EXISTS urls (
		"id" TEXT NOT NULL PRIMARY KEY,		
		"url" TEXT,
		"email" TEXT,
		"username" TEXT,
		"hits" INTEGER,
		"created" TEXT,
		"last_modified" TEXT,
		"url_type" TEXT
	  );