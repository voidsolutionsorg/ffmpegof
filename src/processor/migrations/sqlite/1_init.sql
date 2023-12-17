CREATE TABLE IF NOT EXISTS hosts (
    "id" INTEGER PRIMARY KEY,
    "servername" TEXT NOT NULL UNIQUE,
    "hostname" TEXT NOT NULL,
    "weight" INTEGER DEFAULT 1,
    "created" DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS processes (
    "id" INTEGER PRIMARY KEY,
    "host_id" INTEGER,
    "process_id" INTEGER,
    "cmd" TEXT
);
CREATE TABLE IF NOT EXISTS states (
    "id" INTEGER PRIMARY KEY,
    "host_id" INTEGER,
    "process_id" INTEGER,
    "state" TEXT
)