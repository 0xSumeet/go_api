package database

import (
	"database/sql"

	"github.com/0xSumeet/go_api/internal/models"
	_ "github.com/lib/pq"
)

// Get all products
func GetProducts() ([]models.Product, error) {
	var err error

	query := "SELECT product_id, product_name, category, stock_quantity, price FROM products"
	rows, err := DB.Query(query)
	if err != nil {
		return []models.Product{}, err
	}
	defer rows.Close()

	var products []models.Product

	// loop through the rows and append each product to the slice
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.ProductName, &product.Category, &product.StockQuantity, &product.Price); err != nil {
			return []models.Product{}, err
		}
		// append the data to products
		products = append(products, product)
	}
	return products, nil
}

// Get product by id
func GetProductByID(id int) (models.Product, error) {
	var err error
	var product models.Product
	query := "SELECT product_id, product_name, category, stock_quantity, price FROM products where product_id=$1"
	err = DB.QueryRow(query, id).
		Scan(&product.ID, &product.ProductName, &product.Category, &product.StockQuantity, &product.Price)

	if err == sql.ErrNoRows {
		return models.Product{}, err
	} else if err != nil {
		return models.Product{}, err
	}

	// Return the product as JSON
	return product, nil
}

func GetTotalProductsCount() (int, error) {
	var err error
	var totalProduct int
	query := "SELECT COUNT(*) FROM products"
	err = DB.QueryRow(query).Scan(&totalProduct)
	if err != nil {
		return 0, err
	}
	return totalProduct, nil
}

func PaginateData(pagenumber, limit int) ([]models.Product, error) {
	var offset int

	// Set offset, offset specifies the number of items to skip before starting to display results
	offset = (pagenumber - 1) * limit

	query := "SELECT product_id, product_name, category, stock_quantity, price FROM products LIMIT $1 OFFSET $2"
	rows, err := DB.Query(query, limit, offset)

	defer rows.Close()

	var products []models.Product

	// loop through the rows and append each product to the slice
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.ProductName, &product.Category, &product.StockQuantity, &product.Price); err != nil {
			return []models.Product{}, err
		}

		// append to the product
		products = append(products, product)
	}

	// Check errors during row iteration
	if err = rows.Err(); err != nil {
		return []models.Product{}, err
	}
	return products, nil
}
