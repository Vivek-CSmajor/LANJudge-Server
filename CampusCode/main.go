package main

import (
	// "database/sql"
	// "log"
	// "fmt"
	"campuscode/judgeEngine"
	// "github.com/gin-gonic/gin"
	//"github.com/mattn/go-sqlite3"
	"net/http"
)

// func main(){
// 	db, err := sql.Open("sqlite3", "./contest.db") // driver, file path
// 	if err != nil {
// 		panic(err);
// 	}
// 	defer db.Close() // free up resources when main function ends
// 	createTable := `CREATE TABLE IF NOT EXISTS submissions (
// 		submission_id INTEGER PRIMARY KEY AUTOINCREMENT
// 		problem_id INTEGER,
// 		result TEXT
// 		);`
// 	_, err = db.Exec(createTable)
// 	if err != nil {
// 		panic(err);
// 	}
// 	log.Println("Db initialized");

// 	router := gin.Default()
// 	router.POST("/judge", judge)
	
// }

// func main() {

// 	// C++ code being submitted
// 	code := `
// #include <iostream>
// using namespace std;

// int main(){
//     int a,b;
//     cin >> a >> b;
//     cout << a + b;
// }
// `

// 	// Simulated submission
// 	sub := judgeEngine.Submission{
// 		ID:        1,
// 		ProblemID: 1,
// 		Code:      code,
// 	}

// 	// Expected output for input.txt
// 	expected := "5"

// 	fmt.Println("Submitting code to judge...")

// 	result, err := judgeEngine.JudgeSubmission(sub, expected)

// 	if err != nil {
// 		fmt.Println("Judge error:", err)
// 		return
// 	}

// 	fmt.Println("Judge result:", result)
// }

func main(){
	// Simulated submission
	judgeEngine.InitializeRunTime()
	judgeEngine.InitDB()
	judgeEngine.InitializeQueue()
	judgeEngine.StartWorkers(1)
	http.HandleFunc("/submit", judgeEngine.SubmitHandler)
	http.ListenAndServe(":8080", http.DefaultServeMux)
}