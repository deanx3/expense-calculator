package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Expense struct {
	mgm.DefaultModel `bson:",inline"`
	ExpenseLocation string  `json:"expense_location" bson:"expense_location" form:"expense_location"`
	Category        string  `json:"category" bson:"category" form:"category"`
	Amount         float64 `json:"amount" bson:"amount" form:"amount"`
	Description    string  `json:"description" bson:"description" form:"description"`
}

func initMongoDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

  dbURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_NAME")
	if dbURI == "" || dbName == "" {
		log.Fatal("Missing MongoDB credentials in environment variables")
	}

	if err := mgm.SetDefaultConfig(nil, dbName, options.Client().ApplyURI(dbURI)); err != nil {
		log.Fatal(err)
	}
}

func main() {
	title := "Expense Manager"
	initMongoDB()

	f, _ := os.Create("gin.log")
	gin.DefaultWriter =io.MultiWriter(f, os.Stdout)

	router := gin.Default()

	// Set up template rendering
	router.HTMLRender = ginview.Default()

	// Serve the formF page
	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": false})
	})

	// Handle form submission
	router.POST("/submit", func(ctx *gin.Context) {
		var formData Expense

		// Bind form data
		if err := ctx.ShouldBind(&formData); err != nil {
			ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": err.Error()})
			return
		}
		log.Println("out--->amout",formData.Amount)

		// Insert into MongoDB using mgm
		err := mgm.Coll(&formData).Create(&formData)
		if err != nil {
			ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": err.Error()})
			return
		}

		ctx.Redirect(http.StatusMovedPermanently, "/list")
		// Show success popup
	})

router.GET("/api/suggestions", func(c *gin.Context) {
  query := c.DefaultQuery("expense_location", "") // Get query parameter 'q'
  if query == "" {
    // If query is empty, return an empty suggestions list
    c.HTML(http.StatusOK, "empty_suggestions.html", nil)
    return
  }

  // Define a slice to hold the suggestions
  var suggestions []Expense

  // Build the filter to search for expense_location using a regex query
  filter := bson.M{"expense_location": bson.M{"$regex": query, "$options": "i"}}

  // Fetch data from MongoDB using mgm
  err := mgm.Coll(&Expense{}).SimpleFind(&suggestions, filter)
  if err != nil {
    // If there's an error during the find operation, return an error response
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
    return
  }

  // Prepare the data for rendering
  locations := make([]string, len(suggestions))
  for i, s := range suggestions {
    locations[i] = s.ExpenseLocation
  }

  // Render the suggestions as a list of <li> elements
  c.HTML(http.StatusOK, "suggestions.html", gin.H{
    "suggestions": locations,
  })
})

router.GET("/list", func(c *gin.Context) {
    // Pagination parameters
    page, err := strconv.Atoi(c.DefaultQuery("page", "1")) // Default to page 1 if not provided
    if err != nil || page < 1 {
        page = 1
    }
    limit, err := strconv.Atoi(c.DefaultQuery("limit", "10")) // Default to limit 10 if not provided
    if err != nil || limit < 1 {
        limit = 10
    }

    // Set skip (offset) based on the current page
    skip := (page - 1) * limit

    // Aggregation pipeline
    pipeline := []interface{}{
        // Stage 1: Use $facet to run multiple stages in parallel
        map[string]interface{}{
            "$facet": map[string]interface{}{
                "data": []interface{}{
						map[string]interface{}{
                        "$sort": map[string]interface{}{"created_at": -1}, // Sort by latest created_at
                    },
                    map[string]interface{}{
                        "$skip": skip,
                    },
                    map[string]interface{}{
                        "$limit": limit,
                    },
                },
                "totalCount": []interface{}{
                    map[string]interface{}{
                        "$count": "total", // Count total documents
                    },
                },
            },
        },
    }

    // Run aggregation query
    var result []struct {
        Data       []Expense `bson:"data"`
        TotalCount []struct {
            Total int `bson:"total"`
        } `bson:"totalCount"`
    }

    // Run aggregation on the collection
    cursor, err := mgm.Coll(&Expense{}).Aggregate(mgm.Ctx(), pipeline)
    if err != nil {
        log.Println(err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expenses"})
        return
    }

    defer cursor.Close(mgm.Ctx())

    if err = cursor.All(mgm.Ctx(), &result); err != nil {
        c.JSON(http.StatusOK, gin.H{"error": err.Error()})
        return
    }

    // If there are no results
    if len(result) == 0 {
        c.HTML(http.StatusOK, "list.html", gin.H{
            "expenses":    []Expense{},
            "currentPage": page,
            "totalPages":  0,
            "prevPage":    0,
            "nextPage":    0,
        })
        return
    }

    // Extract total count from the aggregation result
    totalCount := 0
    if len(result[0].TotalCount) > 0 {
        totalCount = result[0].TotalCount[0].Total
    }

    // Calculate total pages
    totalPages := (totalCount + limit - 1) / limit

    // Calculate previous and next page numbers
    prevPage := page - 1
    if prevPage < 1 {
        prevPage = 1
    }

    nextPage := page + 1
    if nextPage > totalPages {
        nextPage = totalPages
    }

    // Render the data into HTML rows and pagination controls
    c.HTML(http.StatusOK, "list.html", gin.H{
        "expenses":    result[0].Data,
        "currentPage": page,
        "totalPages":  totalPages,
        "prevPage":    prevPage,
        "nextPage":    nextPage,
    })
})
	router.Run(":9090")
}
