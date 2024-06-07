package db

import (
	"context"
	"database/sql"
	"go-mine-stats/src/config"
	"log"
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
	queryIndividual    *sql.Stmt
	insertData         *sql.Stmt
	getUsername        *sql.Stmt
	queryUsernameUuids *sql.Stmt
	fkLookups          *foreignKeyLookups
	fkInsertions       *foreignKeyInsertions
}

type foreignKeyLookups struct {
	queryPlayer    *sql.Stmt
	queryStatistic *sql.Stmt
	queryCategory  *sql.Stmt
	queryWorld     *sql.Stmt
}

type foreignKeyInsertions struct {
	insertPlayer    *sql.Stmt
	insertStatistic *sql.Stmt
	insertCategory  *sql.Stmt
	insertWorld     *sql.Stmt
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
	Date     string `json:"date"`
	World    string `json:"world"`
}

type Stat_total struct {
	Category string `json:"category"`
	Item     string `json:"stat"`
	Value    int    `json:"value"`
	World    string `json:"world"`
}

type Stat_cumulative struct {
	Category string `json:"category"`
	Item     string `json:"stat"`
	Value    int    `json:"value"`
	Date     string `json:"date"`
	World    string `json:"world"`
}

type Update_data struct {
	Statistics []*Stat_item
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

	log.Println("COMMIT BEGIN... ")

	for _, player_entries := range data.Statistics {
		uuid := player_entries.Uuid
		date := player_entries.Date
		category := player_entries.Category
		item := player_entries.Item
		world := player_entries.World
		value := player_entries.Value

		// Get the player id for the transaction.
		// If the player does not exist in the players table, create
		// a stub entry for them
		var playerId int64
		err := Monika.fkLookups.queryPlayer.QueryRowContext(ctx, uuid).Scan(playerId)
		if err != nil {
			if err == sql.ErrNoRows {
				var result sql.Result
				result, err = Monika.fkInsertions.insertPlayer.ExecContext(ctx, uuid, nil)
				if err != nil {
					transaction.Rollback()
					log.Fatal("Failed to create new player during transaction!")
					return err
				}
				playerId, err = result.LastInsertId()
				if err != nil {
					transaction.Rollback()
					log.Fatal("Failed to get row id for newly created player during transaction!")
					return err
				}
			} else {
				transaction.Rollback()
				return err
			}
		}

		// Ditto for the category
		var categoryId int64
		err = Monika.fkLookups.queryCategory.QueryRowContext(ctx, category).Scan(categoryId)
		if err != nil {
			if err == sql.ErrNoRows {
				var result sql.Result
				result, err = Monika.fkInsertions.insertCategory.ExecContext(ctx, category)
				if err != nil {
					transaction.Rollback()
					log.Fatal("Failed to create new player during transaction!")
					return err
				}
				categoryId, err = result.LastInsertId()
				if err != nil {
					transaction.Rollback()
					log.Fatal("Failed to get row id for newly created player during transaction!")
					return err
				}
			} else {
				transaction.Rollback()
				return err
			}
		}

		// Ditto for statistic, though we need to use the previously found categoryId
		var statisticId int64
		err = Monika.fkLookups.queryCategory.QueryRowContext(ctx, item).Scan(statisticId)
		if err != nil {
			if err == sql.ErrNoRows {
				var result sql.Result
				result, err = Monika.fkInsertions.insertStatistic.ExecContext(ctx, categoryId, item)
				if err != nil {
					transaction.Rollback()
					log.Fatal("Failed to create new statistic during transaction!")
					return err
				}
				statisticId, err = result.LastInsertId()
				if err != nil {
					transaction.Rollback()
					log.Fatal("Failed to get row id for newly created statistic during transaction!")
					return err
				}
			} else {
				transaction.Rollback()
				return err
			}
		}

		// Finally, we check for world existance
		var worldId int64
		err = Monika.fkLookups.queryCategory.QueryRowContext(ctx, world).Scan(worldId)
		if err != nil {
			if err == sql.ErrNoRows {
				var result sql.Result
				result, err = Monika.fkInsertions.insertWorld.ExecContext(ctx, world, nil)
				if err != nil {
					transaction.Rollback()
					log.Fatal("Failed to create new world during transaction!")
					return err
				}
				worldId, err = result.LastInsertId()
				if err != nil {
					transaction.Rollback()
					log.Fatal("Failed to get row id for newly created world during transaction!")
					return err
				}
			} else {
				transaction.Rollback()
				return err
			}
		}

		// Now we can finally add a single entry to the data table
		_, err = Monika.insertData.ExecContext(ctx, playerId, statisticId, value, date, worldId)
		if err != nil {
			transaction.Rollback()
			log.Fatal("Failed to insert new data into table!")
			return err
		}

	}

	err = transaction.Commit()
	log_error(err, "E_TRANSACTION_FAIL")
	if err != nil {
		transaction.Rollback()
		return err
	}

	log.Print("COMMIT END.")

	return nil

}

