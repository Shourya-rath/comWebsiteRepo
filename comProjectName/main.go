package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
    "embed"
    // "io/fs" // <-- THIS FIXES: "undefined: fs"

	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/starfederation/datastar-go/datastar"

	"comProjectName/backend"
	"comProjectName/components"
	"comProjectName/views"
)

/* func HandleNextSnacks(w http.ResponseWriter, r *http.Request) {
    // 1. Get offset from URL
    // 2. Fetch snacks from DB
    fmt.Println("handling snacks rn")
    sse := datastar.NewSSE(w, r)

    // 3. Send the new snacks (Appending them to the carousel)
    // Datastar handles the SSE formatting for you!
    sse.PatchElementTempl(
        components.SnackCard("Cheetos", 6.50),
        datastar.WithSelector("#snack-carousel"),
        datastar.WithMode("append"),
    )

    // 4. Update the trigger for the NEXT batch
    // sse.PatchElementTempl(components.CarouselNextTrigger(10))
} */

/* TODO : What is the difference between above and this ? */
// func Home(w http.ResponseWriter, r *http.Request) {
//     components.HomePage().Render(r.Context(), w)
// }

func HandleButton(w http.ResponseWriter, r *http.Request) {
    
    
    sse := datastar.NewSSE(w, r)
    
	fmt.Println("button clicked!")
    sse.PatchElements(`<div>This is the new stuff from the server</div>`)
    
}

