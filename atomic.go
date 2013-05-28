package main

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3" // provides sqlite3 driver
	"code.google.com/p/gompd/mpd"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/frio/limitlessled"
	"github.com/ghthor/gowol"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	computers map[string]string
	alarms    map[string]string
	rooms     map[string]string
)

func AlarmHandler(w http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	isSounding := strings.ToLower(request.FormValue("isSounding")) == "true"

	if host, ok := alarms[params["name"]]; ok {
		alarm, err := mpd.Dial("tcp", host)

		if err != nil {
			http.Error(w, "Could not dial alarm server", 500)
		}

		if isSounding {
			alarm.Play(-1)
			fmt.Fprintf(w, "{\"isSounding\": true}")
		} else {
			alarm.Stop()
			fmt.Fprintf(w, "{\"isSounding\": false}")
		}
	} else {
		http.NotFound(w, request)
	}
}

func ComputerHandler(w http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	isAwake := strings.ToLower(request.FormValue("isAwake")) == "true"

	if macAddr, ok := computers[params["name"]]; ok {
		if isAwake {
			wol.SendMagicPacket(macAddr, "255.255.255.255")
			fmt.Fprintf(w, "{\"isAwake\": true}")
		} else {
			fmt.Fprintf(w, "{\"isAwake\": false}")
		}
	} else {
		http.NotFound(w, request)
	}
}

func RoomHandler(w http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	isOn := strings.ToLower(request.FormValue("isOn")) == "true"
	brightness64, err := strconv.ParseInt(request.FormValue("brightness"), 10, 8)
	temperature64, err := strconv.ParseInt(request.FormValue("temperature"), 10, 8)

	if err != nil {
		http.Error(w, "Couldnt read values", 400)
		return
	}

	brightness := int(brightness64)
	temperature := int(temperature64)

	if bridgeAddress, ok := rooms[params["name"]]; ok {
		// initialize DB conn
		var states *limitlessled.States
		conn, err := sql.Open("sqlite3", "./state.db")

		if err != nil {
			http.Error(w, "No stateful DB", 500)
			return
		}

		// defer conn.Close()
		states = &limitlessled.States{conn}
		originalState, err := states.Retrieve()

		if err != nil {
			http.Error(w, "Couldn't get current state", 500)
		}

		newState := limitlessled.Bulb{Brightness: brightness, Temperature: temperature, IsOn: isOn}

		bridge, err := limitlessled.Dial(bridgeAddress)
		go states.Store(bridge.Set(*originalState, newState))

		w.WriteHeader(http.StatusAccepted)
	} else {
		http.NotFound(w, request)
	}
}

func main() {
	compConfig, compErr := ioutil.ReadFile("computers.json")
	alarmConfig, alarmErr := ioutil.ReadFile("alarms.json")
	roomConfig, roomErr := ioutil.ReadFile("rooms.json")

	if compErr == nil && alarmErr == nil && roomErr == nil {
		json.Unmarshal(compConfig, &computers)
		json.Unmarshal(alarmConfig, &alarms)
		json.Unmarshal(roomConfig, &rooms)

		r := mux.NewRouter()
		r.HandleFunc("/computer/{name}/", ComputerHandler)
		r.HandleFunc("/alarm/{name}/", AlarmHandler)
		r.HandleFunc("/room/{name}/", RoomHandler)
		http.Handle("/", r)
		http.ListenAndServe("localhost:5000", nil)
	} else {
		fmt.Println("Either alarms.json or computers.json does not exist")
	}

}
