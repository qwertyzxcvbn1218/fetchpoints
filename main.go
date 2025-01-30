package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	startTime                = 14
	endTime                  = 16
	pointsForTime            = 10
	pointsForItemPairs       = 5
	pointsForQuarterMultiple = 25
	pointsForNoCents         = 50
	pointsForDescription     = 0.2
	pointsForDate            = 6
)

var (
	ReceiptCount int
	Points       int
)

var ReceiptPoints = make(map[int]int)

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total        string `json:"total"`
	Items        []Item `json:"items"`
}

type ReceiptID struct {
	ID int `json:"ID"`
}

type ReturnPoints struct {
	Points int `json:"points"`
}

// countAlphanumericLength returns just the sum of alphanumeric characters in a given string
func countAlphanumericLength(s string) int {
	length := 0
	for _, ch := range s {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			length++
		}
	}
	return length
}

// calcPointsTime returns the points based on the time condition (time after 14:00 and before 16:00)
func calcPointsTime(hour int, minutes int) int {
	timePoints := 0
	if (hour > startTime && hour < endTime) || (hour == startTime && minutes >= 0) || (hour == endTime && minutes == 0) {
		timePoints += pointsForTime
	}
	
	return timePoints
}

// getHoursMinutes returns the hours and minutes
func getHoursMinutes(hoursMins string) (int, int, error) {
	layout := "15:04"

	t, err := time.Parse(layout, hoursMins)
	if err != nil {
		return 0, 0, err
	}
	return t.Hour(), t.Minute(), nil
}

func parseDate(date string) (time.Time, error) {
	layout := "2006-01-02"

	t, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// calcDatePoints calculates points based on a given date
func calcDatePoints(date string) (int, error) {
	datePoints := 0

	d, err := parseDate(date)
	if err != nil {
		log.Printf("Error occurred while calulating point for date: %v", err)
	}
	_, _, day := d.Date()
	if day%2 == 1 {
		datePoints += pointsForDate
	}

	return datePoints, nil

}

// processReceipt is the handler that handles the POST request to /receipts/process endpoint
func processReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt

	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// incrementing the receipt count for every process request made
	ReceiptCount++

	rid := ReceiptID{ID: ReceiptCount}

	ReceiptPoints[ReceiptCount], err = calculatePoints(receipt)
	if err != nil {
		http.Error(w, "please check the date and time of the receipt provided", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rid)

}

// calculatePoints calculates the points accrued by a receipt based on the given rules.
func calculatePoints(receipt Receipt) (int, error) {

	totalPoints := 0
	var basePoints float64
	var lenDescriptionPoints int

	for _, item := range receipt.Items {
		floatVal, _ := strconv.ParseFloat(item.Price, 64)
		basePoints += floatVal
		trimmedDescription := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDescription)%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			descPoints := price * pointsForDescription
			roundedScore := math.Ceil(descPoints)
			lenDescriptionPoints += int(roundedScore)
		}
	}

	// points based on retailer
	retailerLen := countAlphanumericLength(receipt.Retailer)
	totalPoints += retailerLen

	// points for every 2 items in the receipt
	pairOfItems := len(receipt.Items) / 2
	totalPoints += pairOfItems * pointsForItemPairs

	// if the base points is a multiple of 0.25
	wholeNum := basePoints * 4
	if math.Mod(wholeNum, 1) == 0 {
		totalPoints += pointsForQuarterMultiple
	}

	// if there are no cents
	if basePoints == float64(int(basePoints)) {
		totalPoints += pointsForNoCents
	}

	// length of the description
	totalPoints += lenDescriptionPoints

	// points for date
	datePoints, _ := calcDatePoints(receipt.PurchaseDate)
	totalPoints += datePoints

	// points for time
	hours, minutes, err := getHoursMinutes(receipt.PurchaseTime)
	if err != nil {
		log.Printf("Error occurred while getting the hours and minutes: %v", err)
		return totalPoints, err
	}
	timePoints := calcPointsTime(hours, minutes)
	totalPoints += timePoints

	return totalPoints, nil
}

// getPoints is the handler for handling the GET request to /receipts/{id}/points endpoint
func getPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	receiptID := vars["id"]
	if len(receiptID) == 0 {
		http.Error(w, "no receipt id provided", http.StatusBadRequest)
		return
	}

	if len(ReceiptPoints) == 0 {
		http.Error(w, "no receipt with the given id exists", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	rt, err := strconv.Atoi(receiptID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if rt == 0 {
		http.Error(w, "invalid receipt id provided", http.StatusBadRequest)
		return
	}
	rp := ReturnPoints{Points: ReceiptPoints[rt]}
	err = json.NewEncoder(w).Encode(rp)
	if err != nil {
		log.Printf("Error occurred while bulding the response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/receipts/{id}/points", getPoints).Methods("GET")
	r.HandleFunc("/receipts/process", processReceipt).Methods("POST")

	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
