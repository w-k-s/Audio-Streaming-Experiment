package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
)

type FlushWriter struct {
	flusher http.Flusher
	writer  io.Writer
}

func NewFlushWriter(w http.ResponseWriter) *FlushWriter {
	fw := &FlushWriter{writer: w}
	if flusher, ok := w.(http.Flusher); ok {
		fw.flusher = flusher
	}

	return fw
}

func (fw *FlushWriter) Write(bytes []byte) (bytesRead int, err error) {
	bytesRead, err = fw.writer.Write(bytes)
	if fw.flusher != nil {
		fw.flusher.Flush()
	}
	return
}

func RootHandler(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)

}

func CelsiusToFahrenheitHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	celsius, err := strconv.Atoi(vars["temp"])
	if err != nil {
		http.Error(w, "Temperature not provided", http.StatusInternalServerError)
		return
	}

	fahrenheit := (float32(celsius) * 1.80) + 32

	fmt.Fprintf(w, "%.1f", fahrenheit)
}

func AudioHandler(w http.ResponseWriter, r *http.Request) {

	audioFile, err := os.Open("bensound-betterdays.mp3")

	if err != nil {
		panic(err.Error())
	}

	defer audioFile.Close()

	reader := bufio.NewReader(audioFile)
	buffer := make([]byte, 1024)
	fw := NewFlushWriter(w)
	bytesDelivered := 0

	w.Header().Add("Content-type", "audio/mpeg")
	w.Header().Add("Content-Transfer-Encoding", "binary")

	for {
		bytesRead, err := reader.Read(buffer)

		if err != nil && err != io.EOF {
			panic(err.Error())
		}

		bytesDelivered += bytesRead
		if bytesRead == 0 {
			break
		}

		fw.Write(buffer)
	}

	fmt.Printf("Request: %s, Bytes Delivered: %d bytes\n", r.RemoteAddr, bytesDelivered)
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/", RootHandler)
	router.HandleFunc("/celsius/{temp}/fahrenheit", CelsiusToFahrenheitHandler)
	router.HandleFunc("/audio", AudioHandler)

	http.Handle("/", router)

	fmt.Println("Starting server at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
