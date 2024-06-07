package db

import (
	"log"
)

var (
	dbTable string = `
	CREATE TABLE "categories" (
		"id"	INTEGER UNIQUE,
		"category"	TEXT UNIQUE,
		PRIMARY KEY("id" AUTOINCREMENT)
	);
	
	CREATE TABLE "players" (
		"id"	INTEGER UNIQUE,
		"uuid"	TEXT UNIQUE,
		"name"	TEXT,
		PRIMARY KEY("id" AUTOINCREMENT)
	);
	
	CREATE TABLE "statistics" (
		"id"	INTEGER UNIQUE,
		"category"	INTEGER,
		"statistic"	TEXT,
		FOREIGN KEY("category") REFERENCES "categories"("id"),
		PRIMARY KEY("id" AUTOINCREMENT)
	);
	
	CREATE TABLE "worlds" (
		"id"	INTEGER UNIQUE,
		"world_path"	TEXT UNIQUE,
		"world_name"	TEXT,
		PRIMARY KEY("id" AUTOINCREMENT)
	);
	
	CREATE TABLE "data" (
		"id"	INTEGER UNIQUE,
		"player"	INTEGER,
		"statistic"	INTEGER,
		"value"	INTEGER,
		"date"	INTEGER,
		"world"	INTEGER,
		FOREIGN KEY("player") REFERENCES "players"("id"),
		FOREIGN KEY("statistic") REFERENCES "statistics"("id"),
		PRIMARY KEY("id" AUTOINCREMENT)
	);
	
	create view denormal as
		select players.name, categories.category, statistics.statistic, value, date, worlds.world_path from data as ld
		inner join worlds on ld.world = worlds.id
		inner join statistics on ld.statistic = statistics.id
		inner join categories on statistics.category = categories.id
		inner join players on ld.player = players.id;
	`
)

func CheckPragma() int {
	rows := Monika.db.QueryRow("pragma user_version")
	var version int
	err := rows.Scan(&version)

	if err != nil {
		log.Fatal("Encountered error while checking SQLite DB pragma.")
	}

	return version
}

func RunMigration() {
	log.Println("Migrating to new DB schema...")

	// Moving data from the old tables to the new tables
	dataInsertion := `
insert into players select DISTINCT null, uuid, name from usernames;
insert into categories select DISTINCT null, stat_category from historical_stats;
insert into worlds select DISTINCT null, world,  null from historical_stats;
insert into statistics select null, id, stat_name from historical_stats
	inner join categories as stat_category on stat_category.category = historical_stats.stat_category
	group by stat_category, stat_name;

insert into data
	select null, ply.id, stat.id, value, date, wld.id from historical_stats as hist
	left join statistics as stat on stat.statistic = hist.stat_name
	inner join categories as cat on stat.category = cat.id and cat.category = hist.stat_category
	inner join players as ply on ply.uuid = hist.uuid
	inner join worlds as wld on wld.world_path = hist.world;
`

	cleanupTables := `
DROP TABLE usernames;
DROP TABLE stats;
DROP TABLE historical_stats;
VACUUM;
`

	_, err := Monika.db.Exec(dbTable)
	if err != nil {
		log.Fatal("Failed to create new tables during DB migration!")
	}

	_, err = Monika.db.Exec(dataInsertion)
	if err != nil {
		log.Fatal("Failed to populate tables during DB migration!")
	}

	_, err = Monika.db.Exec(cleanupTables)
	if err != nil {
		log.Fatal("Failed to delete old tables during DB migration!")
	}

	// Update user_version
	Monika.db.Exec("PRAGMA user_version = 1")

	log.Println("Finished migration to new DB schema.")
}
