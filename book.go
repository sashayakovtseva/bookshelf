// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package bookshelf

// Book holds metadata about a book.
type Book struct {
	ID            int64  `json:"-",bson:"-"`
	Title         string `json:"title",bson:"title"`
	Author        string `json:"author",bson:"author"`
	PublishedDate string `json:"published_date",bson:"published_date"`
	Description   string `json:"description",bson:"description"`
}

// BookDatabase provides thread-safe access to a database of books.
type BookDatabase interface {
	// ListBooks returns a list of books, ordered by title.
	ListBooks() ([]*Book, error)

	// ListBooksCreatedBy returns a list of books, ordered by title, filtered by
	// the user who created the book entry.
	ListBooksCreatedBy(userID string) ([]*Book, error)

	// GetBook retrieves a book by its ID.
	GetBook(id int64) (*Book, error)

	// AddBook saves a given book, assigning it a new ID.
	AddBook(b *Book) (id int64, err error)

	// DeleteBook removes a given book by its ID.
	DeleteBook(id int64) error

	// UpdateBook updates the entry for a given book.
	UpdateBook(b *Book) error

	// Close closes the database, freeing up any available resources.
	Close()
}
