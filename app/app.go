// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Sample bookshelf is a fully-featured app demonstrating several Google Cloud APIs, including Datastore, Cloud SQL, Cloud Storage.
// See https://cloud.google.com/go/getting-started/tutorial-app
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sashayakovtseva/bookshelf"
)

var DB bookshelf.BookDatabase

func main() {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "localhost"
	}

	var err error
	DB, err = bookshelf.NewMongoDB(mongoURL)
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), handler()))
}

func handler() http.Handler {
	r := mux.NewRouter()
	r.Handle("/", http.RedirectHandler("/books", http.StatusFound))

	r.Methods("POST").Path("/books").
		Handler(appHandler(createHandler))
	r.Methods("GET").Path("/books").
		Handler(appHandler(listHandler))
	r.Methods("POST", "PUT").Path("/books/{id:[0-9]+}").
		Handler(appHandler(updateHandler))
	r.Methods("GET").Path("/books/{id:[0-9]+}").
		Handler(appHandler(detailHandler))
	r.Methods("POST").Path("/books/{id:[0-9]+}:delete").
		Handler(appHandler(deleteHandler)).Name("delete")

	r.Methods("GET").Path("/healthz").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})
	return r
}

// createHandler adds a book to the database.
func createHandler(w http.ResponseWriter, r *http.Request) *appError {
	var book bookshelf.Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		return appErrorf(err, "could not decode json book: %v", err)
	}
	id, err := DB.AddBook(&book)
	if err != nil {
		return appErrorf(err, "could not save book: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/books/%d", id), http.StatusFound)
	return nil
}

// listHandler displays a list with summaries of books in the database.
func listHandler(w http.ResponseWriter, r *http.Request) *appError {
	books, err := DB.ListBooks()
	if err != nil {
		return appErrorf(err, "could not list books: %v", err)
	}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(books)
	if err != nil {
		return appErrorf(err, "could not encode books: %v", err)
	}
	return nil
}

// updateHandler updates the details of a given book.
func updateHandler(w http.ResponseWriter, r *http.Request) *appError {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return appErrorf(err, "bad book id: %v", err)
	}
	var book bookshelf.Book
	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		return appErrorf(err, "could not decode json book: %v", err)
	}
	book.ID = id

	err = DB.UpdateBook(&book)
	if err != nil {
		return appErrorf(err, "could not save book: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/books/%d", book.ID), http.StatusFound)
	return nil
}

// detailHandler displays the details of a given book.
func detailHandler(w http.ResponseWriter, r *http.Request) *appError {
	book, err := bookFromRequest(r)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(book)
	if err != nil {
		return appErrorf(err, "could not encode book: %v", err)
	}
	return nil
}

// bookFromRequest retrieves a book from the database given a book ID in the
// URL's path.
func bookFromRequest(r *http.Request) (*bookshelf.Book, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad book id: %v", err)
	}
	book, err := DB.GetBook(id)
	if err != nil {
		return nil, fmt.Errorf("could not find book: %v", err)
	}
	return book, nil
}

// deleteHandler deletes a given book.
func deleteHandler(w http.ResponseWriter, r *http.Request) *appError {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return appErrorf(err, "bad book id: %v", err)
	}
	err = DB.DeleteBook(id)
	if err != nil {
		return appErrorf(err, "could not delete book: %v", err)
	}
	http.Redirect(w, r, "/books", http.StatusFound)
	return nil
}

// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}
