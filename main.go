package main

import (
	"encoding/json"
	"fmt"
	"kasir-api/database"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port   string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
}

type Categories struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var categories = []Categories{
	{ID: 1, Name: "Electronic", Description: "For electronic devices"},
	{ID: 2, Name: "Clothing", Description: "For clothing items"},
	{ID: 3, Name: "Books", Description: "For books and reading materials"},
}

func getCategoriesByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	for _, c := range categories {
		if c.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(c)
			return
		}
	}

	http.Error(w, "Category not available", http.StatusNotFound)
}

func updateCategoriesByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var updatedCategory Categories
	err = json.NewDecoder(r.Body).Decode(&updatedCategory)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	for i := range categories {
		if categories[i].ID == id {
			categories[i] = updatedCategory

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedCategory)
			return
		}
	}

	http.Error(w, "Category belum ada", http.StatusNotFound)
}
func deleteCategoriesByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	for i, c := range categories {
		if c.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Category successfully deleted",
			})
			categories = append(categories[:i], categories[i+1:]...)
			return
		}
	}
	http.Error(w, "Category not available", http.StatusNotFound)
}

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	config := Config{
		Port:   viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}

	// Setup database
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Setup repository, service, handler
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// Setup routes
	http.HandleFunc("/api/product", productHandler.HandleProducts)
	http.HandleFunc("/api/product/", productHandler.HandleProductByID)

	//localhost:8080/health
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OK",
			"message": "Message is running",
		})
	})

	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(categories)
		} else if r.Method == "POST" {
			var newCategory Categories
			err := json.NewDecoder(r.Body).Decode(&newCategory)
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
			}
			newCategory.ID = len(categories) + 1
			categories = append(categories, newCategory)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(newCategory)
		}
	})

	// GET localhost:8080/categories/{id}
	// PUT localhost:8080/categories/{id}
	// DELETE localhost:8080/categories/{id}
	http.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getCategoriesByID(w, r)
		} else if r.Method == "PUT" {
			updateCategoriesByID(w, r)
		} else if r.Method == "DELETE" {
			deleteCategoriesByID(w, r)
		}
	})

	addr := "localhost:" + config.Port
	fmt.Println("Server running on " + addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Error running server")
	}
}
