# Personal Expense Tracker

This is a **Personal Expense Calculator** designed to help track spending while also experimenting with **Golang View** and **HTMX**. The project provides a simple web interface to list and manage expenses using **Gin (Go Web Framework)** and **MongoDB**.

## Features
- **List Expenses**: View paginated expense records.
- **Sorting**: Expenses are sorted by the latest `created_at` date.
- **Pagination**: Navigate between pages of expenses.
- **HTMX Integration**: Enhances user experience with dynamic updates.
- **MongoDB Aggregation**: Uses MongoDB's aggregation pipeline for efficient data retrieval.

## Technologies Used
- **Golang (Gin Framework)**: Backend API and view rendering.
- **MongoDB**: Database for storing expense records.
- **HTMX**: Enhances interactivity without full-page reloads.
- **HTML/CSS**: Simple frontend UI.

## Setup Instructions

### Prerequisites
Ensure you have the following installed:
- [Go](https://go.dev/) (1.18 or later recommended)
- [MongoDB](https://www.mongodb.com/)
- [Gin](https://github.com/gin-gonic/gin) package
- [mgm](https://github.com/Kamva/mgm) package for MongoDB management

### Installation & Running

1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/personal-expense-tracker.git
   cd personal-expense-tracker
   ```
2. **Install dependencies:**
   ```sh
   go mod tidy
   ```
3. **Set up MongoDB Connection:**
   Update your database connection details in the code if necessary.

4. **Run the server:**
   ```sh
   go run main.go
   ```
5. **Access the application:**
   Open `http://localhost:8080/list` in your browser.

## API Routes

| Method | Endpoint     | Description          |
|--------|-------------|----------------------|
| GET    | `/list`     | List expenses        |

## Example Query Parameters
You can use the following query parameters to paginate the expense list:
```
/list?page=2&limit=10
```

## Contribution
Feel free to fork this repository and submit pull requests for improvements!

## License
This project is licensed under the MIT License.

---
Happy coding! 🚀


