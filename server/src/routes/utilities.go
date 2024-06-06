package routes

import (
	"go-mine-stats/src/config"
	"log"
	"strconv"
	"strings"
	"time"
)

func log_error(err error, context string) {
	if err != nil {
		log.Println(context, err)
	}
}

func convertLimit(limit string) (string, error) {
	maxRespLimit := config.Config_file.API.MaxRespLimit
	maxRespString := strconv.Itoa(maxRespLimit)
	limitNum, err := strconv.Atoi(limit)
	if err != nil {
		return limit, err
	}
	if limitNum > maxRespLimit {
		limit = maxRespString
	}
	if limitNum <= 0 {
		limit = "1"
	}
	return limit, nil
}

func convertOrder(order string) string {
	switch strings.ToLower(order) {
	case "desc":
		return "DESC"
	case "asc":
		return "ASC"
	default:
		return "DESC"
	}
}

func parseDate(date string) (time.Time, error) {
	dateParsed, err := time.Parse(time.RFC3339, date)
	if err != nil {
		unixTime, err := strconv.ParseInt(date, 10, 64)
		dateParsed = time.Unix(unixTime, 0)
		log_error(err, "E_TIME_PARSE_FAIL")
		return dateParsed, err
	}
	return dateParsed, nil
}
