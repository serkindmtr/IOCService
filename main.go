package main

import (
	"IOCService/api"
	"log"
	"net/http"
	"os"
)

func main() {
	f, err := os.OpenFile("IOCService.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Println("Start server")

	//db, err := db.Open("IOCService.db")
	//if err != nil {
	//	log.Println("Fatal error: ", err)
	//	panic(err)
	//}
	//
	//if err := db.Migrate(db); err != nil {
	//	log.Println("Fatal error: ", err)
	//	panic(err)
	//}


	//todo принимаем массив с Attributes -> todo Создать API морду с получением массива Attributes
	api.HandleAndRoute()
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
