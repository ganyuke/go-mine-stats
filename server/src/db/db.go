package db

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	Monika *data
)

type data struct {
	db                 *sql.DB
	queryCategory      *statement_order
	queryTop           *statement_order
	queryTotalCategory *statement_order
	queryTotalStat     *sql.Stmt
	queryTotal         *sql.Stmt
	queryUuid          *sql.Stmt
	insertNew          *sql.Stmt
	insertHistorical   *sql.Stmt
	updateRow          *sql.Stmt
	checkExist         *sql.Stmt
	checkDifference    *sql.Stmt
}

type statement_order struct {
	asc  *sql.Stmt
	desc *sql.Stmt
}
type Stat_item struct {
	Uuid     string `json:"uuid"`
	Category string `json:"category"`
	Item     string `json:"stat"`
	Value    int    `json:"value"`
	Date     int64  `json:"date"`
	World    string `json:"world"`
}

type Stat_total struct {
	Category string `json:"category"`
	Item     string `json:"stat"`
	Value    int    `json:"value"`
	World    string `json:"world"`
}

type Update_data struct {
	Statistics []*Stat_item
}

type Checkers struct {
	Chess int
}

func UpdatePlayerStat(data *Update_data) {

	context := context.Background()
	transaction, err := Monika.db.BeginTx(context, nil)
	log_error(err, "E_TRANSACTION_FAIL")

	log.Println("COMMIT BEGIN")

	for _, player_entries := range data.Statistics {
		uuid := player_entries.Uuid
		date := player_entries.Date
		category := player_entries.Category
		item := player_entries.Item
		world := player_entries.World
		value := player_entries.Value

		log.Println("Inserting in " + category + " the item " + item + " for player " + uuid)

		var check_obj Checkers

		err := Monika.checkDifference.QueryRow(uuid, category, item, value, world).Scan(&check_obj.Chess)
		log_error(err, "E_SCAN_FAIL")

		// Drop change if row is exactly the same.
		if check_obj.Chess == 1 {
			log.Println("No difference in statistic, dropping change...")
			return
		}

		err = Monika.checkExist.QueryRow(uuid, category, item, world).Scan(&check_obj.Chess)
		log_error(err, "E_SCAN_FAIL")

		// Check if the statistic already exists in the current stat table
		if check_obj.Chess == 1 {
			log.Println("Row exists, updating current stats...")
			_, err = Monika.updateRow.ExecContext(context, date, value, uuid, category, item, world)
			if err != nil {
				transaction.Rollback()
				return
			}
		} else {
			log.Println("Row not found, creating new stat...")
			_, err = Monika.insertNew.ExecContext(context, uuid, date, category, item, value, world)
			if err != nil {
				transaction.Rollback()
				return
			}
		}

		// Add statistic to historical table for tracking over time
		_, err = Monika.insertHistorical.ExecContext(context, uuid, date, category, item, value, world)
		if err != nil {
			transaction.Rollback()
			return
		}

	}

	err = transaction.Commit()
	log_error(err, "E_TRANSACTION_FAIL")

	log.Println("COMMIT END")

}

func RetrievePlayerStat(uuid, category, item, world string) (Stat_item, error) {
	log.Println("Retrieving player " + uuid + " stat for " + item + " in category " + category)
	var stat_obj Stat_item
	row := Monika.queryUuid.QueryRow(uuid, category, item, world)
	err := row.Scan(&stat_obj.Uuid, &stat_obj.Category, &stat_obj.Item, &stat_obj.Value, &stat_obj.Date, &stat_obj.World)
	if err != nil {
		log.Print(err)
		return stat_obj, err
	}
	log.Printf("UUID:%s, Category:%s, Item:%s, Value:%d, Mod. Date:%d\n", stat_obj.Uuid,
		stat_obj.Category, stat_obj.Item, stat_obj.Value, stat_obj.Date)
	return stat_obj, err
}

func GetExtrema(category, item, world, order, limit string) []Stat_item {
	log.Println("Retrieving extrema stat for " + item + " in category " + category)
	if order == "ASC" {
		rows, err := Monika.queryTop.asc.Query(category, item, world, limit)
		log_error(err, "E_QUERY_FAIL")
		return makeList(rows)
	} else {
		rows, err := Monika.queryTop.desc.Query(category, item, world, limit)
		log_error(err, "E_QUERY_FAIL")
		return makeList(rows)
	}
}

