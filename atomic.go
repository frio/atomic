package main

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3" // provides sqlite3 driver
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/frio/atomic/alarms"
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
	rooms     map[string]string
)

func ComputerHandler(w http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)
	isAwake := strings.ToLower(request.FormValue("isAwake")) == "true"

	if macAddr, ok := computers[params["name"]]; ok {
		if isAwake {
			wol.SendMagicPacket(macAddr, "255.255.255.255")
			fmt.Fprintf(w, "{\"IsAwake\": true}")
		} else {
			fmt.Fprintf(w, "{\"IsAwake\": false}")
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
	roomConfig, roomErr := ioutil.ReadFile("rooms.json")

	if compErr == nil && roomErr == nil {
		json.Unmarshal(compConfig, &computers)
		json.Unmarshal(roomConfig, &rooms)

		r := mux.NewRouter()
		r.HandleFunc("/computer/{name}/", ComputerHandler)
		r.Handle("/alarm/{id}/", alarms.Resource)
		r.Handle("/alarm/", alarms.Collection)
		r.HandleFunc("/room/{name}/", RoomHandler)
		http.Handle("/", r)
		http.ListenAndServe("localhost:5000", nil)
	} else {
		fmt.Println("computers.json does not exist")
	}

}
