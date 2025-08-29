package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IngredientQuantity represents a parsed ingredient quantity
type IngredientQuantity struct {
	Amount float64
	Unit   string
}

// IngredientAggregator handles ingredient amount aggregation
type IngredientAggregator struct {
	unitConversions map[string]map[string]float64
}

// NewIngredientAggregator creates a new ingredient aggregator
func NewIngredientAggregator() *IngredientAggregator {
	return &IngredientAggregator{
		unitConversions: initUnitConversions(),
	}
}

// initUnitConversions initializes unit conversion table
func initUnitConversions() map[string]map[string]float64 {
	return map[string]map[string]float64{
		// Weight conversions (to grams)
		"weight": {
			"g":   1.0,
			"グラム": 1.0,
			"kg":  1000.0,
			"キロ":  1000.0,
		},
		// Volume conversions (to ml)
		"volume": {
			"ml":     1.0,
			"ミリリットル": 1.0,
			"l":      1000.0,
			"リットル":   1000.0,
			"cc":     1.0,
			"大さじ":    15.0,
			"小さじ":    5.0,
			"カップ":    200.0,
			"合":      180.0,
		},
		// Count (pieces)
		"count": {
			"個":   1.0,
			"本":   1.0,
			"枚":   1.0,
			"切れ":  1.0,
			"かけ":  1.0,
			"パック": 1.0,
			"袋":   1.0,
			"缶":   1.0,
		},
	}
}

// ParseQuantity parses an amount string into quantity and unit
func (a *IngredientAggregator) ParseQuantity(amountStr string) (*IngredientQuantity, error) {
	if amountStr == "" {
		return nil, fmt.Errorf("empty amount string")
	}

	// Remove whitespace
	amountStr = strings.TrimSpace(amountStr)

	// Handle special cases
	if strings.Contains(amountStr, "適量") || strings.Contains(amountStr, "少々") {
		return &IngredientQuantity{Amount: 0, Unit: "適量"}, nil
	}

	// Try different patterns for Japanese-style quantities
	// Pattern 1: "大さじ1" style (unit first, then number)
	unitFirstRe := regexp.MustCompile(`^([^\d]+)(\d+\.?\d*)$`)
	unitFirstMatches := unitFirstRe.FindStringSubmatch(amountStr)

	if len(unitFirstMatches) == 3 {
		amount, err := strconv.ParseFloat(unitFirstMatches[2], 64)
		if err == nil {
			unit := strings.TrimSpace(unitFirstMatches[1])
			return &IngredientQuantity{Amount: amount, Unit: unit}, nil
		}
	}

	// Pattern 2: "1大さじ" style (number first, then unit)
	numberFirstRe := regexp.MustCompile(`^(\d+\.?\d*)\s*(.*)$`)
	numberFirstMatches := numberFirstRe.FindStringSubmatch(amountStr)

	if len(numberFirstMatches) != 3 {
		return &IngredientQuantity{Amount: 0, Unit: "適量"}, nil
	}

	amount, err := strconv.ParseFloat(numberFirstMatches[1], 64)
	if err != nil {
		return &IngredientQuantity{Amount: 0, Unit: "適量"}, nil
	}

	unit := strings.TrimSpace(numberFirstMatches[2])
	if unit == "" {
		unit = "個"
	}

	return &IngredientQuantity{Amount: amount, Unit: unit}, nil
}

// GetUnitType returns the unit type (weight, volume, count) for a given unit
func (a *IngredientAggregator) GetUnitType(unit string) string {
	for unitType, units := range a.unitConversions {
		if _, exists := units[unit]; exists {
			return unitType
		}
	}
	return "unknown"
}

// ConvertToBaseUnit converts quantity to base unit of its type
func (a *IngredientAggregator) ConvertToBaseUnit(qty *IngredientQuantity) (*IngredientQuantity, error) {
	if qty.Unit == "適量" {
		return qty, nil
	}

	unitType := a.GetUnitType(qty.Unit)
	if unitType == "unknown" {
		return qty, nil
	}

	conversion, exists := a.unitConversions[unitType][qty.Unit]
	if !exists {
		return qty, nil
	}

	baseUnit := a.getBaseUnit(unitType)
	baseAmount := qty.Amount * conversion

	return &IngredientQuantity{
		Amount: baseAmount,
		Unit:   baseUnit,
	}, nil
}

// getBaseUnit returns the base unit for each unit type
func (a *IngredientAggregator) getBaseUnit(unitType string) string {
	switch unitType {
	case "weight":
		return "g"
	case "volume":
		return "ml"
	case "count":
		return "個"
	default:
		return "個"
	}
}

// AggregateQuantities aggregates multiple quantities of the same ingredient
func (a *IngredientAggregator) AggregateQuantities(quantities []*IngredientQuantity) (*IngredientQuantity, error) {
	if len(quantities) == 0 {
		return &IngredientQuantity{Amount: 0, Unit: "個"}, nil
	}

	if len(quantities) == 1 {
		return quantities[0], nil
	}

	// Check if any quantity is "適量"
	for _, qty := range quantities {
		if qty.Unit == "適量" {
			return &IngredientQuantity{Amount: 0, Unit: "適量"}, nil
		}
	}

	// Convert all to base units
	baseQuantities := make([]*IngredientQuantity, 0, len(quantities))
	var targetUnitType string

	for i, qty := range quantities {
		baseQty, err := a.ConvertToBaseUnit(qty)
		if err != nil {
			return &IngredientQuantity{Amount: 0, Unit: "適量"}, err
		}

		unitType := a.GetUnitType(baseQty.Unit)
		if i == 0 {
			targetUnitType = unitType
		} else if unitType != "unknown" && unitType != targetUnitType {
			// Different unit types cannot be aggregated
			return &IngredientQuantity{Amount: 0, Unit: "適量"}, nil
		}

		baseQuantities = append(baseQuantities, baseQty)
	}

	// Sum up the base quantities
	totalAmount := 0.0
	baseUnit := ""
	for _, baseQty := range baseQuantities {
		if baseUnit == "" {
			baseUnit = baseQty.Unit
		}
		totalAmount += baseQty.Amount
	}

	// Convert back to appropriate display unit
	result := &IngredientQuantity{Amount: totalAmount, Unit: baseUnit}
	return a.ConvertToDisplayUnit(result), nil
}

// ConvertToDisplayUnit converts to user-friendly display unit
func (a *IngredientAggregator) ConvertToDisplayUnit(qty *IngredientQuantity) *IngredientQuantity {
	if qty.Unit == "適量" {
		return qty
	}

	switch qty.Unit {
	case "g":
		if qty.Amount >= 1000 {
			return &IngredientQuantity{
				Amount: qty.Amount / 1000,
				Unit:   "kg",
			}
		}
	case "ml":
		if qty.Amount >= 1000 {
			return &IngredientQuantity{
				Amount: qty.Amount / 1000,
				Unit:   "l",
			}
		}
	}

	return qty
}

// FormatQuantity formats a quantity for display
func (a *IngredientAggregator) FormatQuantity(qty *IngredientQuantity) string {
	if qty.Unit == "適量" {
		return "適量"
	}

	// Format amount based on its value
	var amountStr string
	if qty.Amount == float64(int(qty.Amount)) {
		amountStr = fmt.Sprintf("%.0f", qty.Amount)
	} else {
		amountStr = fmt.Sprintf("%.1f", qty.Amount)
	}

	return fmt.Sprintf("%s%s", amountStr, qty.Unit)
}
