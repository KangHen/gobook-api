package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// dbConn returns a SQL database connection object.
//
// The returned connection is pinged to verify the connection is valid.
// If the connection is invalid, the function panics.
func dbConn() (db *sql.DB) {
    dbDriver := "mysql"
    dbUser := "root"
    dbPass := ""
    dbName := "go_books_store"
    db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
    
    if err != nil {
        panic(err.Error())
    }

    err = db.Ping()
    if err != nil {
        panic(err.Error())
    }

    return db
}

type BookInput struct {
    Name string `json:"name"`
    CategoryId int `json:"category_id"`
}

type BookCategory struct {
    ID int
    Name string
}

type Book struct{
    ID int `json:"id"`
    Name string `json:"name"`
    CategoryId int `json:"category_id"`
    CreatedAt string `json:"created_at"`
    UpdatedAt *string `json:"updated_at"`
}

type ResponseBooks struct {
	Message string `json:"message"`
	Data []Book `json:"data"`
}

type ResponseBook struct {
	Message string `json:"message"`
	Data Book `json:"data"`
}

type ResponseError struct {
    Message string `json:"message"`
}

func categoryIndex(w http.ResponseWriter, r *http.Request) {
	var category = map[int]string {
        1: "Mythology",
        2: "Math",
        3: "Historical",
        4: "Mystery",
    }

	data := struct {
		Message string `json:"message"`
		Data map[int]string `json:"data"`
	}{
		Message: "Category List",
		Data: category,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(data)
}

// bookIndex responds to GET requests to "/" and shows all books in the database.
func bookIndex(w http.ResponseWriter, r *http.Request) {
    db := dbConn()
    query := `SELECT id, name, category_id, created_at FROM books;`
    
    rows, err := db.Query(query)

    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)

        return
    }

    defer rows.Close()

    books := []Book{}
    
    for rows.Next() {
        book := Book{}
        rows.Scan(&book.ID, &book.Name, &book.CategoryId, &book.CreatedAt)

        books = append(books, book)
    }
    
    data := ResponseBooks{
		Message: "Books List",
		Data: books,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(data)
    
    defer db.Close()
}

// bookShow responds to GET requests to "/books/{id}" and shows a book with matching id
// from the database.
func bookShow(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    bookId := vars["id"]

    db := dbConn()
    query := "SELECT id, name, category_id, created_at, updated_at FROM books WHERE id = ?"

    var b Book

    err := db.QueryRow(query, bookId).Scan(&b.ID, &b.Name, &b.CategoryId, &b.CreatedAt, &b.UpdatedAt)

    if err != nil {
        fmt.Fprintf(w, "Book with %s not found and has some error %s", bookId, err)

        return
    }

    data := ResponseBook{
		Message: "Book Detail",
		Data: b,
	}

	json.NewEncoder(w).Encode(data)
    
    defer db.Close()
}

// bookStore responds to POST requests to "/books/store" and stores the book in the database.
// It will redirect to "/books" if the book is successfully stored.
func bookStore(w http.ResponseWriter, r *http.Request) {
    var book BookInput

    json.NewDecoder(r.Body).Decode(&book)
    var (
        name = book.Name
        category_id = book.CategoryId
        created_at = time.Now().Format("2006-01-02 15:04:05")
    )

    db := dbConn()
    result, err := db.Exec(`INSERT INTO books (name, category_id, created_at) values (?, ?, ?)`, name, category_id, created_at)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)

        data := ResponseError{
            Message: "Error, Book not stored",
        }

		json.NewEncoder(w).Encode(data)

		return
    }

    bookId, err := result.LastInsertId()

    if err != nil {
        w.Header().Add("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)

        data := ResponseError{
            Message: "Error, Last Insert Id not found",
        }

		json.NewEncoder(w).Encode(data)

		return
    }
    
	data := struct {
		Message string `json:"message"`
		Success bool `json:"success"`
	} {
		Message: "Book Stored",
		Success: bookId > 0,
	}

    w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(data)

    defer db.Close()
}

// bookUpdate responds to POST requests to "/books/update/{id}" and updates the book
// with matching id in the database.
//
// It will redirect to "/books" if the book is successfully updated.
func bookUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

    var bookId = mux.Vars(r)["id"]

    db := dbConn()
    query := "UPDATE books SET name = ?, category_id = ?, updated_at = ? WHERE id = ?"

    _, err := db.Exec(query, r.FormValue("name"), r.FormValue("category_id"), time.Now().Format("2006-01-02 15:04:05"), bookId)

    if err != nil {
        data := ResponseError{
            Message: err.Error(),
        }

		json.NewEncoder(w).Encode(data)
        return
    }

    data := struct {
		Message string `json:"message"`
		Success bool `json:"success"`
	} {
		Message: "Book Updated",
		Success: true,
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(data)

	defer db.Close()
}

// bookDelete responds to POST requests to "/books/delete/{id}" and deletes the book
// with matching id in the database.
//
// It will redirect to "/books" if the book is successfully deleted.
func bookDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

    var bookId = mux.Vars(r)["id"]

    db := dbConn()
    query := "DELETE FROM books WHERE id = ?"

    _, err := db.Exec(query, bookId)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)

		data := ResponseError{
			Message: err.Error(),
		}

		json.NewEncoder(w).Encode(data)

		return
    }

    data := struct {
		Message string `json:"message"`
		Success bool `json:"success"`
	} {
		Message: "Book Deleted",
		Success: true,
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(data)

	defer db.Close()
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Adjust "*" to your domain for security
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass to the next handler
		next.ServeHTTP(w, r)
	})
}

// main is the main entry point for the application.
// It creates a new router and sets up the routes for the books. It then
// starts the server and listens on port 8000.
func main() {
    defer dbConn().Close()

    r := mux.NewRouter()

    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{"message": "Welcome to the Books API"}`))
    })

	r.HandleFunc("/categories", categoryIndex)

    /**
    * Create a new subrouter for books
    * Define the routes for the books
    */
    bookRouter := r.PathPrefix("/books").Subrouter()
    bookRouter.HandleFunc("", bookIndex)
    bookRouter.HandleFunc("/show/{id}", bookShow)
    bookRouter.HandleFunc("/store", bookStore)
    bookRouter.HandleFunc("/update/{id}", bookUpdate)
    bookRouter.HandleFunc("/delete/{id}", bookDelete)

    r.Use(enableCORS)

	fmt.Println("Server is running on port 8000")
    log.Fatal(http.ListenAndServe(":8000", r))
}