CREATE TABLE IF NOT EXISTS "urls"(
    	id varchar(255) PRIMARY KEY,
    	full_url varchar(500) UNIQUE NOT NULL,
    	short_url varchar(255) UNIQUE NOT NULL,
        user_id varchar(255) NOT NULL
	);