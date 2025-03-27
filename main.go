package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Transection struct {
	mgm.DefaultModel `bson:",inline"`
	ExpenseLocation  string  `json:"expense_location" bson:"expense_location" form:"expense_location"`
	Category         string  `json:"category" bson:"category" form:"category"`
	Amount           float64 `json:"amount" bson:"amount" form:"amount"`
	ExpenseType      string  `json:"expense_type" bson:"expense_type"`
	Description      string  `json:"description" bson:"description" form:"description"`
	SourceName       string  `json:"source_name" bson:"source_name"`
}

type Balance struct {
	mgm.DefaultModel `bson:",inline"`
	Balance          float64 `json:"balance" bson:"balance" `
	SourceName       string  `json:"source_name" bson:"source_name"`
}

type ExpenseRequest struct {
	ExpenseType     string  `json:"expense_type" form:"expense_type"` // inbound, outbound, transfer
	ExpenseLocation string  `json:"expense_location,omitempty" form:"expense_location,omitempty"`
	Category        string  `json:"category,omitempty" form:"category,omitempty"`
	Amount          float64 `json:"amount" form:"amount"`
	Description     string  `json:"description,omitempty" form:"description,omitempty"`
	BankName        string  `json:"bank_name,omitempty" form:"bank_name,omitempty"` // For inbound & transfer
	FromBank        string  `json:"from_bank,omitempty" form:"from_bank,omitempty"` // For transfer
	ToBank          string  `json:"to_bank,omitempty" form:"to_bank,omitempty"`     // For transfer
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
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	router := gin.Default()

	// Set up template rendering
	router.HTMLRender = ginview.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title})
	})

	router.POST("/submit", func(ctx *gin.Context) {
		var formData ExpenseRequest

		// Bind form data
		if err := ctx.ShouldBind(&formData); err != nil {
			ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": err.Error()})
			return
		}

		var balanceRecord Balance
		err := mgm.Coll(&balanceRecord).First(bson.M{"source_name": formData.BankName}, &balanceRecord)
		if err != nil && err.Error() != "mongo: no documents in result" {
			ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": err.Error()})
			return
		}

		if err != nil {
			// If no balance record exists, initialize it
			if err.Error() == "mongo: no documents in result" {
				balanceRecord = Balance{
					Balance:    0,
					SourceName: formData.BankName,
				}
				_ = mgm.Coll(&balanceRecord).Create(&balanceRecord)
			}
		}

		switch formData.ExpenseType {
		case "inbound":
			balanceRecord.Balance = formData.Amount + balanceRecord.Balance

		case "outbound":
			if balanceRecord.Balance < formData.Amount {
				ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": "Insufficient balance"})
				return
			}
			balanceRecord.Balance -= formData.Amount

		case "transfer":
			var fromBalance Balance
			err := mgm.Coll(&fromBalance).First(bson.M{"source_name": formData.FromBank}, &fromBalance)
			if err != nil || fromBalance.Balance < formData.Amount {
				ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": "Insufficient balance in source account"})
				return
			}
			if err != nil && err.Error() == "mongo: no documents in result" {
				fromBalance = Balance{
					Balance:    0,
					SourceName: formData.FromBank,
				}
				_ = mgm.Coll(&fromBalance).Create(&fromBalance)
			}

			var toBalance Balance
			err = mgm.Coll(&toBalance).First(bson.M{"source_name": formData.ToBank}, &toBalance)
			if err != nil && err.Error() != "mongo: no documents in result" {
				ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": err.Error()})
				return
			}

			if err != nil && err.Error() == "mongo: no documents in result" {
				toBalance = Balance{
					Balance:    0,
					SourceName: formData.ToBank,
				}
				_ = mgm.Coll(&toBalance).Create(&toBalance)
			}

			fromBalance.Balance -= formData.Amount
			toBalance.Balance += formData.Amount

			_ = mgm.Coll(&fromBalance).Update(&fromBalance)
			_ = mgm.Coll(&toBalance).Update(&toBalance)
		}

		_ = mgm.Coll(&balanceRecord).Update(&balanceRecord)

		transaction := Transection{
			ExpenseLocation: formData.ExpenseLocation,
			Category:        formData.Category,
			Amount:          formData.Amount,
			ExpenseType:     formData.ExpenseType,
			Description:     formData.Description,
			SourceName:      formData.BankName,
		}

		if err := mgm.Coll(&transaction).Create(&transaction); err != nil {
			ctx.HTML(http.StatusOK, "form.html", gin.H{"title": title, "success": false, "error": err.Error()})
			return
		}

		ctx.Redirect(http.StatusMovedPermanently, "/list")
	})

	router.GET("/api/suggestions", func(c *gin.Context) {
		query := c.DefaultQuery("expense_location", "") // Get query parameter 'q'
		if query == "" {
			// If query is empty, return an empty suggestions list
			c.HTML(http.StatusOK, "empty_suggestions.html", nil)
			return
		}

		// Define a slice to hold the suggestions
		var suggestions []Transection

		// Build the filter to search for expense_location using a regex query
		filter := bson.M{"expense_location": bson.M{"$regex": query, "$options": "i"}}

		// Fetch data from MongoDB using mgm
		err := mgm.Coll(&Transection{}).SimpleFind(&suggestions, filter)
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
			Data       []Transection `bson:"data"`
			TotalCount []struct {
				Total int `bson:"total"`
			} `bson:"totalCount"`
		}

		// Run aggregation on the collection
		cursor, err := mgm.Coll(&Transection{}).Aggregate(mgm.Ctx(), pipeline)
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
				"expenses":    []Transection{},
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

	router.GET("/dashboard", func(c *gin.Context) {
		kpiData := getKPI()
		c.HTML(http.StatusOK, "dashboard.html", gin.H{
			"TotalMonthlySpending":     kpiData.TotalMonthlySpending,
			"TotalYearlySpending":      kpiData.TotalYearlySpending,
			"AverageDailySpending":     kpiData.AverageDailySpending,
			"PercentageIncomeSpent":    kpiData.PercentageIncomeSpent,
			"TotalAvailableMoney":      kpiData.TotalAvailableMoney,
			"TopExpensiveTransactions": kpiData.TopExpensiveTransactions,
			"LatestTransactions":       kpiData.LatestTransactions,
			"SpendingTrendDaily":       kpiData.SpendingTrendDaily,
			"DailyDates":               kpiData.DailyDates,
			"TopSpendingCategories":    kpiData.TopSpendingCategories,
		})
	})

	router.Run(":9090")
}

