package db

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

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
	queryCumulative    *statement_order
	queryTotalStat     *sql.Stmt
	queryTotal         *sql.Stmt
	queryDate          *sql.Stmt
	queryUuid          *sql.Stmt
	insertNew          *sql.Stmt
	insertHistorical   *sql.Stmt
	updateRow          *sql.Stmt
	checkExist         *sql.Stmt
	checkDifference    *sql.Stmt
	checkWorld         *sql.Stmt
	insertUsername     *sql.Stmt
	getUsername        *sql.Stmt
}

type statement_order struct {
	asc  *sql.Stmt
	desc *sql.Stmt
}
type Stat_item struct {
	Uuid     string    `json:"uuid"`
	Category string    `json:"category"`
	Item     string    `json:"stat"`
	Value    int       `json:"value"`
	Date     time.Time `json:"date"`
	World    string    `json:"world"`
}

type Stat_total struct {
	Category string `json:"category"`
	Item     string `json:"stat"`
	Value    int    `json:"value"`
	World    string `json:"world"`
}

type Stat_cumulative struct {
	Category string    `json:"category"`
	Item     string    `json:"stat"`
	Value    int       `json:"value"`
	Date     time.Time `json:"date"`
	World    string    `json:"world"`
}

type Update_data struct {
	Statistics []*Stat_item
}

type checkers struct {
	chess int
}

type Username struct {
	Name string `json:"name"`
	Uuid string `json:"uuid"`
}

func UpdatePlayerStat(data *Update_data) error {

	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	transaction, err := Monika.db.BeginTx(ctx, nil)
	log_error(err, "E_TRANSACTION_FAIL")
	if err != nil {
		return err
	}

	log.Println("COMMIT BEGIN")

	for _, player_entries := range data.Statistics {
		uuid := player_entries.Uuid
		date := player_entries.Date
		category := player_entries.Category
		item := player_entries.Item
		world := player_entries.World
		value := player_entries.Value

		// log.Println("Inserting in " + category + " the item " + item + " for player " + uuid)

		var check_obj checkers

		err := Monika.checkDifference.QueryRowContext(ctx, uuid, category, item, value, world).Scan(&check_obj.chess)
		log_error(err, "E_SCAN_FAIL")
		if err != nil {
			return err
		}

		// Drop change if row is exactly the same.
		if check_obj.chess == 1 {
			log.Println("No difference in statistic, dropping change...")
			return nil
		}

		err = Monika.checkExist.QueryRowContext(ctx, uuid, category, item, world).Scan(&check_obj.chess)
		log_error(err, "E_SCAN_FAIL")
		if err != nil {
			return err
		}

		// Check if the statistic already exists in the current stat table
		if check_obj.chess == 1 {
			// log.Println("Row exists, updating current stats...")
			_, err = Monika.updateRow.ExecContext(ctx, date, value, uuid, category, item, world)
			if err != nil {
				transaction.Rollback()
				return err
			}
		} else {
			// log.Println("Row not found, creating new stat...")
			_, err = Monika.insertNew.ExecContext(ctx, uuid, date, category, item, value, world)
			if err != nil {
				transaction.Rollback()
				return err
			}
		}

		// Add statistic to historical table for tracking over time
		_, err = Monika.insertHistorical.ExecContext(ctx, uuid, date, category, item, value, world)
		if err != nil {
			transaction.Rollback()
			return err
		}

	}

	err = transaction.Commit()
	log_error(err, "E_TRANSACTION_FAIL")
	if err != nil {
		transaction.Rollback()
		return err
	}

	log.Println("COMMIT END")

	return nil

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

func GetStatDateRange(uuid, category, item, world, startDate, endDate string) []Stat_item {
	log.Println("Retrieving stat " + item + " between " + startDate + " and " + endDate + " for category " + category)
	startDateParsed, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		unixTime, err := strconv.ParseInt(startDate, 10, 64)
		startDateParsed = time.Unix(unixTime, 0)
		log_error(err, "E_TIMEPARSE_FAIL")
	}
	endDateParsed, err := time.Parse(time.RFC3339, endDate)
	if err != nil {
		unixTime, err := strconv.ParseInt(endDate, 10, 64)
		endDateParsed = time.Unix(unixTime, 0)
		log_error(err, "E_TIMEPARSE_FAIL")
	}

	rows, err := Monika.queryDate.Query(uuid, category, item, world, startDateParsed, startDateParsed, endDateParsed)
	log_error(err, "E_QUERY_FAIL")
	return makeList(rows)
}

