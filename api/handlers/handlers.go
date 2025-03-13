package handlers

import (
	"fmt"
	"net/http"
)

func ListNotes(w http.ResponseWriter, r *http.Request) {

}

func GetNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintln(w, "Note Id =", id)
}

func CreateNote(w http.ResponseWriter, r *http.Request) {

}

func UpdateNote(w http.ResponseWriter, r *http.Request) {

}

func DeleteNote(w http.ResponseWriter, r *http.Request) {

}