func RetrieveTotalCategory(category, world string) (Stat_total, error) {
	var stat_obj Stat_total
	log.Println("Retrieving total stats for entire category " + category)
	row := Monika.queryTotal.QueryRow(category, world)
	err := row.Scan(&stat_obj.Category, &stat_obj.Value, &stat_obj.World)
	if err != nil {
		log.Print(err)
		return stat_obj, err
	}
	return stat_obj, err
}

func RetrieveTotalStat(category, item, world string) (Stat_total, error) {
	var stat_obj Stat_total
	log.Println("Retrieving total stats for " + item + " in category " + category)
	row := Monika.queryTotalStat.QueryRow(category, item, world)
	err := row.Scan(&stat_obj.Category, &stat_obj.Item, &stat_obj.Value, &stat_obj.World)
	if err != nil {
		log.Print(err)
		return stat_obj, err
	}
	return stat_obj, err
}

func GetTotalStatsForCategory(category, world, order, limit string) []Stat_total {
	log.Println("Retrieving total stats for category " + category)
	if order == "ASC" {
		rows, err := Monika.queryTotalCategory.asc.Query(category, world, limit)
		log_error(err, "E_QUERY_FAIL")
		return makeListTotal(rows)
	} else {
		rows, err := Monika.queryTotalCategory.desc.Query(category, world, limit)
		log_error(err, "E_QUERY_FAIL")
		return makeListTotal(rows)
	}
}

func GetStatsForCategory(category, world, order, limit string) []Stat_item {
	log.Println("Retrieving extrema stats for category " + category)
	if order == "ASC" {
		rows, err := Monika.queryCategory.asc.Query(category, world, limit)
		log_error(err, "E_QUERY_FAIL")
		return makeList(rows)
	} else {
		rows, err := Monika.queryCategory.desc.Query(category, world, limit)
		log_error(err, "E_QUERY_FAIL")
		return makeList(rows)
	}
}

func makeListTotal(rows *sql.Rows) []Stat_total {
	var stat_obj Stat_total
	var list []Stat_total
	for rows.Next() {
		rows.Scan(&stat_obj.Category, &stat_obj.Item, &stat_obj.Value, &stat_obj.World)
		log.Printf("Category: %s, Item: %s, Value: %d\n",
			stat_obj.Category, stat_obj.Item, stat_obj.Value)
		list = append(list, stat_obj)
	}
	return list
}

func makeList(rows *sql.Rows) []Stat_item {
	var stat_obj Stat_item
	var list []Stat_item
	for rows.Next() {
		rows.Scan(&stat_obj.Uuid, &stat_obj.Category, &stat_obj.Item, &stat_obj.Value, &stat_obj.Date, &stat_obj.World)
		log.Printf("UUID: %s, Category: %s, Item: %s, Value: %d, Mod. Date: %d\n", stat_obj.Uuid,
			stat_obj.Category, stat_obj.Item, stat_obj.Value, stat_obj.Date)
		list = append(list, stat_obj)
	}
	return list
}

func DbConnect(firstRun bool) *data {
	db, err := sql.Open("sqlite3", "./stats.db")
	if firstRun {
		initDb(db)
	}
	log_error(err, "E_CONNECTION_FAIL")
	return prepareStatements(db)
}

func initDb(db *sql.DB) {
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
	log_error(err, "E_TABLE_FAIL")
	log.Println("Database created.")
}

