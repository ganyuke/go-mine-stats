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

func Init_poll_official() *pollster {
	polling_center := &pollster{ticker: time.NewTicker(time.Minute * time.Duration(config.Config_file.PollSpeed)), new: make(chan string)}
	go polling_center.Vote()
	return polling_center
}

func (pollster_object *pollster) Vote() {
	for {
		select {
		case <-pollster_object.ticker.C:
			for _, v := range pollster_object.watching {
				go Poll_json(v)
			}
		case new_file := <-pollster_object.new:
			pollster_object.watching = append(pollster_object.watching, new_file)
			println("Now monitoring file: " + new_file)
		}
	}
}

func (pollster_object *pollster) Monitor(new_file string) {
	pollster_object.new <- new_file
}

func (pollster_object *pollster) Remove(file string) {
	for i, v := range pollster_object.watching {
		if v == file {
			pollster_object.watching = append(pollster_object.watching[:i], pollster_object.watching[i+1:]...)
		}
	}
}
