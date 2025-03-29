package main

import (
	"deanx3/expense-calculator/handlers"
	"io"
	"log"
	"net/http"
	"os"

	// "sort"
	// "strconv"
	// "time"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kamva/mgm/v3"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func initMongoDB() {
	err := godotenv.Load()
	if err != nil {
		// log.Fatal("Error loading .env file")
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

	router.POST("/submit", handlers.SubmitForm)

	router.GET("/list", handlers.Listing)

	router.GET("/dashboard", handlers.Dashboard)
	router.Run(":9090")
//
// 	router.GET("/api/suggestions", func(c *gin.Context) {
// 		query := c.DefaultQuery("expense_location", "") // Get query parameter
// 		if query == "" {
// 			c.HTML(http.StatusOK, "empty_suggestions.html", nil)
// 			return
// 		}
//
// 		var suggestions []Transection
// 		filter := bson.M{"expense_location": bson.M{"$regex": query, "$options": "i"}}
// 		err := mgm.Coll(&Transection{}).SimpleFind(&suggestions, filter)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
// 			return
// 		}
//
// 		locations := make([]string, len(suggestions))
// 		for i, s := range suggestions {
// 			locations[i] = s.ExpenseLocation
// 		}
//
// 		c.HTML(http.StatusOK, "suggestions.html", gin.H{
// 			"suggestions": locations,
// 		})
// 	})
//

//
// 	router.GET("/dashboard", func(c *gin.Context) {
// 		kpiData := getKPI()
// 		c.HTML(http.StatusOK, "dashboard.html", gin.H{
// 			"TotalMonthlySpending":      kpiData.TotalMonthlySpending,
// 			"TotalYearlySpending":       kpiData.TotalYearlySpending,
// 			"AverageDailySpending":      kpiData.AverageDailySpending,
// 			"PercentageIncomeSpent":     kpiData.PercentageIncomeSpent,
// 			"TotalAvailableMoney":       kpiData.TotalAvailableMoney,
// 			"TopExpensiveTransactions":  kpiData.TopExpensiveTransactions,
// 			"LatestTransactions":        kpiData.LatestTransactions,
// 			"SpendingTrendDaily":        kpiData.SpendingTrendDaily,
// 			"DailyDates":                kpiData.DailyDates,
// 			"TopSpendingCategories":     kpiData.TopSpendingCategories,
// 			"BalanceSourceNames":        kpiData.BalanceSourceNames,
// 			"BalanceAmounts":            kpiData.BalanceAmounts,
// 			"ExpensivePurchaseCategories": kpiData.ExpensivePurchaseCategories,
// 			"ExpensivePurchaseAmounts":  kpiData.ExpensivePurchaseAmounts,
// 		})
// 	})
//
// 	router.Run(":9090")
// }
//
// type CategoryData struct {
// 	Category string
// 	Amount   float64
// }
//
// // KPIData holds all the data to be passed to the dashboard.
// type KPIData struct {
// 	TotalMonthlySpending      float64
// 	TotalYearlySpending       float64
// 	AverageDailySpending      float64
// 	PercentageIncomeSpent     float64
// 	TotalAvailableMoney       float64
// 	TopExpensiveTransactions  []Transection
// 	LatestTransactions        []Transection
// 	SpendingTrendDaily        []float64
// 	DailyDates                []string
// 	TopSpendingCategories     []CategoryData
// 	BalanceSourceNames        []string
// 	BalanceAmounts            []float64
// 	ExpensivePurchaseCategories []string
// 	ExpensivePurchaseAmounts  []float64
// }
//
// func getKPI() KPIData {
// 	// Collections
// 	transactionCollection := mgm.Coll(&Transection{})
// 	balanceCollection := mgm.Coll(&Balance{})
//
// 	now := time.Now()
// 	currentMonth := now.Month()
// 	startOfMonth := time.Date(now.Year(), currentMonth, 1, 0, 0, 0, 0, time.Local)
// 	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
// 	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)
//
// 	// Retrieve monthly transactions excluding transfers
// 	monthlyTransactions := []Transection{}
// 	err := transactionCollection.SimpleFind(&monthlyTransactions, bson.M{
// 		"created_at":   bson.M{"$gte": startOfMonth, "$lte": endOfMonth},
// 		"expense_type": bson.M{"$ne": "transfer"},
// 	})
// 	if err != nil {
// 		log.Println("Error fetching monthly transactions:", err)
// 	}
//
// 	// Calculate Total Monthly Spending & category totals (exclude transfers)
// 	var totalMonthlySpending float64
// 	categoryTotals := make(map[string]float64)
// 	for _, t := range monthlyTransactions {
// 		if t.ExpenseType == "outbound" {
// 			totalMonthlySpending += t.Amount
// 			categoryTotals[t.Category] += t.Amount
// 		}
// 	}
//
// 	// Retrieve yearly transactions excluding transfers
// 	yearlyTransactions := []Transection{}
// 	err = transactionCollection.SimpleFind(&yearlyTransactions, bson.M{
// 		"created_at":   bson.M{"$gte": startOfYear},
// 		"expense_type": bson.M{"$ne": "transfer"},
// 	})
// 	if err != nil {
// 		log.Println("Error fetching yearly transactions:", err)
// 	}
// 	var totalYearlySpending float64
// 	for _, t := range yearlyTransactions {
// 		if t.ExpenseType == "outbound" {
// 			totalYearlySpending += t.Amount
// 		}
// 	}
//
// 	// Average Daily Spending for current month
// 	daysInMonth := startOfMonth.AddDate(0, 1, 0).Sub(startOfMonth).Hours() / 24
// 	averageDailySpending := totalMonthlySpending / daysInMonth
//
// 	// Fetch balance data for available money and balance chart
// 	balances := []Balance{}
// 	err = balanceCollection.SimpleFind(&balances, bson.M{})
// 	if err != nil {
// 		log.Println("Error fetching balances:", err)
// 	}
// 	var totalAvailableMoney float64
// 	var balanceSourceNames []string
// 	var balanceAmounts []float64
// 	var totalIncome float64
// 	for _, balance := range balances {
// 		totalIncome += balance.Balance
// 		totalAvailableMoney += balance.Balance
// 		balanceSourceNames = append(balanceSourceNames, balance.SourceName)
// 		balanceAmounts = append(balanceAmounts, balance.Balance)
// 	}
//
// 	// Calculate % of Income Spent (avoid division by zero)
// 	percentageOfIncomeSpent := 0.0
// 	if totalIncome != 0 {
// 		percentageOfIncomeSpent = (totalMonthlySpending / totalIncome) * 100
// 	}
//
// 	// Retrieve top expensive transactions for current month (exclude transfers)
// 	topExpensiveTransactions := []Transection{}
// 	err = transactionCollection.SimpleFind(&topExpensiveTransactions, bson.M{
// 		"created_at":   bson.M{"$gte": startOfMonth, "$lte": endOfMonth},
// 		"expense_type": bson.M{"$ne": "transfer"},
// 	})
// 	if err != nil {
// 		log.Println("Error fetching top transactions:", err)
// 	}
// 	sort.Slice(topExpensiveTransactions, func(i, j int) bool {
// 		return topExpensiveTransactions[i].Amount > topExpensiveTransactions[j].Amount
// 	})
// 	if len(topExpensiveTransactions) > 10 {
// 		topExpensiveTransactions = topExpensiveTransactions[:10]
// 	}
//
// 	// Prepare data for the Most Expensive Purchases pie chart
// 	var expensivePurchaseCategories []string
// 	var expensivePurchaseAmounts []float64
// 	for _, t := range topExpensiveTransactions {
// 		expensivePurchaseCategories = append(expensivePurchaseCategories, t.Category)
// 		expensivePurchaseAmounts = append(expensivePurchaseAmounts, t.Amount)
// 	}
//
// 	// Build Daily Spending Trend & Dates for current month
// 	var dailySpending []float64
// 	var dailyDates []string
// 	for day := 1; day <= int(daysInMonth); day++ {
// 		dayStart := time.Date(now.Year(), currentMonth, day, 0, 0, 0, 0, time.Local)
// 		dayEnd := dayStart.Add(24 * time.Hour).Add(-time.Second)
// 		var dailyTotal float64
// 		for _, t := range monthlyTransactions {
// 			if t.CreatedAt.After(dayStart) && t.CreatedAt.Before(dayEnd) {
// 				dailyTotal += t.Amount
// 			}
// 		}
// 		dailySpending = append(dailySpending, dailyTotal)
// 		dailyDates = append(dailyDates, dayStart.Format("2006-01-02"))
// 	}
//
// 	// Retrieve latest transactions (exclude transfers) and take top 20
// 	latestTransactions := []Transection{}
// 	err = transactionCollection.SimpleFind(&latestTransactions, bson.M{
// 		"expense_type": bson.M{"$ne": "transfer"},
// 	})
// 	if err != nil {
// 		log.Println("Error fetching latest transactions:", err)
// 	}
// 	sort.Slice(latestTransactions, func(i, j int) bool {
// 		return latestTransactions[i].CreatedAt.After(latestTransactions[j].CreatedAt)
// 	})
// 	if len(latestTransactions) > 20 {
// 		latestTransactions = latestTransactions[:20]
// 	}
//
// 	// Calculate Top 3 Spending Categories
// 	var topSpendingCategories []CategoryData
// 	for cat, amt := range categoryTotals {
// 		topSpendingCategories = append(topSpendingCategories, CategoryData{Category: cat, Amount: amt})
// 	}
// 	sort.Slice(topSpendingCategories, func(i, j int) bool {
// 		return topSpendingCategories[i].Amount > topSpendingCategories[j].Amount
// 	})
// 	if len(topSpendingCategories) > 3 {
// 		topSpendingCategories = topSpendingCategories[:3]
// 	}
//
// 	return KPIData{
// 		TotalMonthlySpending:      totalMonthlySpending,
// 		TotalYearlySpending:       totalYearlySpending,
// 		AverageDailySpending:      averageDailySpending,
// 		PercentageIncomeSpent:     percentageOfIncomeSpent,
// 		TotalAvailableMoney:       totalAvailableMoney,
// 		TopExpensiveTransactions:  topExpensiveTransactions,
// 		LatestTransactions:        latestTransactions,
// 		SpendingTrendDaily:        dailySpending,
// 		DailyDates:                dailyDates,
// 		TopSpendingCategories:     topSpendingCategories,
// 		BalanceSourceNames:        balanceSourceNames,
// 		BalanceAmounts:            balanceAmounts,
// 		ExpensivePurchaseCategories: expensivePurchaseCategories,
// 		ExpensivePurchaseAmounts:  expensivePurchaseAmounts,
// 	}
}
//
// 	router.GET("/api/suggestions", func(c *gin.Context) {
// 		query := c.DefaultQuery("expense_location", "") // Get query parameter
// 		if query == "" {
// 			c.HTML(http.StatusOK, "empty_suggestions.html", nil)
// 			return
// 		}
//
// 		var suggestions []Transection
// 		filter := bson.M{"expense_location": bson.M{"$regex": query, "$options": "i"}}
// 		err := mgm.Coll(&Transection{}).SimpleFind(&suggestions, filter)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
// 			return
// 		}
//
// 		locations := make([]string, len(suggestions))
// 		for i, s := range suggestions {
// 			locations[i] = s.ExpenseLocation
// 		}
//
// 		c.HTML(http.StatusOK, "suggestions.html", gin.H{
// 			"suggestions": locations,
// 		})
// 	})
//
// 	router.GET("/list", func(c *gin.Context) {
// 		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
// 		if err != nil || page < 1 {
// 			page = 1
// 		}
// 		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
// 		if err != nil || limit < 1 {
// 			limit = 10
// 		}
// 		skip := (page - 1) * limit
//
// 		pipeline := []interface{}{
// 			map[string]interface{}{
// 				"$facet": map[string]interface{}{
// 					"data": []interface{}{
// 						map[string]interface{}{"$sort": map[string]interface{}{"created_at": -1}},
// 						map[string]interface{}{"$skip": skip},
// 						map[string]interface{}{"$limit": limit},
// 					},
// 					"totalCount": []interface{}{
// 						map[string]interface{}{"$count": "total"},
// 					},
// 				},
// 			},
// 		}
//
// 		var result []struct {
// 			Data       []Transection `bson:"data"`
// 			TotalCount []struct {
// 				Total int `bson:"total"`
// 			} `bson:"totalCount"`
// 		}
//
// 		cursor, err := mgm.Coll(&Transection{}).Aggregate(mgm.Ctx(), pipeline)
// 		if err != nil {
// 			log.Println(err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expenses"})
// 			return
// 		}
// 		defer cursor.Close(mgm.Ctx())
//
// 		if err = cursor.All(mgm.Ctx(), &result); err != nil {
// 			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
// 			return
// 		}
//
// 		if len(result) == 0 {
// 			c.HTML(http.StatusOK, "list.html", gin.H{
// 				"expenses":    []Transection{},
// 				"currentPage": page,
// 				"totalPages":  0,
// 				"prevPage":    0,
// 				"nextPage":    0,
// 			})
// 			return
// 		}
//
// 		totalCount := 0
// 		if len(result[0].TotalCount) > 0 {
// 			totalCount = result[0].TotalCount[0].Total
// 		}
// 		totalPages := (totalCount + limit - 1) / limit
// 		prevPage := page - 1
// 		if prevPage < 1 {
// 			prevPage = 1
// 		}
// 		nextPage := page + 1
// 		if nextPage > totalPages {
// 			nextPage = totalPages
// 		}
//
// 		c.HTML(http.StatusOK, "list.html", gin.H{
// 			"expenses":    result[0].Data,
// 			"currentPage": page,
// 			"totalPages":  totalPages,
// 			"prevPage":    prevPage,
// 			"nextPage":    nextPage,
// 		})
// 	})
//
// 	router.GET("/dashboard", func(c *gin.Context) {
// 		kpiData := getKPI()
// 		c.HTML(http.StatusOK, "dashboard.html", gin.H{
// 			"TotalMonthlySpending":      kpiData.TotalMonthlySpending,
// 			"TotalYearlySpending":       kpiData.TotalYearlySpending,
// 			"AverageDailySpending":      kpiData.AverageDailySpending,
// 			"PercentageIncomeSpent":     kpiData.PercentageIncomeSpent,
// 			"TotalAvailableMoney":       kpiData.TotalAvailableMoney,
// 			"TopExpensiveTransactions":  kpiData.TopExpensiveTransactions,
// 			"LatestTransactions":        kpiData.LatestTransactions,
// 			"SpendingTrendDaily":        kpiData.SpendingTrendDaily,
// 			"DailyDates":                kpiData.DailyDates,
// 			"TopSpendingCategories":     kpiData.TopSpendingCategories,
// 			"BalanceSourceNames":        kpiData.BalanceSourceNames,
// 			"BalanceAmounts":            kpiData.BalanceAmounts,
// 			"ExpensivePurchaseCategories": kpiData.ExpensivePurchaseCategories,
// 			"ExpensivePurchaseAmounts":  kpiData.ExpensivePurchaseAmounts,
// 		})
// 	})
//
// }
//
// type CategoryData struct {
// 	Category string
// 	Amount   float64
// }
//
// // KPIData holds all the data to be passed to the dashboard.
// type KPIData struct {
// 	TotalMonthlySpending      float64
// 	TotalYearlySpending       float64
// 	AverageDailySpending      float64
// 	PercentageIncomeSpent     float64
// 	TotalAvailableMoney       float64
// 	TopExpensiveTransactions  []Transection
// 	LatestTransactions        []Transection
// 	SpendingTrendDaily        []float64
// 	DailyDates                []string
// 	TopSpendingCategories     []CategoryData
// 	BalanceSourceNames        []string
// 	BalanceAmounts            []float64
// 	ExpensivePurchaseCategories []string
// 	ExpensivePurchaseAmounts  []float64
// }
//
// func getKPI() KPIData {
// 	// Collections
// 	transactionCollection := mgm.Coll(&Transection{})
// 	balanceCollection := mgm.Coll(&Balance{})
//
// 	now := time.Now()
// 	currentMonth := now.Month()
// 	startOfMonth := time.Date(now.Year(), currentMonth, 1, 0, 0, 0, 0, time.Local)
// 	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
// 	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)
//
// 	// Retrieve monthly transactions excluding transfers
// 	monthlyTransactions := []Transection{}
// 	err := transactionCollection.SimpleFind(&monthlyTransactions, bson.M{
// 		"created_at":   bson.M{"$gte": startOfMonth, "$lte": endOfMonth},
// 		"expense_type": bson.M{"$ne": "transfer"},
// 	})
// 	if err != nil {
// 		log.Println("Error fetching monthly transactions:", err)
// 	}
//
// 	// Calculate Total Monthly Spending & category totals (exclude transfers)
// 	var totalMonthlySpending float64
// 	categoryTotals := make(map[string]float64)
// 	for _, t := range monthlyTransactions {
// 		if t.ExpenseType == "outbound" {
// 			totalMonthlySpending += t.Amount
// 			categoryTotals[t.Category] += t.Amount
// 		}
// 	}
//
// 	// Retrieve yearly transactions excluding transfers
// 	yearlyTransactions := []Transection{}
// 	err = transactionCollection.SimpleFind(&yearlyTransactions, bson.M{
// 		"created_at":   bson.M{"$gte": startOfYear},
// 		"expense_type": bson.M{"$ne": "transfer"},
// 	})
// 	if err != nil {
// 		log.Println("Error fetching yearly transactions:", err)
// 	}
// 	var totalYearlySpending float64
// 	for _, t := range yearlyTransactions {
// 		if t.ExpenseType == "outbound" {
// 			totalYearlySpending += t.Amount
// 		}
// 	}
//
// 	// Average Daily Spending for current month
// 	daysInMonth := startOfMonth.AddDate(0, 1, 0).Sub(startOfMonth).Hours() / 24
// 	averageDailySpending := totalMonthlySpending / daysInMonth
//
// 	// Fetch balance data for available money and balance chart
// 	balances := []Balance{}
// 	err = balanceCollection.SimpleFind(&balances, bson.M{})
// 	if err != nil {
// 		log.Println("Error fetching balances:", err)
// 	}
// 	var totalAvailableMoney float64
// 	var balanceSourceNames []string
// 	var balanceAmounts []float64
// 	var totalIncome float64
// 	for _, balance := range balances {
// 		totalIncome += balance.Balance
// 		totalAvailableMoney += balance.Balance
// 		balanceSourceNames = append(balanceSourceNames, balance.SourceName)
// 		balanceAmounts = append(balanceAmounts, balance.Balance)
// 	}
//
// 	// Calculate % of Income Spent (avoid division by zero)
// 	percentageOfIncomeSpent := 0.0
// 	if totalIncome != 0 {
// 		percentageOfIncomeSpent = (totalMonthlySpending / totalIncome) * 100
// 	}
//
// 	// Retrieve top expensive transactions for current month (exclude transfers)
// 	topExpensiveTransactions := []Transection{}
// 	err = transactionCollection.SimpleFind(&topExpensiveTransactions, bson.M{
// 		"created_at":   bson.M{"$gte": startOfMonth, "$lte": endOfMonth},
// 		"expense_type": bson.M{"$ne": "transfer"},
// 	})
// 	if err != nil {
// 		log.Println("Error fetching top transactions:", err)
// 	}
// 	sort.Slice(topExpensiveTransactions, func(i, j int) bool {
// 		return topExpensiveTransactions[i].Amount > topExpensiveTransactions[j].Amount
// 	})
// 	if len(topExpensiveTransactions) > 10 {
// 		topExpensiveTransactions = topExpensiveTransactions[:10]
// 	}
//
// 	// Prepare data for the Most Expensive Purchases pie chart
// 	var expensivePurchaseCategories []string
// 	var expensivePurchaseAmounts []float64
// 	for _, t := range topExpensiveTransactions {
// 		expensivePurchaseCategories = append(expensivePurchaseCategories, t.Category)
// 		expensivePurchaseAmounts = append(expensivePurchaseAmounts, t.Amount)
// 	}
//
// 	// Build Daily Spending Trend & Dates for current month
// 	var dailySpending []float64
// 	var dailyDates []string
// 	for day := 1; day <= int(daysInMonth); day++ {
// 		dayStart := time.Date(now.Year(), currentMonth, day, 0, 0, 0, 0, time.Local)
// 		dayEnd := dayStart.Add(24 * time.Hour).Add(-time.Second)
// 		var dailyTotal float64
// 		for _, t := range monthlyTransactions {
// 			if t.CreatedAt.After(dayStart) && t.CreatedAt.Before(dayEnd) {
// 				dailyTotal += t.Amount
// 			}
// 		}
// 		dailySpending = append(dailySpending, dailyTotal)
// 		dailyDates = append(dailyDates, dayStart.Format("2006-01-02"))
// 	}
//
// 	// Retrieve latest transactions (exclude transfers) and take top 20
// 	latestTransactions := []Transection{}
// 	err = transactionCollection.SimpleFind(&latestTransactions, bson.M{
// 		"expense_type": bson.M{"$ne": "transfer"},
// 	})
// 	if err != nil {
// 		log.Println("Error fetching latest transactions:", err)
// 	}
// 	sort.Slice(latestTransactions, func(i, j int) bool {
// 		return latestTransactions[i].CreatedAt.After(latestTransactions[j].CreatedAt)
// 	})
// 	if len(latestTransactions) > 20 {
// 		latestTransactions = latestTransactions[:20]
// 	}
//
// 	// Calculate Top 3 Spending Categories
// 	var topSpendingCategories []CategoryData
// 	for cat, amt := range categoryTotals {
// 		topSpendingCategories = append(topSpendingCategories, CategoryData{Category: cat, Amount: amt})
// 	}
// 	sort.Slice(topSpendingCategories, func(i, j int) bool {
// 		return topSpendingCategories[i].Amount > topSpendingCategories[j].Amount
// 	})
// 	if len(topSpendingCategories) > 3 {
// 		topSpendingCategories = topSpendingCategories[:3]
// 	}
//
// 	return KPIData{
// 		TotalMonthlySpending:      totalMonthlySpending,
// 		TotalYearlySpending:       totalYearlySpending,
// 		AverageDailySpending:      averageDailySpending,
// 		PercentageIncomeSpent:     percentageOfIncomeSpent,
// 		TotalAvailableMoney:       totalAvailableMoney,
// 		TopExpensiveTransactions:  topExpensiveTransactions,
// 		LatestTransactions:        latestTransactions,
// 		SpendingTrendDaily:        dailySpending,
// 		DailyDates:                dailyDates,
// 		TopSpendingCategories:     topSpendingCategories,
// 		BalanceSourceNames:        balanceSourceNames,
// 		BalanceAmounts:            balanceAmounts,
// 		ExpensivePurchaseCategories: expensivePurchaseCategories,
// 		ExpensivePurchaseAmounts:  expensivePurchaseAmounts,
// 	}
// }
