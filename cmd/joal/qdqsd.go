package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	myHelloWorldFunc2 := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"libelle": "montruc"}`)
	}
	handlerFromHelloFunc := http.HandlerFunc(myHelloWorldFunc2)
	http.Handle("/trucs", handlerFromHelloFunc)

	log.Println("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
