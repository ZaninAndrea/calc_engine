package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

type FundamentalUnit struct {
	ID               string
	DisplayValue     string
	Aliases          []string
	BaseUnit         string
	ConversionFactor float64
	ConversionShift  float64
}

var UnitAliasesMap map[string]string = map[string]string{}
var UnitTable map[string]FundamentalUnit = map[string]FundamentalUnit{
	// metric lengths
	"meter":      {"meter", "m", []string{"m", "meter", "metre"}, "meter", 1, 0},
	"decimeter":  {"decimeter", "dm", []string{"dm", "decimetre", "decimeter"}, "meter", math.Pow10(-1), 0},
	"centimeter": {"centimeter", "cm", []string{"cm", "centimetre", "centimeter"}, "meter", math.Pow10(-2), 0},
	"millimeter": {"millimeter", "mm", []string{"mm", "millimetre", "millimeter"}, "meter", math.Pow10(-3), 0},
	"micrometer": {"micrometer", "μm", []string{"μm", "micrometre", "micrometer"}, "meter", math.Pow10(-6), 0},
	"nanometer":  {"nanometer", "nm", []string{"nm", "nanometre", "nanometer"}, "meter", math.Pow10(-9), 0},
	"picometer":  {"picometer", "pm", []string{"pm", "picometre", "picometer"}, "meter", math.Pow10(-12), 0},
	"femtometer": {"femtometer", "fm", []string{"fm", "femtometre", "femtometer"}, "meter", math.Pow10(-15), 0},
	"decameter":  {"decameter", "dam", []string{"dam", "decametre", "decameter"}, "meter", math.Pow10(1), 0},
	"hectometer": {"hectometer", "hm", []string{"hm", "hectometre", "hectometer"}, "meter", math.Pow10(2), 0},
	"kilometer":  {"kilometer", "km", []string{"km", "kilometre", "kilometer"}, "meter", math.Pow10(3), 0},
	"megameter":  {"megameter", "Mm", []string{"Mm", "megametre", "megameter"}, "meter", math.Pow10(6), 0},
	"gigameter":  {"gigameter", "Gm", []string{"Gm", "gigametre", "gigameter"}, "meter", math.Pow10(9), 0},
	// imperial lengths
	"inch":          {"inch", "in", []string{"in", "inch", "inches"}, "meter", 0.0254, 0},
	"foot":          {"foot", "ft", []string{"ft", "foot", "feet"}, "meter", 0.3048, 0},
	"yard":          {"yard", "yd", []string{"yd", "yard", "yards"}, "meter", 0.9144, 0},
	"mile":          {"mile", "mi", []string{"mi", "mile", "miles"}, "meter", 1609.344, 0},
	"nautical_mile": {"nautical_mile", "nmi", []string{"nmi", "nautical_mile"}, "meter", 1852, 0},
	// astronomical lengths
	"lunar_distance":    {"lunar_distance", "ld", []string{"ld", "lunar_distance", "lunar_distances"}, "meter", 384_402_000, 0},
	"astronomical_unit": {"astronomical_unit", "au", []string{"au", "astronomical_unit", "astronomical_units"}, "meter", 149_597_870_700, 0},
	"light_year":        {"light_year", "ly", []string{"ly", "light_year", "light_years"}, "meter", 9_460_730_472_580_800, 0},

	// metric weight
	"kilogram":  {"kilogram", "kg", []string{"kg", "kilogram"}, "kilogram", 1, 0},
	"hectogram": {"hectogram", "hg", []string{"hg", "hectogram"}, "kilogram", math.Pow10(-1), 0},
	"decagram":  {"decagram", "dag", []string{"dag", "decagram"}, "kilogram", math.Pow10(-2), 0},
	"gram":      {"gram", "g", []string{"g", "gram"}, "kilogram", math.Pow10(-3), 0},
	"decigram":  {"decigram", "dg", []string{"dg", "decigram"}, "kilogram", math.Pow10(-4), 0},
	"centigram": {"centigram", "cg", []string{"cg", "centigram"}, "kilogram", math.Pow10(-5), 0},
	"milligram": {"milligram", "mg", []string{"mg", "milligram"}, "kilogram", math.Pow10(-6), 0},
	"microgram": {"microgram", "µg", []string{"µg", "microgram"}, "kilogram", math.Pow10(-6), 0},
	"tonne":     {"tonne", "ton", []string{"MG", "megagram", "tonne", "ton"}, "kilogram", math.Pow10(3), 0},
	// imperial weight
	"pound": {"pound", "lbs", []string{"lbs", "pound", "pounds"}, "kilogram", 0.45359237, 0},
	"ounce": {"ounce", "oz", []string{"oz", "ounce", "ounces"}, "kilogram", 0.028349523125, 0},

	// time
	"second":      {"second", "s", []string{"s", "second", "seconds"}, "second", 1, 0},
	"millisecond": {"millisecond", "ms", []string{"ms", "millisecond", "milliseconds"}, "second", math.Pow10(-3), 0},
	"minute":      {"minute", "min", []string{"min", "minute", "minutes"}, "second", 60, 0},
	"hour":        {"hour", "hours", []string{"hour", "hours"}, "second", 3600, 0},
	"day":         {"day", "days", []string{"day", "day", "days"}, "second", 86400, 0},
	"month":       {"month", "month", []string{"month", "months"}, "second", 2592000, 0},
	"year":        {"year", "year", []string{"year", "years"}, "second", 31556952, 0},

	// temperature
	"celsius":    {"celsius", "°C", []string{"C", "°C", "celsius"}, "celsius", 1, 0},
	"fahrenheit": {"fahrenheit", "°F", []string{"F", "°F", "fahrenheit"}, "celsius", float64(5) / 9, -32},
	"kelvin":     {"kelvin", "K", []string{"K", "kelvin"}, "celsius", 1, -273.15},

	// electic current
	"ampere": {"ampere", "A", []string{"A"}, "ampere", 1, 0},

	// currencies (exchange rates overridden at runtime with exchangeratesapi.io)
	"eur": {"eur", "€", []string{"€", "eur", "EUR"}, "eur", 1, 0},
	"usd": {"usd", "$", []string{"$", "usd", "USD"}, "eur", 0.84, 0},
	"gbp": {"gbp", "£", []string{"£", "gbp", "GBP"}, "eur", 1.17, 0},
	"cny": {"cny", "¥", []string{"cny", "CNY"}, "eur", 0.13, 0},
	"cad": {"cad", "CAD", []string{"cad", "CAD"}, "eur", 0.67, 0},

	// degrees
	"radians": {"radians", "rad", []string{"rad", "radians"}, "radians", 1, 0},
	"degrees": {"degrees", "deg", []string{"deg", "degrees"}, "radians", math.Pi / 180, 0},

	// pressure
	"pascal":                {"pascal", "Pa", []string{"Pa", "pascal"}, "pascal", 1, 0},
	"bar":                   {"bar", "bar", []string{"bar"}, "bar", 100_000, 0},
	"atmosphere":            {"atmosphere", "atm", []string{"atm", "atmosphere"}, "pascal", 101325, 0},
	"millimeter_of_mercury": {"millimeter_of_mercury", "mmHg", []string{"millimeter_of_mercury", "mmHg"}, "pascal", float64(101325) / 760, 0},

	// Data
	"bit":      {"bit", "bit", []string{"b", "bit"}, "bit", 1, 0},
	"byte":     {"byte", "B", []string{"B", "byte"}, "bit", 8, 0},
	"kilobit":  {"kilobit", "kbit", []string{"kbit", "kb", "kilobit"}, "bit", math.Pow10(3), 0},
	"kibibit":  {"kibibit", "Kibit", []string{"Kib", "Kibit", "kibibit"}, "bit", (1 << 10), 0},
	"kilobyte": {"kilobyte", "kB", []string{"kB", "kilobyte"}, "bit", 8 * math.Pow10(3), 0},
	"kibibyte": {"kibibyte", "KiB", []string{"KiB", "kibibyte"}, "bit", 8 * (1 << 10), 0},
	"megabit":  {"megabit", "Mbit", []string{"Mbit", "megabit"}, "bit", math.Pow10(6), 0},
	"mebibit":  {"mebibit", "Mibit", []string{"Mibit", "mebibit"}, "bit", (1 << 20), 0},
	"megabyte": {"megabyte", "MB", []string{"MB", "megabyte"}, "bit", 8 * math.Pow10(6), 0},
	"mebibyte": {"mebibyte", "MiB", []string{"MiB", "mebibyte"}, "bit", 8 * (1 << 20), 0},
	"gigabit":  {"gigabit", "Gbit", []string{"Gbit", "gigabit"}, "bit", math.Pow10(9), 0},
	"gibibit":  {"gibibit", "Gibit", []string{"Gibit", "gibibit"}, "bit", (1 << 30), 0},
	"gigabyte": {"gigabyte", "GB", []string{"GB", "gigabyte"}, "bit", 8 * math.Pow10(9), 0},
	"gibibyte": {"gibibyte", "GiB", []string{"GiB", "gibibyte"}, "bit", 8 * (1 << 30), 0},
	"terabit":  {"terabit", "Tbit", []string{"Tbit", "terabit"}, "bit", math.Pow10(12), 0},
	"tebibit":  {"tebibit", "Tibit", []string{"Tibit", "tibibit"}, "bit", (1 << 40), 0},
	"terabyte": {"terabyte", "TB", []string{"TB", "terabyte"}, "bit", 8 * math.Pow10(12), 0},
	"tebibyte": {"tebibyte", "TiB", []string{"TiB", "tibibyte"}, "bit", 8 * (1 << 40), 0},
	"petabit":  {"petabit", "Pbit", []string{"Pbit", "petabit"}, "bit", math.Pow10(15), 0},
	"pebibit":  {"pebibit", "Pibit", []string{"Pibit", "pebibit"}, "bit", (1 << 50), 0},
	"petabyte": {"petabyte", "PB", []string{"PB", "petabyte"}, "bit", 8 * math.Pow10(15), 0},
	"pebibyte": {"pebibyte", "PiB", []string{"PiB", "pebibyte"}, "bit", 8 * (1 << 50), 0},
}

