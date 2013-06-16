package alarms

import (
	"code.google.com/p/gompd/mpd"
	"database/sql"
	"encoding/json"
	"github.com/frio/restful"
)

var (
	Resource   *restful.Resource
	Collection *restful.Collection
)

type Alarm struct {
	Room       string
	Host       string
	IsSounding bool
}

func post(encoded json.Decoder) (interface{}, error) {
	conn, err := sql.Open("sqlite3", "./state.db")
	defer conn.Close()

	if err != nil {
		return nil, err
	}

	var alarm Alarm
	err = encoded.Decode(&alarm)

	if err != nil {
		return nil, err
	}

	query := "INSERT INTO alarm(id, host, isSounding) VALUES(?, ?, ?)"
	_, err = conn.Exec(query, &alarm.Room, &alarm.Host, &alarm.IsSounding)

	if err != nil {
		return nil, err
	}

	return &alarm, nil
}

func get(id string) (interface{}, error) {
	conn, err := sql.Open("sqlite3", "./state.db")
	defer conn.Close()

	if err != nil {
		return nil, err
	}

	alarm := Alarm{}

	query := "SELECT id, host, isSounding FROM alarm WHERE id = ?"
	err = conn.QueryRow(query, id).Scan(&alarm.Room, &alarm.Host, &alarm.IsSounding)

	if err != nil {
		return nil, err
	}

	return *&alarm, nil
}

func put(originalFrom interface{}, encodedTo json.Decoder) (interface{}, error) {
	var to, from Alarm

	from = originalFrom.(Alarm)
	err := encodedTo.Decode(&to)

	if err != nil {
		return nil, err
	}

	alarm, err := mpd.Dial("tcp", from.Host)

	if err != nil {
		return from, err
	}

	if to.IsSounding {
		err = alarm.Play(-1)
	} else {
		err = alarm.Stop()
	}

	if err != nil {
		return from, err
	}

	return to, nil
}

func init() {
	Resource = &restful.Resource{
		Get: get,
		Put: put,
	}
	Collection = &restful.Collection{
		Post: post,
	}
}
