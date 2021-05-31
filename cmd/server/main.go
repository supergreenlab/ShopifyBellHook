/*
 * Copyright (C) 2021  SuperGreenLab <towelie@supergreenlab.com>
 * Author: Constantin Clauzel <constantin.clauzel@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	// "fmt"
	// "io"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

var moneyStreamer beep.StreamSeekCloser
var moneySampleRate beep.SampleRate
var receivedStreamer beep.StreamSeekCloser
var receivedSampleRate beep.SampleRate

var sampleRate = beep.SampleRate(22050)

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string // Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)                             // Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host)) // Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	return strings.Join(request, "\n")
}

func OrderPaidHandle(w http.ResponseWriter, req *http.Request) {
	d := map[string]interface{}{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&d); err != nil {
		http.Error(w, fmt.Sprintf("Json parse error %q", err), http.StatusBadRequest)
		return
	}

	moneyStreamer.Seek(0)
	speaker.Play(beep.Resample(4, moneySampleRate, sampleRate, moneyStreamer))

	w.Write([]byte("OK"))
	fmt.Println("Order handle done")
}

var sessionCache = map[string]interface{}{}

func CrispReceivedHandle(w http.ResponseWriter, req *http.Request) {
	d := map[string]interface{}{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&d); err != nil {
		http.Error(w, fmt.Sprintf("Json parse error %q", err), http.StatusBadRequest)
		return
	}

	data := d["data"].(map[string]interface{})
	sessionId, ok := data["session_id"].(string)
	if ok == false {
		w.Write([]byte("NOK"))
		fmt.Println("Crisp received handle: missing session_id")
		return
	}
	if _, ok := sessionCache[sessionId]; ok != true {
		receivedStreamer.Seek(0)
		speaker.Play(beep.Resample(4, receivedSampleRate, sampleRate, receivedStreamer))
		sessionCache[data["session_id"].(string)] = true
	}

	w.Write([]byte("OK"))
	fmt.Println("Crisp received handle done")

}

func openStream(path string) (beep.StreamSeekCloser, beep.SampleRate) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	st, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return st, format.SampleRate
}

func main() {

	moneyStreamer, moneySampleRate = openStream("/usr/local/share/shopifybellhook/money.wav")
	receivedStreamer, receivedSampleRate = openStream("/usr/local/share/shopifybellhook/received.wav")

	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	http.HandleFunc("/order/paid", OrderPaidHandle)
	http.HandleFunc("/crisp/received", CrispReceivedHandle)

	err := http.ListenAndServe(":4200", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
