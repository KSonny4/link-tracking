CREATE TABLE IF NOT EXISTS pixels (
		"id" TEXT NOT NULL PRIMARY KEY,		
		"url" TEXT,
		"email" TEXT,
		"username" TEXT,
		"hits" INTEGER,
		"created" TEXT,
		"last_modified" TEXT,
		"note" TEXT
	  );