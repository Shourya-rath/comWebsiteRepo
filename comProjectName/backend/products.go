package backend

import (
	"context"
	"fmt"
	"os"
	
	"github.com/jackc/pgx/v5/pgxpool"
)


type Product struct {
	ID          int
	NameEn      string
	NameHi      string
	Slug        string
	Price       int
	Category    string
	Image       *string
	Description *string
}

var Pool *pgxpool.Pool
func Connect() error {
	database_url := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(
		context.Background(),
		// os.Getenv("DATABASE_URL"),
		database_url,
	)

	if err != nil {
		return err
	}

	Pool = pool

	return nil
}
// instead of fetching by start and end 
// it fetch by number of products (limit)
func GetProductsAfterID(lastID int, limit int) ([]Product, error) {
	query := `
		SELECT
			id,
			name_en,
			name_hi,
			slug,
			price,
			category,
			image,
			description
		FROM products
		WHERE id > $1
		ORDER BY id
		LIMIT $2
	`

	rows, err := Pool.Query(
		context.Background(),
		query,
		lastID,
		limit,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var products []Product

	for rows.Next() {
		var p Product

		err := rows.Scan(
			&p.ID,
			&p.NameEn,
			&p.NameHi,
			&p.Slug,
			&p.Price,
			&p.Category,
			&p.Image,
			&p.Description,
		)

		if err != nil {
			return nil, err
		}

		products = append(products, p)
	}

	return products, nil
}
func GetSingleProduct(id int, fail error)(*Product, error){
	if fail != nil {
		fmt.Println("failed parsing the string properly")
		return nil, fail
	}
	query := `
		SELECT
			id,
			name_en,
			name_hi,
			slug,
			price,
			category,
			image,
			description
		FROM products
		WHERE id = $1
	`

	var product *Product = &Product{} ;
	err := Pool.QueryRow(
		context.Background(),
		query,
		id,
	).Scan(
		&product.ID,
		&product.NameEn,
		&product.NameHi,
		&product.Slug,
		&product.Price,
		&product.Category,
		&product.Image,
		&product.Description,
	)
	
	if err != nil {
		return nil, err
	}
	return product,err
}

// SearchProductsFuzzy queries products using pg_trgm for typo-tolerant matching
func SearchProductsFuzzy(searchQuery string, limit int) ([]Product, error) {
    // The % operator uses the trigram index to find matches.
    // We sort by similarity so the best matches show up first.
    query := `
        SELECT
            id,
            name_en,
            name_hi,
            slug,
            price,
            category,
            image,
            description
        FROM products
        WHERE name_en % $1 OR name_hi % $1
        ORDER BY similarity(name_en, $1) DESC, similarity(name_hi, $1) DESC
        LIMIT $2
    `

    rows, err := Pool.Query(
        context.Background(),
        query,
        searchQuery,
        limit,
    )

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var products []Product

    for rows.Next() {
        var p Product
        err := rows.Scan(
            &p.ID,
            &p.NameEn,
            &p.NameHi,
            &p.Slug,
            &p.Price,
            &p.Category,
            &p.Image,
            &p.Description,
        )
        if err != nil {
            return nil, err
        }
        products = append(products, p)
    }

    return products, nil
}
/* -------------------------------------------------------- */
/* old way with the hardcoded array */
type Product_old struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Category string `json:"category"`
	Image    string `json:"image"`
}
func GetProducts() []Product_old{
	products := []Product_old{
		{
			ID:       1,
			Name:     "wet Fruit 1",
			Price:    499,
			Category: "category 1",
			Image:    "dfruit.svg",
		},
		{
			ID:       2,
			Name:     "Dry Fruit 2",
			Price:    1999,
			Category: "category 2",
			Image:    "dfruit.svg",
		},
		{
			ID:       3,
			Name:     "Dry Fruit 3",
			Price:    799,
			Category: "category 3",
			Image:    "dfruit.svg",
		},
		{
			ID:       4,
			Name:     "Dry Fruit 4",
			Price:    2999,
			Category: "category 4",
			Image:    "dfruit.svg",
		},
		{
			ID:       5,
			Name:     "Dry Fruit 5",
			Price:    1499,
			Category: "category 5",
			Image:    "dfruit.svg",
		},
	}

	return products

}