type CategoryData struct {
	Category string
	Amount   float64
}

// KPIData holds all the data to be passed to the dashboard.
type KPIData struct {
	TotalMonthlySpending     float64
	TotalYearlySpending      float64
	AverageDailySpending     float64
	PercentageIncomeSpent    float64
	TotalAvailableMoney      float64
	TopExpensiveTransactions []Transection
	LatestTransactions       []Transection
	SpendingTrendDaily       []float64
	DailyDates               []string
	TopSpendingCategories    []CategoryData
}

func getKPI() KPIData {
	// Collections
	transactionCollection := mgm.Coll(&Transection{})
	balanceCollection := mgm.Coll(&Balance{})

	// Get current time and calculate start and end of month/year
	now := time.Now()
	currentMonth := now.Month()
	startOfMonth := time.Date(now.Year(), currentMonth, 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)

	// Retrieve monthly transactions
	monthlyTransactions := []Transection{}
	err := transactionCollection.SimpleFind(&monthlyTransactions, bson.M{
		"created_at": bson.M{"$gte": startOfMonth, "$lte": endOfMonth},
	})
	if err != nil {
		log.Println("Error fetching monthly transactions:", err)
	}

	// Calculate Total Monthly Spending & build category totals
	var totalMonthlySpending float64
	categoryTotals := make(map[string]float64)
	for _, t := range monthlyTransactions {
		if t.ExpenseType == "outbound" {
			totalMonthlySpending += t.Amount
			categoryTotals[t.Category] += t.Amount
		}
	}

	// Calculate Total Yearly Spending
	yearlyTransactions := []Transection{}
	err = transactionCollection.SimpleFind(&yearlyTransactions, bson.M{
		"created_at": bson.M{"$gte": startOfYear},
	})
	if err != nil {
		log.Println("Error fetching yearly transactions:", err)
	}
	var totalYearlySpending float64
	for _, t := range yearlyTransactions {
		if t.ExpenseType == "outbound" {
			totalYearlySpending += t.Amount
		}
	}

	// Calculate Average Daily Spending
	daysInMonth := startOfMonth.AddDate(0, 1, 0).Sub(startOfMonth).Hours() / 24
	averageDailySpending := totalMonthlySpending / daysInMonth

	// Fetch Balance data (assuming all balances represent income/available money)
	balances := []Balance{}
	err = balanceCollection.SimpleFind(&balances, bson.M{})
	if err != nil {
		log.Println("Error fetching balances:", err)
	}
	var totalIncome float64
	var totalAvailableMoney float64
	for _, balance := range balances {
		totalIncome += balance.Balance
		totalAvailableMoney += balance.Balance
	}

	// Calculate % of Income Spent (guard against division by zero)
	percentageOfIncomeSpent := 0.0
	if totalIncome != 0 {
		percentageOfIncomeSpent = (totalMonthlySpending / totalIncome) * 100
	}

	// Retrieve and sort Top Expensive Transactions for current month (descending order)
	topExpensiveTransactions := []Transection{}
	err = transactionCollection.SimpleFind(&topExpensiveTransactions, bson.M{
		"created_at": bson.M{"$gte": startOfMonth, "$lte": endOfMonth},
	})
	if err != nil {
		log.Println("Error fetching top transactions:", err)
	}
	sort.Slice(topExpensiveTransactions, func(i, j int) bool {
		return topExpensiveTransactions[i].Amount > topExpensiveTransactions[j].Amount
	})
	if len(topExpensiveTransactions) > 10 {
		topExpensiveTransactions = topExpensiveTransactions[:10]
	}

	// Build Daily Spending Trend & corresponding Dates for current month
	var dailySpending []float64
	var dailyDates []string
	for day := 1; day <= int(daysInMonth); day++ {
		dayStart := time.Date(now.Year(), currentMonth, day, 0, 0, 0, 0, time.Local)
		dayEnd := dayStart.Add(24 * time.Hour).Add(-time.Second)
		var dailyTotal float64
		for _, t := range monthlyTransactions {
			if t.CreatedAt.After(dayStart) && t.CreatedAt.Before(dayEnd) {
				dailyTotal += t.Amount
			}
		}
		dailySpending = append(dailySpending, dailyTotal)
		dailyDates = append(dailyDates, dayStart.Format("2006-01-02"))
	}

	// Retrieve and sort Latest Transactions (descending by CreatedAt) and take top 20
	latestTransactions := []Transection{}
	err = transactionCollection.SimpleFind(&latestTransactions, bson.M{})
	if err != nil {
		log.Println("Error fetching latest transactions:", err)
	}
	sort.Slice(latestTransactions, func(i, j int) bool {
		return latestTransactions[i].CreatedAt.After(latestTransactions[j].CreatedAt)
	})
	if len(latestTransactions) > 20 {
		latestTransactions = latestTransactions[:20]
	}

	// Calculate Top 3 Spending Categories from monthly data
	var topSpendingCategories []CategoryData
	for cat, amt := range categoryTotals {
		topSpendingCategories = append(topSpendingCategories, CategoryData{Category: cat, Amount: amt})
	}
	sort.Slice(topSpendingCategories, func(i, j int) bool {
		return topSpendingCategories[i].Amount > topSpendingCategories[j].Amount
	})
	if len(topSpendingCategories) > 3 {
		topSpendingCategories = topSpendingCategories[:3]
	}

	return KPIData{
		TotalMonthlySpending:     totalMonthlySpending,
		TotalYearlySpending:      totalYearlySpending,
		AverageDailySpending:     averageDailySpending,
		PercentageIncomeSpent:    percentageOfIncomeSpent,
		TotalAvailableMoney:      totalAvailableMoney,
		TopExpensiveTransactions: topExpensiveTransactions,
		LatestTransactions:       latestTransactions,
		SpendingTrendDaily:       dailySpending,
		DailyDates:               dailyDates,
		TopSpendingCategories:    topSpendingCategories,
	}
}
