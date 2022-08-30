package main

import (
	"testing"
	"time"
)

// TestGetYearAreaCodeData description
//
// createTime: 2022-08-26 18:35:12
//
// author: hailaz
func TestGetYearAreaCodeData(t *testing.T) {
	now := time.Now()
	GetYearAreaCodeData(2021)
	t.Log(time.Since(now))
}
