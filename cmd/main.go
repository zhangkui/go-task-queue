package main

import (
	"log"
	"net/http"

	"go-task-queue/internal/handler"
	"go-task-queue/internal/store"
)

func main() {
	store := store.NewMemoryStore()
	mux := handler.NewRouter(store)
	log.Println("go-task-queue starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
