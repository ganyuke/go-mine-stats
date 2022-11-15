package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func log_error(err error, context string) {
	if err != nil {
		log.Fatal(context, err)
	}
}

func get_database() *sql.DB {
	db, err := sql.Open("sqlite3", "./stats.db")
	log_error(err, "Error while opening sqlite3 database:")

	return db
}

func Init_db() {

	db := get_database()

	defer db.Close()

	tables := `
	CREATE TABLE stats (
		num INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		uuid TEXT NOT NULL,
		date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		stat_category TEXT NOT NULL,
		stat_name TEXT NOT NULL,
		value INTEGER NOT NULL,
		world TEXT NOT NULL
	);	
	CREATE TABLE historical_stats (
			num INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			uuid TEXT NOT NULL,
			date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			stat_category TEXT NOT NULL,
			stat_name TEXT NOT NULL,
			value INTEGER NOT NULL,
			world TEXT NOT NULL
		);
	`

	_, err := db.Exec(tables)

	log_error(err, "Error while creating tables in sqlite3 database:")

	defer db.Close()

	log.Println("Database created.")

}

type stat_item struct {
	Uuid     string `json:"uuid"`
	Category string `json:"category"`
	Item     string `json:"stat"`
	Value    int    `json:"value"`
	Date     string `json:"date"`
}

type Checkers struct {
	Chess int
}

func UpdatePlayerStat(uuid, date, category, item, world string, value int) {

	log.Println("Inserting in " + category + " the item " + item + " for player " + uuid)

	db := get_database()

	check_row, err := db.Prepare(
		`SELECT EXISTS(
			SELECT 1 
			FROM stats 
			WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND value = ? AND world = ?
			LIMIT 1);
		`)
	log_error(err, "Error while checking if row exists")

	var check_obj Checkers

	check_row.QueryRow(uuid, category, item, value, world).Scan(&check_obj.Chess)

	// Drop change if row is exactly the same.
	if check_obj.Chess == 1 {
		log.Println("No difference in statistic, dropping change...")
		return
	}

	check_row, err = db.Prepare(
		`SELECT EXISTS(
			SELECT 1 
			FROM stats 
			WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND world = ?
			LIMIT 1);
		`)
	log_error(err, "Error while checking if row exists")

	check_row.QueryRow(uuid, category, item, world).Scan(&check_obj.Chess)

	// Check if the statistic already exists in the current stat table
	if check_obj.Chess == 1 {
		log.Println("Row exists, updating current stats...")
		update_row :=
			`
			UPDATE stats
			SET date = ?, value = ?
			WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND world = ?
			`
		prep, err := db.Prepare(update_row)
		log_error(err, "Error while preparing to update player data:")
		_, err = prep.Exec(date, value, uuid, category, item, world)
		log_error(err, "Error while updating player data:")
	} else {
		log.Println("Row not found, creating new stat...")
		prep, err := db.Prepare("INSERT INTO stats (uuid, date, stat_category, stat_name, value, world) VALUES (?, ?, ?, ?, ?, ?)")
		log_error(err, "Error while preparing to insert historical player data:")
		_, err = prep.Exec(uuid, date, category, item, value, world)
		log_error(err, "Error while inserting historical player data:")
	}

	// Add statistic to tracking over time
	prep, err := db.Prepare("INSERT INTO historical_stats (uuid, date, stat_category, stat_name, value, world) VALUES (?, ?, ?, ?, ?, ?)")
	log_error(err, "Error while preparing to insert player data:")

	_, err = prep.Exec(uuid, date, category, item, value, world)
	log_error(err, "Error while inserting player data:")

	defer db.Close()

}

func RetrievePlayerStat(uuid, category, item, world string) (stat_item, error) {

	log.Println("Retrieving player " + uuid + " stat for " + item + " in category " + category)

	db := get_database()

	default_select :=
		`SELECT uuid, stat_category, stat_name, value, date 
	FROM stats 
	WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND world = ?
	ORDER BY date DESC 
	LIMIT 1`

	prep, err := db.Prepare(default_select)
	log_error(err, "Error while preparing to read player data:")

	row := prep.QueryRow(uuid, category, item, world)

	var stat_obj stat_item

	if err := row.Scan(&stat_obj.Uuid, &stat_obj.Category, &stat_obj.Item, &stat_obj.Value, &stat_obj.Date); err != nil {
		log.Print(err)
		return stat_obj, err
	}

	log.Printf("UUID:%s, Category:%s, Item:%s, Value:%d, Mod. Date:%s\n", stat_obj.Uuid,
		stat_obj.Category, stat_obj.Item, stat_obj.Value, stat_obj.Date)
	defer db.Close()

	return stat_obj, err

}

func prepare_statement(sql_statement string, connection *sql.DB) *sql.Stmt {

	prep, err := connection.Prepare(sql_statement)
	log_error(err, "Error while preparing to read player data:")

	return prep

}

func make_list(rows *sql.Rows) []stat_item {

	var stat_obj stat_item
	var list []stat_item

	for rows.Next() {
		rows.Scan(&stat_obj.Uuid, &stat_obj.Category, &stat_obj.Item, &stat_obj.Value, &stat_obj.Date)
		log.Printf("UUID:%s, Category:%s, Item:%s, Value:%d, Mod. Date:%s\n", stat_obj.Uuid,
			stat_obj.Category, stat_obj.Item, stat_obj.Value, stat_obj.Date)
		list = append(list, stat_obj)
	}

	return list

}

func GetExtrema(category, item, world, order, limit string) []stat_item {

	log.Println("Retrieving extrema stat for " + item + " in category " + category)

	sql_statement :=
		`SELECT uuid, stat_category, stat_name, value, date
		FROM stats
		WHERE stat_category = ? AND stat_name = ? AND world = ?
		GROUP BY uuid
		ORDER BY value ` + order + `
		LIMIT ?`

	db := get_database()

	prep := prepare_statement(sql_statement, db)

	rows, err := prep.Query(category, item, world, limit)
	log_error(err, "Error while querying player data:")

	defer db.Close()

	return make_list(rows)

}

func GetStatsForCategory(category, world, order, limit string) []stat_item {

	log.Println("Retrieving extrema stats for category " + category)

	sql_statement :=
		`SELECT uuid, stat_category, stat_name, value, date
		FROM stats
		WHERE stat_category = ? AND world = ?
		GROUP BY uuid
		ORDER BY value ` + order + `
		LIMIT ?`

	db := get_database()

	prep := prepare_statement(sql_statement, db)

	rows, err := prep.Query(category, world, limit)
	log_error(err, "Error while querying player data:")

	defer db.Close()

	return make_list(rows)

}