func LoadUnitAliases() {
	for _, unit := range UnitTable {
		for _, str := range unit.Aliases {
			UnitAliasesMap[str] = unit.ID
		}
	}
}

func (u FundamentalUnit) String() string {
	return u.DisplayValue
}

func AreUnitsCompatible(u FundamentalUnit, v FundamentalUnit) bool {
	return u.BaseUnit == v.BaseUnit
}

func ConvertFundamentalUnits(value float64, from FundamentalUnit, to FundamentalUnit, exp float64) float64 {
	if !AreUnitsCompatible(from, to) {
		panic("Trying to convert incompatible units")
	}

	// Avoid converting from a unit to itself
	if from.ID == to.ID {
		return value
	}

	value = (value + from.ConversionShift) * math.Pow(from.ConversionFactor, exp)
	value = (value / math.Pow(to.ConversionFactor, exp)) - to.ConversionShift

	return value
}

type UnitExponent struct {
	Unit     FundamentalUnit
	Exponent float64
}

type CompositeUnit struct {
	UnitsList []UnitExponent
}

func (cu *CompositeUnit) IsEmpty() bool {
	return len(cu.UnitsList) == 0
}

func (cu CompositeUnit) IsCompatible(other CompositeUnit) bool {
	cu.Sort()
	other.Sort()

	if len(cu.UnitsList) != len(other.UnitsList) {
		return false
	}

	for i := 0; i < len(cu.UnitsList); i++ {
		if cu.UnitsList[i].Exponent != other.UnitsList[i].Exponent ||
			!AreUnitsCompatible(cu.UnitsList[i].Unit, other.UnitsList[i].Unit) {
			return false
		}
	}

	return true
}

