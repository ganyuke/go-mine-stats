package stats

import (
	"go-mine-stats/src/config"
	"time"
)

var (
	Poll_official *pollster
)

// Thank you to https://stackoverflow.com/a/16912587 for this super elegant solution for polling :)

type pollster struct {
	ticker   *time.Ticker
	new      chan string
	watching []string
}

func InitPollOfficial() *pollster {
	polling_center := &pollster{ticker: time.NewTicker(time.Minute * time.Duration(config.Config_file.Scan.PollSpeed)), new: make(chan string)}
	go polling_center.Vote()
	return polling_center
}

func (pollster_object *pollster) Vote() {
	for {
		select {
		case <-pollster_object.ticker.C:
			for _, v := range pollster_object.watching {
				go pollDir(v)
			}
		case new_dir := <-pollster_object.new:
			pollster_object.watching = append(pollster_object.watching, new_dir)
			println("Now monitoring directory: " + new_dir)
		}
	}
}

func (pollster_object *pollster) Monitor(new_dir string) {
	pollster_object.new <- new_dir
}

func (pollster_object *pollster) Remove(dir string) {
	for i, v := range pollster_object.watching {
		if v == dir {
			pollster_object.watching = append(pollster_object.watching[:i], pollster_object.watching[i+1:]...)
		}
	}
}
