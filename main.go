// main.go
package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/shopspring/decimal"

	_ "github.com/go-sql-driver/mysql"
)

type Category struct {
	ID   int
	Name string
}

type Product struct {
	ID       int
	Name     string
	Price    decimal.Decimal
	Category Category
	PriceFmt string // formatted price for display
}

var db *sql.DB
var tmpl = template.Must(template.ParseFiles("template/index.html"))

func main() {
	var err error
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/bookstore")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/tambah", tambahHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/hapus", hapusHandler)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	products := []Product{}
	categories := []Category{}

	rows, _ := db.Query(`SELECT p.id, p.name, p.price, c.id, c.name FROM products p JOIN categories c ON p.category_id = c.id`)
	defer rows.Close()
	for rows.Next() {
		var p Product
		var priceStr string
		rows.Scan(&p.ID, &p.Name, &priceStr, &p.Category.ID, &p.Category.Name)
		p.Price, _ = decimal.NewFromString(priceStr)
		p.PriceFmt = "Rp " + humanize.CommafWithDigits(p.Price.InexactFloat64(), 2)
		products = append(products, p)
	}

	catRows, _ := db.Query("SELECT id, name FROM categories")
	defer catRows.Close()
	for catRows.Next() {
		var c Category
		catRows.Scan(&c.ID, &c.Name)
		categories = append(categories, c)
	}

	tmpl.Execute(w, map[string]interface{}{
		"Products":   products,
		"Categories": categories,
	})
}

func tambahHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		priceStr := r.FormValue("price")
		categoryID := r.FormValue("category_id")
		price, _ := decimal.NewFromString(priceStr)
		db.Exec("INSERT INTO products (name, price, category_id) VALUES (?, ?, ?)", name, price, categoryID)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		name := r.FormValue("name")
		priceStr := r.FormValue("price")
		categoryID := r.FormValue("category_id")
		price, _ := decimal.NewFromString(priceStr)
		_, err := db.Exec("UPDATE products SET name=?, price=?, category_id=? WHERE id=?", name, price, categoryID, id)
		if err != nil {
			log.Println(err)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func hapusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	_, err := db.Exec("DELETE FROM products WHERE id=?", id)
	if err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
