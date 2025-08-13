package controllers

import (
	"math"
	"time"
)


func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}


func round (num float64) int {
	return int(num + math.Copysign(0.5,num))
}

func toFixed(num float64, precision int) float64{
	output := math.Pow(10, float64(precision))

	return  float64(round(num*output)) / output
}