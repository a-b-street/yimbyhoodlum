CREATE TABLE IF NOT EXISTS proposals (
	-- A UUID
	id VARCHAR(36) PRIMARY KEY,
	-- In the "us/seattle/udistrict" form
	map_name TEXT NOT NULL CHECK(length(map_name) < 50),
	-- The MapEdits JSON blob
	json BLOB NOT NULL,
	-- An enum state machine
	-- 0 = awaiting moderation,
	-- 1 = public
	-- 2 = spam (keep it, but never publicly list it)
	moderated INTEGER,
	-- Unix time of submission
	time INTEGER
);
