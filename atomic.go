package main

import (
	"code.google.com/p/gompd/mpd"
	"encoding/json"
	"fmt"
	"github.com/ghthor/gowol"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strings"
)

var computers map[string]string
var alarms map[string]string

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

func main() {
	compConfig, compErr := ioutil.ReadFile("computers.json")
	alarmConfig, alarmErr := ioutil.ReadFile("alarms.json")

	if compErr == nil && alarmErr == nil {
		json.Unmarshal(compConfig, &computers)
		json.Unmarshal(alarmConfig, &alarms)

		r := mux.NewRouter()
		r.HandleFunc("/computer/{name}/", ComputerHandler)
		r.HandleFunc("/alarm/{name}/", AlarmHandler)
		http.Handle("/", r)
		http.ListenAndServe("localhost:5000", nil)
	} else {
		fmt.Println("Either alarms.json or computers.json does not exist")
	}

}