func Home(w http.ResponseWriter, r *http.Request) {
    
    var arr []backend.Product  
    arr,err := backend.GetProductsAfterID(0,8)
    if err != nil {
        fmt.Println("ERROR CONNECTING TO THE DATABASE : ")
        fmt.Println(err)
    }
    // fmt.Println(arr)
    // for i,val := range arr{
    //     fmt.Printf("element %d : ",i)
    //     fmt.Printf("%d ",val.ID)
    //     fmt.Printf("%s ",val.NameEn)
    //     fmt.Printf("%s ",val.NameHi)
    //     fmt.Printf("%d ",val.Price)
    //     fmt.Printf("%s ",val.Category)
    //     fmt.Printf("%s \n",val.Image)
    // }
    fmt.Printf("\n-------------------\n")

    // renders the entire home page layout
    component := components.SearchSection(arr)
    
    /* use templ.Handler for updating the component for every request */
    // http.Handle("/", templ.Handler( component ))
    templ.Handler(component).ServeHTTP(w, r)
    
}
func ProductsPage(w http.ResponseWriter, r *http.Request) {
    
    id := strings.TrimPrefix(r.URL.Path,"/products/")
    product,err := backend.GetSingleProduct(strconv.Atoi(id))
    if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "server error", 500)
		return
	}
    
    desc_filler := "no desc"
    if product.Description == nil {
        product.Description = &desc_filler
    }
    
    component := components.ProductPageView(*product)

	templ.Handler(component).ServeHTTP(w, r)
}
func HandleNextProducts(w http.ResponseWriter, r *http.Request) {
    // Only accept POST (matching data-on:click=@post)
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 1. Get lastID from query parameters
    lastIDStr := r.URL.Query().Get("lastID")
    lastID, err := strconv.Atoi(lastIDStr)
    if err != nil {
        http.Error(w, "Invalid lastID parameter", http.StatusBadRequest)
        return
    }

    // 2. Fetch the next block of items from DB
    const limit = 8 
    products, err := backend.GetProductsAfterID(lastID, limit)
    if err != nil {
        fmt.Println("DB Error fetching next items:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Initialize Datastar SSE Stream
    sse := datastar.NewSSE(w, r)

    // Case A: IF DATABASE HAS NO MORE ITEMS: Completely delete the button
    if len(products) == 0 {
        sse.RemoveElementByID("pagination-controls")
        return
    }

    // Case B: We found items -> Append them into the grid element
    sse.PatchElementTempl(
        components.ProductGridItems(products),
        datastar.WithSelectorID("productGrid"),
        datastar.WithModeAppend(),
    )

    // Case C: Check if we reached the end of the DB records
    if len(products) < limit {
        // If the returned block is smaller than our limit, DB is empty now
        sse.RemoveElementByID("pagination-controls")
    } else {
        // Otherwise, replace old button with updated tracking details
        newLastID := products[len(products)-1].ID
        sse.PatchElementTempl(
            components.LoadMoreButton(newLastID),
            datastar.WithSelectorID("pagination-controls"),
            datastar.WithModeOuter(), // Replaces wrapper content cleanly
        )
    }
}

type SearchSignals struct {
    Search string `json:"search"`
}

func HandleSearchProducts(w http.ResponseWriter, r *http.Request) {
    // Datastar defaults to POST for sending signals safely
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 1. Decode the signals sent by Datastar
    var signals SearchSignals
    if err := json.NewDecoder(r.Body).Decode(&signals); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    const limit = 8
    var products []backend.Product
    var err error

    // 2. Decide whether to fetch default data or execute fuzzy search
    if signals.Search == "" {
        // If search input is completely cleared, fetch the first page of products
        products, err = backend.GetProductsAfterID(0, limit)
    } else {
        products, err = backend.SearchProductsFuzzy(signals.Search, limit)
    }

    if err != nil {
        fmt.Println("Search DB Error:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 3. Initialize Datastar SSE Stream
    sse := datastar.NewSSE(w, r)

    // Case A: No matching items found
    // if len(products) == 0 {
    //     sse.PatchElements(
    //         `<p class="text-center col-span-full py-8 text-lg">No products found matching your search 😢</p>`,
    //         datastar.WithSelectorID("productGrid"),
    //     )
    //     sse.RemoveElementByID("pagination-controls")
    //     return
    // }

    // Case B: We found matching items -> Replace the grid content entirely
    sse.PatchElementTempl(
        components.ProductGridItems(products),
        datastar.WithSelectorID("productGrid"),
        datastar.WithModeInner(),
        // Note: Default behavior without WithModeAppend() replaces the inner HTML cleanly
    )

    // Case C: Update pagination controls
    if signals.Search != "" {
        // Hide pagination/load-more while the user is actively sorting via a search query
        sse.RemoveElementByID("pagination-controls")
    } else {
        // Re-display the load more button if they cleared their search query
        newLastID := products[len(products)-1].ID
        
        sse.PatchElementTempl(
            components.LoadMoreButton(newLastID),
            datastar.WithSelectorID("pagination-wrapper"),
            datastar.WithModeInner(), 
        )
    }
}
func LandingPage(w http.ResponseWriter, r *http.Request) {
    // Prevent root route from grabbing random subpaths
    if r.URL.Path != "/" {
        http.NotFound(w, r)
        return
    }

    pageTitle := "Vanaushadhi | Traditional Indian Herbal Care"
    
    // Compose your new landing layouts together (from your "comWebsite/views" directory)
    // Make sure your import statement at the top reads: "comWebsite/views"
    component := views.ViewsLayout(pageTitle, views.Index())

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    err := component.Render(r.Context(), w)
    if err != nil {
        fmt.Printf("Pipeline Error rendering landing layouts: %v\n", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
func ContactPage(w http.ResponseWriter, r *http.Request) {
    // Prevent root route from grabbing random subpaths
    // if r.URL.Path != "/contact" {
    //     http.NotFound(w, r)
    //     return
    // }

    pageTitle := "Vanaushadhi | Contact Us"
    
    // Compose your new landing layouts together (from your "comWebsite/views" directory)
    // Make sure your import statement at the top reads: "comWebsite/views"
    component := views.ViewsLayout(pageTitle, views.ContactSection())

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    err := component.Render(r.Context(), w)
    if err != nil {
        fmt.Printf("Pipeline Error rendering landing layouts: %v\n", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
/* type Product struct {
	ID          int
	NameEn      string
	NameHi      string
	Slug        string
	Price       int
	Category    string
	Image       *string
	Description *string
} */

// Following comment is necessary
//go:embed static
var staticEmbedFS embed.FS
func main() {
    
    appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// Only attempt to load .env if we are working locally
	if appEnv == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Critical: Error loading .env file in development: %v", err)
		}
		log.Println("Loaded configuration from local .env file")
	} else {
		log.Printf("Running in %s mode. Using system environment variables.", appEnv)
	}
    
    backend.Connect()
    
	port := os.Getenv("PORT")
	
    
    // 2. Serve Static Assets (Important for output.css, theme.css, and images)
    /*fs := http.FileServer(http.Dir("static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    */
    /*
    fs := http.FileServer(http.Dir("static"))

    // Create a custom handler to fix Vercel's missing MIME type environment maps
    staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasSuffix(r.URL.Path, ".css") {
            w.Header().Set("Content-Type", "text/css; charset=utf-8")
        }
        // Strip the prefix and serve the file via standard FileServer
        http.StripPrefix("/static/", fs).ServeHTTP(w, r)
    })

    // Bind it to your mux
    http.Handle("/static/", staticHandler)
    */
    // 3. Extract the "static" sub-filesystem safely
	
    // 2. Serve Static Assets (Embedded and explicitly typed for Vercel)
    // 1. Pass the raw embedded filesystem straight to the file server
    fileServer := http.FileServer(http.FS(staticEmbedFS))

    // 2. Handle the route with the MIME type fix
    staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Explicitly set the CSS header to satisfy the browser's strict nosniff requirement
        if strings.HasSuffix(r.URL.Path, ".css") {
            w.Header().Set("Content-Type", "text/css; charset=utf-8")
        }
        
        // CRITICAL: Do NOT use http.StripPrefix here.
        // The filesystem contains the "static" folder at its root level,
        // so it needs the incoming path to keep "/static/css/theme.css" intact!
        fileServer.ServeHTTP(w, r)
    })

    // Bind it directly to the global HTTP router
    http.Handle("/static/", staticHandler)
    
    http.HandleFunc("/", LandingPage)                           
    http.HandleFunc("/shop", Home)  
    http.HandleFunc("/contact/",ContactPage)                            
    http.HandleFunc("/products/",ProductsPage)
    http.HandleFunc("/products/next", HandleNextProducts)
    http.HandleFunc("/products/search", HandleSearchProducts)
    
    fmt.Println("server engine hosting live at http://localhost:",port)
    log.Fatal(http.ListenAndServe(":" + port, nil))
}
