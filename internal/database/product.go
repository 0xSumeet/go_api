package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Product struct {
	ID            int       `json:"id"`
	ProductName   string    `json:"product_name"`
	Category      string    `json:"category"`
	StockQuantity int       `json:"stock_quantity"`
	Price         float64   `json:"price"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

func CreateProduct(product *Product) (*Product, error) {
	var productID int
	// Modify the query to return the ID and created_at timestamp
	query := "INSERT INTO products (product_name, category, stock_quantity, price) VALUES ($1, $2, $3, $4) RETURNING product_id"

	// Execute the query and get the new product's ID
	err := DB.QueryRow(query, product.ProductName, product.Category, product.StockQuantity, product.Price).
		Scan(&productID)
	if err != nil {
		return nil, fmt.Errorf("could not create product: %v", err)
	}

	// Return the product response with the new ID
	productResponse := &Product{
		ID:            productID,
		ProductName:   product.ProductName,
		Category:      product.Category,
		StockQuantity: product.StockQuantity,
		Price:         product.Price,
	}
	return productResponse, nil
}

func UpdateProductField(product *Product) (*Product, error) {
	var updatedProduct Product

	//	query := `UPDATE products
	//        SET product_name = COALESCE($1, product_name),
	//            category = COALESCE($2, category),
	//            stock_quantity = COALESCE($3, stock_quantity),
	//            price = COALESCE($4, price),
	//            updated_at = NOW()
	//        WHERE product_id = $5
	//        RETURNING product_id, product_name, category, stock_quantity, price, updated_at;`
	query := `UPDATE products
        SET 
            product_name = COALESCE(NULLIF($1, ''), product_name),
            category = COALESCE(NULLIF($2, ''), category),
            stock_quantity = COALESCE(NULLIF($3, 0), stock_quantity),
            price = COALESCE(NULLIF($4, 0), price),
            updated_at = NOW()
        WHERE product_id = $5
        RETURNING product_id, product_name, category, stock_quantity, price, updated_at;`

	// Execute the query
	err := DB.QueryRow(query, product.ProductName, product.Category, product.StockQuantity, product.Price, product.ID).
		Scan(&updatedProduct.ID, &updatedProduct.ProductName, &updatedProduct.Category, &updatedProduct.StockQuantity, &updatedProduct.Price, &updatedProduct.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("could not update product: %v", err)
	}
	// Return the updated product details
	return &updatedProduct, nil
}

//
//func AddTryProduct(product *Product) (*Product, error) {
//	var err error
//	query := "INSERT INTO products (product_name, category, stock_quantity, price) VALUES ($1, $2, $3, $4)"
//	err = DB.QueryRow(
//		query,
//		product.ProductName,
//		product.Category,
//		product.StockQuantity,
//		product.Price,
//	)
//	if err != nil {
//		return &Product{}, fmt.Errorf("could not create user")
//	}
//
//	productResponse := &Product{
//		ProductName:   product.ProductName,
//		Category:      product.Category,
//		StockQuantity: product.StockQuantity,
//		Price:         product.Price,
//	}
//	return productResponse, nil
//}

// Get all products
func GetProducts() ([]Product, error) {
	var err error

	query := "SELECT product_id, product_name, category, stock_quantity, price FROM products"
	rows, err := DB.Query(query)
	if err != nil {
		return []Product{}, err
	}
	defer rows.Close()

	var products []Product

	// loop through the rows and append each product to the slice
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.ProductName, &product.Category, &product.StockQuantity, &product.Price); err != nil {
			return []Product{}, err
		}
		// append the data to products
		products = append(products, product)
	}
	return products, nil
}

// Get product by id
func GetProductByID(id int) (Product, error) {
	var err error
	var product Product
	query := "SELECT product_id, product_name, category, stock_quantity, price FROM products where product_id=$1"
	err = DB.QueryRow(query, id).
		Scan(&product.ID, &product.ProductName, &product.Category, &product.StockQuantity, &product.Price)

	if err == sql.ErrNoRows {
		return Product{}, err
	} else if err != nil {
		return Product{}, err
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

func PaginateData(pagenumber, limit int) ([]Product, error) {
	var offset int

	// Set offset, offset specifies the number of items to skip before starting to display results
	offset = (pagenumber - 1) * limit

	query := "SELECT product_id, product_name, category, stock_quantity, price FROM products LIMIT $1 OFFSET $2"
	rows, err := DB.Query(query, limit, offset)

	defer rows.Close()

	var products []Product

	// loop through the rows and append each product to the slice
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.ProductName, &product.Category, &product.StockQuantity, &product.Price); err != nil {
			return []Product{}, err
		}

		// append to the product
		products = append(products, product)
	}

	// Check errors during row iteration
	if err = rows.Err(); err != nil {
		return []Product{}, err
	}
	return products, nil
}

//func UpdateProduct(product *Product) (*Product, error) {
//	var product Product
//	/*
//	   query := // Update Query here
//	   updateQuery :=
//
//	   query = // Fetch Query here
//	   fetchQuery
//	*/
//	return &product, nil
//}