func GetCumulativeStat(category, item, world, order string) []Stat_cumulative {
	log.Println("Retrieving cumulative stats for category " + category)
	if order == "ASC" {
		rows, err := Monika.queryCumulative.asc.Query(category, item, world)
		log_error(err, "E_QUERY_FAIL")
		return makeListCumulative(rows)
	} else {
		rows, err := Monika.queryCumulative.desc.Query(category, item, world)
		log_error(err, "E_QUERY_FAIL")
		return makeListCumulative(rows)
	}
}

func GetWorld(world string) bool {
	log.Println("Checking if \"" + world + "\" in database...")
	var exists int
	row := Monika.checkWorld.QueryRow(world)
	err := row.Scan(&exists)
	if err != nil {
		log.Print(err)
	}
	return exists != 0
}

func InsertUsernames(list []Username) error {
	for _, obj := range list {
		_, err := Monika.insertUsername.Exec(obj.Uuid, obj.Name)
		if err != nil {
			log_error(err, "E_INSERT_FAIL")
			return err
		}
	}
	return nil
}

func GetUsernameFromUuid(uuids string) ([]Username, error) {
	var player_list []Username
	for _, uuid := range strings.Split(uuids, ",") {
		log.Println("Retriving display name for player " + uuid + "...")
		var player Username
		row := Monika.getUsername.QueryRow(uuid)
		err := row.Scan(&player.Uuid, &player.Name)
		if err != nil {
			return player_list, err
		}
		player_list = append(player_list, player)
	}
	return player_list, nil
}

func makeListCumulative(rows *sql.Rows) []Stat_cumulative {
	var stat_obj Stat_cumulative
	var list []Stat_cumulative
	for rows.Next() {
		rows.Scan(&stat_obj.Category, &stat_obj.Item, &stat_obj.Date, &stat_obj.World, &stat_obj.Value)
		log.Printf("Category: %s, Item: %s, Value: %d\n",
			stat_obj.Category, stat_obj.Item, stat_obj.Value)
		list = append(list, stat_obj)
	}
	return list
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
			stat_obj.Category, stat_obj.Item, stat_obj.Value, stat_obj.Date.UTC())
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
	CREATE TABLE usernames (
		num INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		uuid TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL
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
		queryCumulative: &statement_order{
			asc: prepareFunc(
				`with subtracting as (SELECT
							date,
							value,
							LAG ( value, 1, 0 ) OVER (partition BY uuid ORDER BY date ) prev_val,
							world,
							stat_category,
							stat_name
						FROM
							historical_stats 
						WHERE
							stat_category = ? AND
							stat_name = ? AND 
							world = ?),
				difference as (select
					world,
					stat_category,
					stat_name, 
					date, 
					(value-prev_val) as diff_val from subtracting)
				SELECT
					stat_category,
					stat_name,
					date,
					world,  
					sum(diff_val) over (order by date) AS value from difference 
					ORDER BY date ASC`,
			),
			desc: prepareFunc(
				`with subtracting as (SELECT
					date,
					value,
					LAG ( value, 1, 0 ) OVER (partition BY uuid ORDER BY date ) prev_val,
					world,
					stat_category,
					stat_name
				FROM
					historical_stats 
				WHERE
					stat_category = ? AND
					stat_name = ? AND 
					world = ?),
				difference as (select
					world,
					stat_category,
					stat_name, 
					date, 
					(value-prev_val) as diff_val from subtracting)
				SELECT
					stat_category,
					stat_name,
					date,
					world,  
					sum(diff_val) over (order by date) AS value from difference 
					ORDER BY date DESC`,
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
		queryDate: prepareFunc(
			`SELECT uuid, stat_category, stat_name, value, date, world
			FROM historical_stats WHERE uuid = ? AND stat_category = ? AND stat_name = ? AND world = ? 
			AND date
			BETWEEN COALESCE( (SELECT date FROM historical_stats WHERE date <= ? LIMIT 1),? )
			  AND ?`,
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
		checkWorld: prepareFunc(
			`SELECT EXISTS(
				SELECT 1 
				FROM stats 
				WHERE world = ?
				LIMIT 1);
			`,
		),
		insertUsername: prepareFunc(
			`INSERT INTO usernames 
			(uuid, name) 
			VALUES (?, ?)
			ON CONFLICT(uuid) DO
			UPDATE SET name=excluded.name`,
		),
		getUsername: prepareFunc(
			`SELECT uuid, name
			FROM usernames
			WHERE uuid = ?
			LIMIT 1;
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
	case context == "E_TIMEPARSE_FAIL":
		context = "Error while converting string to date: "
	default:
		context = "An error has occured: "
	}

	if err != nil {
		log.Println(context, err)
	}
}
