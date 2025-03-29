package handlers

import (
	"deanx3/expense-calculator/models"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
)

func SubmitForm(ctx *gin.Context) {
	var formData models.ExpenseRequest

	// Bind form data
	if err := ctx.ShouldBind(&formData); err != nil {
		ctx.HTML(http.StatusOK, "form.html", gin.H{"success": false, "error": err.Error()})
		return
	}

	if formData.ExpenseType == "inbound" || formData.ExpenseType == "outbound" {
		sourceName := formData.BankName
		var balanceRecord models.Balance
		balanceRecord.FetchOrCreate(sourceName)

		if formData.ExpenseType == "inbound" {
			balanceRecord.Balance = formData.Amount + balanceRecord.Balance
			fmt.Println(balanceRecord.Balance, formData.Amount)
		} else {
			//CHeck balance
			if balanceRecord.Balance < formData.Amount {
				ctx.HTML(http.StatusOK, "form.html", gin.H{"success": false, "error": "Insufficient balance"})
				return
			}

			balanceRecord.Balance -= formData.Amount

		}
			err := mgm.Coll(&balanceRecord).Update(&balanceRecord)
			ctx.HTML(http.StatusOK, "form.html", gin.H{"success": false, "error": err.Error()})
	}

	if formData.ExpenseType == "transfer" {
		var fromBalance models.Balance
		fromBalance.FetchOrCreate(formData.FromBank)

		var toBalance models.Balance
		toBalance.FetchOrCreate(formData.ToBank)

		fromBalance.Balance -= formData.Amount
		toBalance.Balance += formData.Amount

		_ = mgm.Coll(&fromBalance).Update(&fromBalance)
		_ = mgm.Coll(&toBalance).Update(&toBalance)

	}

	transaction := models.Transection{
		ExpenseLocation: formData.ExpenseLocation,
		Category:        formData.Category,
		Amount:          formData.Amount,
		ExpenseType:     formData.ExpenseType,
		Description:     formData.Description,
		SourceName:      formData.BankName,
	}

	if err := mgm.Coll(&transaction).Create(&transaction); err != nil {
		ctx.HTML(http.StatusOK, "form.html", gin.H{"success": false, "errors": err.Error()})
		return
	}

	ctx.Redirect(http.StatusMovedPermanently, "/list")

}

func Listing(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	skip := (page - 1) * limit

	pipeline := []interface{}{
		map[string]interface{}{
			"$facet": map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{"$sort": map[string]interface{}{"created_at": -1}},
					map[string]interface{}{"$skip": skip},
					map[string]interface{}{"$limit": limit},
				},
				"totalCount": []interface{}{
					map[string]interface{}{"$count": "total"},
				},
			},
		},
	}

	var result []struct {
		Data       []models.Transection `bson:"data"`
		TotalCount []struct {
			Total int `bson:"total"`
		} `bson:"totalCount"`
	}

	cursor, err := mgm.Coll(&models.Transection{}).Aggregate(mgm.Ctx(), pipeline)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve expenses"})
		return
	}
	defer cursor.Close(mgm.Ctx())

	if err = cursor.All(mgm.Ctx(), &result); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	if len(result) == 0 {
		c.HTML(http.StatusOK, "list.html", gin.H{
			"expenses":    []models.Transection{},
			"currentPage": page,
			"totalPages":  0,
			"prevPage":    0,
			"nextPage":    0,
		})
		return
	}

	totalCount := 0
	if len(result[0].TotalCount) > 0 {
		totalCount = result[0].TotalCount[0].Total
	}
	totalPages := (totalCount + limit - 1) / limit
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = totalPages
	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"expenses":    result[0].Data,
		"currentPage": page,
		"totalPages":  totalPages,
		"prevPage":    prevPage,
		"nextPage":    nextPage,
	})
}