func (cu *CompositeUnit) Sort() {
	sort.Slice(cu.UnitsList, func(i int, j int) bool {
		if cu.UnitsList[i].Exponent > 0 && cu.UnitsList[j].Exponent < 0 {
			return true
		}

		if cu.UnitsList[i].Exponent < 0 && cu.UnitsList[j].Exponent > 0 {
			return false
		}

		return cu.UnitsList[i].Unit.DisplayValue < cu.UnitsList[j].Unit.DisplayValue
	})
}

func (cu *CompositeUnit) SortByBaseUnitName() {
	sort.Slice(cu.UnitsList, func(i int, j int) bool {
		return cu.UnitsList[i].Unit.BaseUnit < cu.UnitsList[j].Unit.BaseUnit
	})
}

func (cu CompositeUnit) String() string {
	cu.Sort()
	s := ""

	positive := true
	for _, factor := range cu.UnitsList {
		if positive && factor.Exponent < 0 {
			if s == "" {
				s = "1"
			}

			s += " /"
		}

		if s != "" {
			s += " "
		}

		exp := factor.Exponent
		if exp < 0 {
			exp = -exp
		}

		if exp != 1 {
			s += fmt.Sprintf("%s^%s", factor.Unit.String(), strconv.FormatFloat(exp, 'f', -1, 32))
		} else {
			s += factor.Unit.String()
		}
	}

	return s
}

