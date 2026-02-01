package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"strconv"
	"strings"
	"os"
	"log"
	"categories-api/database"
	"categories-api/repositories"
	"categories-api/services"
	"categories-api/handlers"

	"github.com/spf13/viper"
)

type Config struct {
	Port    string `mapstructure:"PORT"`
	DBConn	string `mapstructure:"DB_CONN"`
}

type Category struct {
	ID		int	`json:"id"`
	Name		string	`json:"name"`
	Description	string	`json:"description"`
}

var categories = []Category{
	{ID: 1, Name: "Economy", Description: "This is economy class"},
	{ID: 2, Name: "Bisnis", Description: "This is bisnis class"},
}

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	config := Config{
		Port: viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}

	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()



	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)



	http.HandleFunc("/api/produk", productHandler.HandleProducts)
	http.HandleFunc("/api/produk/", productHandler.HandleProductByID)










	http.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getCategoryByID(w, r)
		} else if r.Method == "PUT" {
			updateCategory(w, r)
		} else if r.Method == "DELETE" {
			deleteCategory(w, r)
		}
	})

	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(categories)

		} else if r.Method == "POST" {
			var newCategory Category
			err := json.NewDecoder(r.Body).Decode(&newCategory)
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			newCategory.ID = len(categories) + 1
			categories = append(categories, newCategory)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newCategory)
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OK",
			"message": "API Running",
		})
	})

	addr := "0.0.0.0:" + config.Port
	fmt.Println("Server running di", addr)

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("gagal running server", err)
	}
}

func getCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	for _, c := range categories {
		if c.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(c)
			return
		}
	}

	http.Error(w, "Category is not found", http.StatusNotFound)
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	var updateCategory Category
	err = json.NewDecoder(r.Body).Decode(&updateCategory)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	for i := range categories {
		if categories[i].ID == id {
			updateCategory.ID = id
			categories[i] = updateCategory

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updateCategory)
			return
		}
	}

	http.Error(w, "Category is not found", http.StatusNotFound)
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Category ID", http.StatusBadRequest)
		return
	}

	for i, c := range categories {
		if c.ID == id {
			categories = append(categories[:i], categories[i+1:]...)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "sukses delete",
			})
			return
		}
	}

	http.Error(w, "Category is not found", http.StatusNotFound)
}