func Dashboard(ctx *gin.Context) {
	// Collections
	transactionCollection := mgm.Coll(&models.Transection{})
	balanceCollection := mgm.Coll(&models.Balance{})

	now := time.Now()
	currentMonth := now.Month()
	startOfMonth := time.Date(now.Year(), currentMonth, 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)

	// Retrieve monthly transactions excluding transfers
	monthlyTransactions := []models.Transection{}
	err := transactionCollection.SimpleFind(&monthlyTransactions, bson.M{
		"created_at":   bson.M{"$gte": startOfMonth, "$lte": endOfMonth},
		"expense_type": bson.M{"$ne": "transfer"},
	})
	if err != nil {
		fmt.Println("Error fetching monthly transactions:", err)
	}

	// Calculate Total Monthly Spending & category totals (exclude transfers)
	var totalMonthlySpending float64
	categoryTotals := make(map[string]float64)
	for _, t := range monthlyTransactions {
		if t.ExpenseType == "outbound" {
			totalMonthlySpending += t.Amount
			categoryTotals[t.Category] += t.Amount
		}
	}

	// Retrieve yearly transactions excluding transfers
	yearlyTransactions := []models.Transection{}
	err = transactionCollection.SimpleFind(&yearlyTransactions, bson.M{
		"created_at":   bson.M{"$gte": startOfYear},
		"expense_type": bson.M{"$ne": "transfer"},
	})
	if err != nil {
		fmt.Println("Error fetching yearly transactions:", err)
	}

	var totalYearlySpending float64
	for _, t := range yearlyTransactions {
		if t.ExpenseType == "outbound" {
			totalYearlySpending += t.Amount
		}
	}

	// Average Daily Spending for current month
	daysInMonth := startOfMonth.AddDate(0, 1, 0).Sub(startOfMonth).Hours() / 24
	averageDailySpending := totalMonthlySpending / daysInMonth

	// Fetch balance data for available money and balance chart
	balances := []models.Balance{}
	err = balanceCollection.SimpleFind(&balances, bson.M{})
	if err != nil {
		fmt.Println("Error fetching balances:", err)
	}

	var totalAvailableMoney float64
	var balanceSourceNames []string
	var balanceAmounts []float64
	var totalIncome float64
	for _, balance := range balances {
		totalIncome += balance.Balance
		totalAvailableMoney += balance.Balance
		balanceSourceNames = append(balanceSourceNames, balance.SourceName)
		balanceAmounts = append(balanceAmounts, balance.Balance)
	}

	// Calculate % of Income Spent (avoid division by zero)
	percentageOfIncomeSpent := 0.0
	if totalIncome != 0 {
		percentageOfIncomeSpent = (totalMonthlySpending / totalIncome) * 100
	}

	// Retrieve top expensive transactions for current month (exclude transfers)
	topExpensiveTransactions := []models.Transection{}
	err = transactionCollection.SimpleFind(&topExpensiveTransactions, bson.M{
		"created_at":   bson.M{"$gte": startOfMonth, "$lte": endOfMonth},
		"expense_type": bson.M{"$ne": "transfer"},
	})
	if err != nil {
		fmt.Println("Error fetching top transactions:", err)
	}

	sort.Slice(topExpensiveTransactions, func(i, j int) bool {
		return topExpensiveTransactions[i].Amount > topExpensiveTransactions[j].Amount
	})
	if len(topExpensiveTransactions) > 10 {
		topExpensiveTransactions = topExpensiveTransactions[:10]
	}

	// Prepare data for the Most Expensive Purchases pie chart
	var expensivePurchaseCategories []string
	var expensivePurchaseAmounts []float64
	for _, t := range topExpensiveTransactions {
		expensivePurchaseCategories = append(expensivePurchaseCategories, t.Category)
		expensivePurchaseAmounts = append(expensivePurchaseAmounts, t.Amount)
	}

	// Build Daily Spending Trend & Dates for current month
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

	// Retrieve latest transactions (exclude transfers) and take top 20
	latestTransactions := []models.Transection{}
	err = transactionCollection.SimpleFind(&latestTransactions, bson.M{
		"expense_type": bson.M{"$ne": "transfer"},
	})
	if err != nil {
		fmt.Println("Error fetching latest transactions:", err)
	}
	sort.Slice(latestTransactions, func(i, j int) bool {
		return latestTransactions[i].CreatedAt.After(latestTransactions[j].CreatedAt)
	})
	if len(latestTransactions) > 20 {
		latestTransactions = latestTransactions[:20]
	}

	// Calculate Top 3 Spending Categories
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

	ctx.HTML(http.StatusOK, "dashboard.html", gin.H{
		"TotalMonthlySpending":        totalMonthlySpending,
		"TotalYearlySpending":         totalYearlySpending,
		"AverageDailySpending":        averageDailySpending,
		"PercentageIncomeSpent":       percentageOfIncomeSpent,
		"TotalAvailableMoney":         totalAvailableMoney,
		"TopExpensiveTransactions":    topExpensiveTransactions,
		"LatestTransactions":          latestTransactions,
		"SpendingTrendDaily":          dailySpending,
		"DailyDates":                  dailyDates,
		"TopSpendingCategories":       topSpendingCategories,
		"BalanceSourceNames":          balanceSourceNames,
		"BalanceAmounts":              balanceAmounts,
		"ExpensivePurchaseCategories": expensivePurchaseCategories,
		"ExpensivePurchaseAmounts":    expensivePurchaseAmounts,
	})
	return
}

type CategoryData struct {
	Category string
	Amount   float64
}

// KPIData holds all the data to be passed to the dashboard.
type KPIData struct {
	TotalMonthlySpending        float64
	TotalYearlySpending         float64
	AverageDailySpending        float64
	PercentageIncomeSpent       float64
	TotalAvailableMoney         float64
	TopExpensiveTransactions    []models.Transection
	LatestTransactions          []models.Transection
	SpendingTrendDaily          []float64
	DailyDates                  []string
	TopSpendingCategories       []CategoryData
	BalanceSourceNames          []string
	BalanceAmounts              []float64
	ExpensivePurchaseCategories []string
	ExpensivePurchaseAmounts    []float64
}