func ConvertCompositeUnits(value float64, from CompositeUnit, to CompositeUnit) (float64, error) {
	if !from.IsCompatible(to) {
		return 0, fmt.Errorf("Units are not compatible")
	}
	from.Sort()
	to.Sort()

	// BUG: composite units containing temperatures are broken
	for i := 0; i < len(from.UnitsList); i++ {
		value = ConvertFundamentalUnits(value, from.UnitsList[i].Unit, to.UnitsList[i].Unit, from.UnitsList[i].Exponent)
	}

	return value, nil
}

func CompositeUnitExponentiation(cu CompositeUnit, exp float64) CompositeUnit {
	newUnit := CompositeUnit{UnitsList: []UnitExponent{}}
	for i := 0; i < len(cu.UnitsList); i++ {
		newUnit.UnitsList = append(newUnit.UnitsList, UnitExponent{
			Unit:     cu.UnitsList[i].Unit,
			Exponent: cu.UnitsList[i].Exponent * exp,
		})
	}

	return newUnit
}

func CompositeUnitProduct(valueA float64, valueB float64, a CompositeUnit, b CompositeUnit) (float64, CompositeUnit) {
	a.SortByBaseUnitName()
	b.SortByBaseUnitName()

	product := CompositeUnit{UnitsList: []UnitExponent{}}
	aIndex := 0
	bIndex := 0
	value := valueA * valueB

	for aIndex < len(a.UnitsList) || bIndex < len(b.UnitsList) {
		if aIndex == len(a.UnitsList) {
			product.UnitsList = append(product.UnitsList, b.UnitsList[bIndex])
			bIndex++
		} else if bIndex == len(b.UnitsList) {
			product.UnitsList = append(product.UnitsList, a.UnitsList[aIndex])
			aIndex++
		} else if AreUnitsCompatible(a.UnitsList[aIndex].Unit, b.UnitsList[bIndex].Unit) {
			unit := a.UnitsList[aIndex]
			unit.Exponent = a.UnitsList[aIndex].Exponent + b.UnitsList[bIndex].Exponent

			value = ConvertFundamentalUnits(
				value,
				b.UnitsList[bIndex].Unit,
				a.UnitsList[aIndex].Unit,
				b.UnitsList[bIndex].Exponent,
			)

			if unit.Exponent != 0 {
				product.UnitsList = append(product.UnitsList, unit)
			}

			aIndex++
			bIndex++
		} else if a.UnitsList[aIndex].Unit.BaseUnit < b.UnitsList[bIndex].Unit.BaseUnit {
			product.UnitsList = append(product.UnitsList, a.UnitsList[aIndex])
			aIndex++
		} else {
			product.UnitsList = append(product.UnitsList, b.UnitsList[bIndex])
			bIndex++
		}
	}

	product.Sort()

	return value, product
}
func CompositeUnitDivision(valueA float64, valueB float64, a CompositeUnit, b CompositeUnit) (float64, CompositeUnit) {
	b = CompositeUnitExponentiation(b, -1)

	return CompositeUnitProduct(valueA, 1/valueB, a, b)
}