func prepareStatements(connection *sql.DB) *data {
	prepareFunc := func(statement string) *sql.Stmt {
		prep, err := connection.Prepare(statement)
		log_error(err, "E_PREPARE_FAIL")
		return prep
	}

	init_data := &data{
		db: connection,
		queryCategory: &statement_order{
			asc: prepareFunc(
				`SELECT uuid, stat_category, stat_name, value, date, world 
				FROM stats
				WHERE stat_category = ? AND world = ?
				GROUP BY uuid
				ORDER BY value ASC
				LIMIT ?`,
			),
			desc: prepareFunc(
				`SELECT uuid, stat_category, stat_name, value, date, world 
				FROM stats
				WHERE stat_category = ? AND world = ?
				GROUP BY uuid
				ORDER BY value DESC
				LIMIT ?`,
			),
		},
		queryTop: &statement_order{
			asc: prepareFunc(
				`SELECT uuid, stat_category, stat_name, value, date, world 
				FROM stats
				WHERE stat_category = ? AND stat_name = ? AND world = ?
				GROUP BY uuid
				ORDER BY value ASC
				LIMIT ?`,
			),
			desc: prepareFunc(
				`SELECT uuid, stat_category, stat_name, value, date, world 
				FROM stats
				WHERE stat_category = ? AND stat_name = ? AND world = ?
				GROUP BY uuid
				ORDER BY value DESC
				LIMIT ?`,
			),
		},
		queryTotalCategory: &statement_order{
			asc: prepareFunc(
				`SELECT stat_category, stat_name, SUM(value) AS sumVal, world 
				FROM stats 
				WHERE stat_category = ? AND world = ?
				GROUP BY stat_name
				ORDER BY sumVal ASC
				LIMIT ?`,
			),
			desc: prepareFunc(
				`SELECT stat_category, stat_name, SUM(value) AS sumVal, world 
				FROM stats 
				WHERE stat_category = ? AND world = ?
				GROUP BY stat_name
				ORDER BY sumVal DESC 
				LIMIT ?`,
			),
		},
		queryTotalStat: prepareFunc(
			`SELECT stat_category, stat_name, SUM(value), world 
			FROM stats 
			WHERE stat_category = ? AND stat_name = ? AND world = ?
			`,
		),
		queryTotal: prepareFunc(
			`SELECT stat_category, SUM(value), world 
			FROM stats 
			WHERE stat_category = ? AND world = ?
			`,
		),
		queryUuid: prepareFunc(
			`SELECT uuid, stat_category, stat_name, value, date, world 
			FROM stats 
			WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND world = ?
			ORDER BY date DESC 
			LIMIT 1`,
		),
		insertNew: prepareFunc(
			`INSERT INTO stats 
			(uuid, date, stat_category, stat_name, value, world) 
			VALUES (?, ?, ?, ?, ?, ?)`,
		),
		insertHistorical: prepareFunc(
			`INSERT INTO historical_stats 
			(uuid, date, stat_category, stat_name, value, world) 
			VALUES (?, ?, ?, ?, ?, ?)`,
		),
		updateRow: prepareFunc(
			`
			UPDATE stats
			SET date = ?, value = ?
			WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND world = ?
			`,
		),
		checkExist: prepareFunc(
			`SELECT EXISTS(
				SELECT 1 
				FROM stats 
				WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND world = ? 
				LIMIT 1);
			`,
		),
		checkDifference: prepareFunc(
			`SELECT EXISTS(
				SELECT 1 
				FROM stats 
				WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND value = ? AND world = ?
				LIMIT 1);
			`,
		),
	}
	return init_data
}

func log_error(err error, context string) {
	switch {
	case context == "E_CONNECTION_FAIL":
		context = "Error while opening SQL database: "
	case context == "E_TABLE_FAIL":
		context = "Error while generating database tables: "
	case context == "E_PREPARE_FAIL":
		context = "Error while preparing SQL statements: "
	case context == "E_QUERY_FAIL":
		context = "Error while querying player data: "
	case context == "E_INSERT_FAIL":
		context = "Error while inserting new player data: "
	case context == "E_SELECT_FAIL":
		context = "Error while reading player data: "
	case context == "E_CHECK_FAIL":
		context = "Error while checking existance of player data: "
	case context == "E_UPDATE_FAIL":
		context = "Error while updating player data: "
	case context == "E_SCAN_FAIL":
		context = "Error while scanning SQL database: "
	case context == "E_TRANSACTION_FAIL":
		context = "Error while intiating SQL transaction: "
	default:
		context = "An error has occured: "
	}

	if err != nil {
		log.Fatal(context, err)
	}
}
