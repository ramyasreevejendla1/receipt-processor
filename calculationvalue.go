package main

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type verifyable interface {
    verify() (valid bool)
}

type shortDescription string
func (v *shortDescription) verify() (vaild bool) {
    m, err := regexp.Match("^[\\w\\s\\-]+$", []byte(*v))
    if !m || err != nil { return }

    return true
}

type price string
func (v *price) verify() (vaild bool) {
    m, err := regexp.Match("^\\d+\\.\\d{2}$", []byte(*v))
    if !m || err != nil { return }

    return true
}

type Item struct  {
    ShortDescription shortDescription `json:"shortDescription"`
    Price price `json:"price"`
}

func (v *Item) verify() (valid bool) {
    return (v.ShortDescription.verify() &&
            v.Price.verify())
}

type retailer string
func (v *retailer) verify() (vaild bool) {
    m, err := regexp.Match("^[\\w\\s\\-&]+$", []byte(*v))
    if !m || err != nil { return }

    return true
}

type purchaseDate string
func (v *purchaseDate) verify() (vaild bool) {
    //date
    m, err := regexp.Match("^\\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])$", []byte(*v))
    if !m || err != nil { return }

    return true
}

type purchaseTime string
func (v *purchaseTime) verify() (vaild bool) {
    m, err := regexp.Match("^([01]\\d|2[0-3]):[0-5]\\d$", []byte(*v))
    if !m || err != nil { return }

    return true
}

type total string
func (v *total) verify() (vaild bool) {
    m, err := regexp.Match("^\\d+\\.\\d{2}$", []byte(*v))
    if !m || err != nil { return }

    return true
}

type itemArray []Item
func (arr *itemArray) verify() (vaild bool) {
    for _, v := range *arr {
        if !v.verify() { return }
    }

    return true
}

type Receipt struct {
    Retailer retailer `json:"retailer"`
    PurchaseDate purchaseDate `json:"purchaseDate"`
    PurchaseTime purchaseTime `json:"purchaseTime"`
    Items itemArray `json:"items"`
    Total total `json:"total"`
}

func (v *Receipt) verify() (valid bool) {
    return (v.Retailer.verify() &&
            v.PurchaseDate.verify() &&
            v.PurchaseTime.verify() &&
            v.Items.verify() &&
            v.Total.verify())
}


func (r *Receipt) calculatePoints() (points int) {
    //number of alphanumeric chars in retailer name
    for _, c := range r.Retailer {
        if unicode.IsLetter(c) || unicode.IsNumber(c) { points++ }
    }
    total, _ := strconv.ParseFloat(string(r.Total), 64) 
    //we can not check the error because we already check the format
    total100 := int(total * 100) //we multiply total with 100 and convert to int. its easier
    if (total100 % 100) == 0 {
        points += 50 //50 points if .00
    }
    if (total100 % 25) == 0 {
        points += 25 //25 points if divisible by .25
    }
    points += 5 * (int(len(r.Items) / 2)) //5 points for every 2 items on receipt

    for _, item := range r.Items { //if trimmed length of item description is 3 multiply price with .2, round up = points earned
        description := strings.TrimSpace(string(item.ShortDescription))
        if len(description) % 3 == 0 {
            price, _ := strconv.ParseFloat(string(item.Price), 64) //safe to ignore error because we checked already

            points += int(math.Ceil(0.2 * price))
        }
    }

    //if the PurchaseDate is odd, 6 points
    //just check if the last digit of the day (last char in string) is odd
    if int(r.PurchaseDate[len(r.PurchaseDate) - 1]) % 2 == 1 {
        points += 6
    }

    //10 points if time of purchase is after 2:00pm and before 4:00pm
    purchaseHour, _ := strconv.ParseInt(string(r.PurchaseTime[0:2]), 10, 64) //again ignore error
    if purchaseHour == 14 || purchaseHour == 15 {
        points += 10
    }

    return points
}