func RetrievePlayerStat(uuid, category, item, world string) (Stat_item, error) {
	log.Println("Retrieving player " + uuid + " stat for " + item + " in category " + category)
	var stat_obj Stat_item
	row := Monika.queryIndividual.QueryRow(uuid, category, item, world)
	err := row.Scan(&stat_obj.Uuid, &stat_obj.Category, &stat_obj.Item, &stat_obj.Value, &stat_obj.Date, &stat_obj.World)
	if err != nil {
		log.Print(err)
		return stat_obj, err
	}
	log.Printf("UUID:%s, Category:%s, Item:%s, Value:%d, Mod. Date:%s\n", stat_obj.Uuid,
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

func GetStatDateRange(uuid, category, item, world string, startDate, endDate time.Time) []Stat_item {
	log.Println("Retrieving stat " + item + " between " + startDate.UTC().String() + " and " + endDate.UTC().String() + " for category " + category)
	rows, err := Monika.queryDate.Query(uuid, category, item, world, startDate, startDate, endDate)
	log_error(err, "E_QUERY_FAIL")
	return makeList(rows)
}

func GetCumulativeStat(category, item, world, order string, startDate, endDate time.Time) []Stat_cumulative {
	log.Println("Retrieving cumulative stats for category " + category)
	if order == "ASC" {
		rows, err := Monika.queryCumulative.asc.Query(category, item, world, startDate, endDate)
		log_error(err, "E_QUERY_FAIL")
		return makeListCumulative(rows)
	} else {
		rows, err := Monika.queryCumulative.desc.Query(category, item, world, startDate, endDate)
		log_error(err, "E_QUERY_FAIL")
		return makeListCumulative(rows)
	}
}

func GetWorld(world string) bool {
	log.Println("Checking if \"" + world + "\" in database...")
	var exists int
	row := Monika.fkLookups.queryWorld.QueryRow(world)
	err := row.Scan(&exists)
	if err != nil {
		log.Print(err)
	}
	return exists != 0
}

func InsertUsernames(list []config.Username) error {
	for _, obj := range list {
		_, err := Monika.fkInsertions.insertPlayer.Exec(obj.Uuid, obj.Name)
		if err != nil {
			log_error(err, "E_INSERT_FAIL")
			return err
		}
	}
	return nil
}

func GetPlayers() []config.Username {
	rows, err := Monika.queryUsernameUuids.Query()
	log_error(err, "E_QUERY_FAIL")
	return makeListUsernames(rows)
}

func GetUsernameFromUuid(uuids string) ([]config.Username, error) {
	var player_list []config.Username
	for _, uuid := range strings.Split(uuids, ",") {
		if config.Config_file.CheckBlacklist(uuid) {
			continue
		}
		log.Println("Retriving display name for player " + uuid + "...")
		var player config.Username
		row := Monika.getUsername.QueryRow(uuid)
		err := row.Scan(&player.Uuid, &player.Name)
		if err != nil {
			return player_list, err
		}
		player_list = append(player_list, player)
	}
	return player_list, nil
}

func makeListUsernames(rows *sql.Rows) []config.Username {
	var player_list config.Username
	var list []config.Username
	for rows.Next() {
		rows.Scan(&player_list.Uuid, &player_list.Name)
		list = append(list, player_list)
	}
	return list
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
		log.Printf("UUID: %s, Category: %s, Item: %s, Value: %d, Mod. Date: %s\n", stat_obj.Uuid,
			stat_obj.Category, stat_obj.Item, stat_obj.Value, stat_obj.Date)
		list = append(list, stat_obj)
	}
	return list
}

func DbConnect(firstRun bool, dbPath string) *data {
	db, err := sql.Open("sqlite3", dbPath)
	if firstRun {
		initDb(db)
	}
	log_error(err, "E_CONNECTION_FAIL")
	return prepareStatements(db)
}

func initDb(db *sql.DB) {
	tables := dbTable
	_, err := db.Exec(tables)
	log_error(err, "E_TABLE_FAIL")
	Monika.db.Exec("PRAGMA user_version = 1")
	log.Println("New database created.")
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
				`select name, category, statistic, value, max(date), world_path from denormal
				where category = ? and world_path = ?
				group by name, category, statistic, world_path
				order by value asc
				limit ?`,
			),
			desc: prepareFunc(
				`select name, category, statistic, value, max(date), world_path from denormal
				where category = ? and world_path = ?
				group by name, category, statistic, world_path
				order by value desc
				limit ?`,
			),
		},
		queryTop: &statement_order{
			asc: prepareFunc(
				`select name, category, statistic, value, max(date), world_path from denormal
				where category = ? and statistic = ? and world_path = ?
				group by name, category, statistic, world_path
				order by value asc
				limit ?`,
			),
			desc: prepareFunc(
				`select name, category, statistic, value, max(date), world_path from denormal
				where category = ? and statistic = ? and world_path = ?
				group by name, category, statistic, world_path
				order by value desc
				limit ?`,
			),
		},
		queryTotalCategory: &statement_order{
			asc: prepareFunc(
				`select category, statistic, sum(value) as sumVal, world_path from denormal
				where category = ? and world_path = ?
				group by category, statistic, world_path
				order by sumVal desc
				limit ?`,
			),
			desc: prepareFunc(
				`select category, statistic, sum(value) as sumVal, world_path from denormal
				where category = ? and world_path = ?
				group by category, statistic, world_path
				order by sumVal desc
				limit ?`,
			),
		},
		queryCumulative: &statement_order{
			asc: prepareFunc(
				`with subtracting as (SELECT
					date,
					value,
					LAG ( value, 1, 0 ) OVER (partition BY name ORDER BY date ) prev_val,
					world_path,
					category,
					statistic
				FROM
					denormal 
				WHERE
					category = ? AND
					statistic = ? AND 
					world_path = ?),
				difference as (select
					world_path,
					category,
					statistic, 
					date, 
					(value-prev_val) as diff_val from subtracting),
				summation as (SELECT
					category,
					statistic,
					date,
					world_path,  
					sum(diff_val) over (order by date) AS value from difference)
				SELECT
					category,
					statistic,
					date,
					world_path,
					value
					FROM summation
					WHERE date BETWEEN ? AND ? 
					ORDER BY date ASC
					`,
			),
			desc: prepareFunc(
				`with subtracting as (SELECT
					date,
					value,
					LAG ( value, 1, 0 ) OVER (partition BY name ORDER BY date ) prev_val,
					world_path,
					category,
					statistic
				FROM
					denormal 
				WHERE
					category = ? AND
					statistic = ? AND 
					world_path = ?),
				difference as (select
					world_path,
					category,
					statistic, 
					date, 
					(value-prev_val) as diff_val from subtracting),
				summation as (SELECT
					category,
					statistic,
					date,
					world_path,  
					sum(diff_val) over (order by date) AS value from difference)
				SELECT
					category,
					statistic,
					date,
					world_path,
					value
					FROM summation
					WHERE date BETWEEN ? AND ? 
					ORDER BY date DESC
					`,
			),
		},
		queryTotalStat: prepareFunc(
			`select category, statistic, sum(value) over(), world_path from (
				select name, category, statistic, value, world_path from denormal
				where category = ? and statistic = ? and world_path = ?
				order by date desc 
				) group by name
				limit 1
			`,
		),
		queryTotal: prepareFunc(
			`select category, sum(value), world_path from denormal
			where category = ? and world_path = ?
			group by category, world_path
			`,
		),
		queryIndividual: prepareFunc(
			`select name, category, statistic, value, date, world_path from denormal
			where name = ? and category = ? and statistic = ? and world_path = ?
			order by date desc`,
		),
		queryDate: prepareFunc(
			`select name, category, statistic, value, date, world_path
			FROM denormal WHERE name = ? AND category = ? AND statistic = ? AND world_path = ? 
			AND date
			BETWEEN COALESCE( (SELECT date FROM denormal WHERE date <= ? LIMIT 1),? )
			  AND ?`,
		),
		getUsername: prepareFunc(
			`SELECT uuid, name
			FROM players
			WHERE uuid = ?
			LIMIT 1;
			`,
		),
		queryUsernameUuids: prepareFunc(
			`SELECT uuid, name
			FROM players
			`,
		),
		insertData: prepareFunc(
			`insert into data(player, statistic, value, date, world)
			 values (?,?,?,?,?)
			 `,
		),
		fkLookups: &foreignKeyLookups{
			queryPlayer: prepareFunc(
				`select distinct id from players
				where uuid = ?`,
			),
			queryStatistic: prepareFunc(
				`select distinct statistics.id from statistics
				where category = ? and statistic = ?`,
			),
			queryCategory: prepareFunc(
				`select distinct id from categories
				where category = ?`,
			),
			queryWorld: prepareFunc(
				`select distinct id from worlds
				where world_path = ?`,
			),
		},
		fkInsertions: &foreignKeyInsertions{
			insertPlayer: prepareFunc(
				`INSERT INTO players(uuid, name) VALUES(?, ?)
				ON CONFLICT(uuid) DO UPDATE SET name=excluded.name;`,
			),
			insertCategory: prepareFunc(
				`INSERT INTO categories(category) VALUES(?);`,
			),
			insertStatistic: prepareFunc(
				`INSERT INTO statistics(category, statistic) VALUES(?, ?);`,
			),
			insertWorld: prepareFunc(
				`INSERT INTO worlds(world_path, world_name) VALUES(?, ?);`,
			),
		},
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
