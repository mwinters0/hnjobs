CREATE TABLE "hnjobs" (
    "id"	INTEGER NOT NULL UNIQUE,
    "parent"	INTEGER NOT NULL,
    "company"	TEXT NOT NULL,
    "text"	TEXT NOT NULL,
    "time"	INTEGER NOT NULL,
    "fetched_time"	INTEGER NOT NULL,
    "reviewed_time"	INTEGER,
    "why"	TEXT,
    "why_not"	TEXT,
    "score"	INTEGER NOT NULL DEFAULT 0,
    "read"	INTEGER NOT NULL DEFAULT 0,
    "interested"	INTEGER NOT NULL DEFAULT 1,
    "priority"	INTEGER NOT NULL DEFAULT 0,
    "applied"	INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY("id")
);

CREATE TABLE "hnstories" (
    "id"	INTEGER NOT NULL UNIQUE,
    "kids"	TEXT,
    "time"	INTEGER NOT NULL,
    "title"	TEXT,
    "fetched_time"	INTEGER NOT NULL,
    PRIMARY KEY("id")
);