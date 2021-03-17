package main

import "testing"

func TestCompositeUnitString(t *testing.T) {
	cu := CompositeUnit{
		UnitsList: []UnitExponent{
			{Unit: UnitTable["meter"], Exponent: 2},
			{Unit: UnitTable["second"], Exponent: -1},
		},
	}

	if cu.String() != "m^2 / s" {
		t.Errorf("Composite unit string should be m^2 / s, %s was returned instead", cu.String())
	}

	cu = CompositeUnit{
		UnitsList: []UnitExponent{
			{Unit: UnitTable["centimeter"], Exponent: 2},
		},
	}

	if cu.String() != "cm^2" {
		t.Errorf("Composite unit string should be cm^2, %s was returned instead", cu.String())
	}

	cu = CompositeUnit{
		UnitsList: []UnitExponent{
			{Unit: UnitTable["eur"], Exponent: 1},
			{Unit: UnitTable["meter"], Exponent: -2},
		},
	}

	if cu.String() != "€ / m^2" {
		t.Errorf("Composite unit string should be € / m^2, %s was returned instead", cu.String())
	}
}

func TestSort(t *testing.T) {
	cu := CompositeUnit{
		UnitsList: []UnitExponent{
			{Unit: UnitTable["second"], Exponent: -1},
			{Unit: UnitTable["meter"], Exponent: 2},
		},
	}

	got := cu.String()
	if got != "m^2 / s" {
		t.Errorf("Composite unit string sorted should be m^2 / s, %s was returned instead", got)
	}

	cu = CompositeUnit{
		UnitsList: []UnitExponent{
			{Unit: UnitTable["meter"], Exponent: 1},
			{Unit: UnitTable["eur"], Exponent: 1},
			{Unit: UnitTable["usd"], Exponent: 1},
		},
	}

	got = cu.String()
	if got != "$ m €" {
		t.Errorf("Composite unit string should be m € $, %s was returned instead", got)
	}
}

func TestFundamentalUnitConversion(t *testing.T) {
	got := ConvertFundamentalUnits(5, UnitTable["celsius"], UnitTable["fahrenheit"], 1)

	if got != 41 {
		t.Errorf("5 celsius should convert to 41 fahrenheit, got %f instead", got)
	}

	got = ConvertFundamentalUnits(5, UnitTable["kilometer"], UnitTable["millimeter"], 1)
	if got != 5_000_000 {
		t.Errorf("5 kilometers should convert to 5.000.000 millimeters, got %f instead", got)
	}

	got = ConvertFundamentalUnits(5, UnitTable["kilometer"], UnitTable["millimeter"], 1)
	if got != 5_000_000 {
		t.Errorf("5 kilometers should convert to 5.000.000 millimeters, got %f instead", got)
	}

	got = ConvertFundamentalUnits(1, UnitTable["month"], UnitTable["hour"], 1)
	if got != 720 {
		t.Errorf("1 month should convert to 720 hours, got %f instead", got)
	}

	got = ConvertFundamentalUnits(1, UnitTable["meter"], UnitTable["centimeter"], 2)
	if got != 10000 {
		t.Errorf("1 m^2 should convert to 10.000 cm^2, got %f instead", got)
	}
}
