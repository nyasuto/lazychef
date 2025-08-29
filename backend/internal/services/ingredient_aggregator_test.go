package services

import (
	"testing"
)

func TestNewIngredientAggregator(t *testing.T) {
	aggregator := NewIngredientAggregator()
	if aggregator == nil {
		t.Fatal("Expected non-nil aggregator")
	}
}

func TestParseQuantity(t *testing.T) {
	aggregator := NewIngredientAggregator()

	tests := []struct {
		input    string
		expected *IngredientQuantity
		hasError bool
	}{
		{"2個", &IngredientQuantity{Amount: 2, Unit: "個"}, false},
		{"100g", &IngredientQuantity{Amount: 100, Unit: "g"}, false},
		{"2.5kg", &IngredientQuantity{Amount: 2.5, Unit: "kg"}, false},
		{"大さじ1", &IngredientQuantity{Amount: 1, Unit: "大さじ"}, false},
		{"適量", &IngredientQuantity{Amount: 0, Unit: "適量"}, false},
		{"少々", &IngredientQuantity{Amount: 0, Unit: "適量"}, false},
		{"", nil, true},
		{"2", &IngredientQuantity{Amount: 2, Unit: "個"}, false},
		{"3 本", &IngredientQuantity{Amount: 3, Unit: "本"}, false},
	}

	for _, test := range tests {
		result, err := aggregator.ParseQuantity(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s', but got none", test.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for input '%s': %v", test.input, err)
			continue
		}

		if result.Amount != test.expected.Amount || result.Unit != test.expected.Unit {
			t.Errorf("For input '%s', expected %+v, got %+v",
				test.input, test.expected, result)
		}
	}
}

func TestGetUnitType(t *testing.T) {
	aggregator := NewIngredientAggregator()

	tests := []struct {
		unit     string
		expected string
	}{
		{"g", "weight"},
		{"kg", "weight"},
		{"ml", "volume"},
		{"大さじ", "volume"},
		{"個", "count"},
		{"本", "count"},
		{"unknown", "unknown"},
	}

	for _, test := range tests {
		result := aggregator.GetUnitType(test.unit)
		if result != test.expected {
			t.Errorf("For unit '%s', expected '%s', got '%s'",
				test.unit, test.expected, result)
		}
	}
}

func TestConvertToBaseUnit(t *testing.T) {
	aggregator := NewIngredientAggregator()

	tests := []struct {
		input    *IngredientQuantity
		expected *IngredientQuantity
	}{
		{&IngredientQuantity{Amount: 1, Unit: "kg"}, &IngredientQuantity{Amount: 1000, Unit: "g"}},
		{&IngredientQuantity{Amount: 2, Unit: "大さじ"}, &IngredientQuantity{Amount: 30, Unit: "ml"}},
		{&IngredientQuantity{Amount: 3, Unit: "個"}, &IngredientQuantity{Amount: 3, Unit: "個"}},
		{&IngredientQuantity{Amount: 0, Unit: "適量"}, &IngredientQuantity{Amount: 0, Unit: "適量"}},
	}

	for _, test := range tests {
		result, err := aggregator.ConvertToBaseUnit(test.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			continue
		}

		if result.Amount != test.expected.Amount || result.Unit != test.expected.Unit {
			t.Errorf("For input %+v, expected %+v, got %+v",
				test.input, test.expected, result)
		}
	}
}

func TestAggregateQuantities(t *testing.T) {
	aggregator := NewIngredientAggregator()

	tests := []struct {
		name     string
		input    []*IngredientQuantity
		expected *IngredientQuantity
	}{
		{
			name: "Same units",
			input: []*IngredientQuantity{
				{Amount: 2, Unit: "個"},
				{Amount: 3, Unit: "個"},
			},
			expected: &IngredientQuantity{Amount: 5, Unit: "個"},
		},
		{
			name: "Different weight units",
			input: []*IngredientQuantity{
				{Amount: 500, Unit: "g"},
				{Amount: 1, Unit: "kg"},
			},
			expected: &IngredientQuantity{Amount: 1.5, Unit: "kg"},
		},
		{
			name: "Volume units",
			input: []*IngredientQuantity{
				{Amount: 2, Unit: "大さじ"},
				{Amount: 1, Unit: "小さじ"},
			},
			expected: &IngredientQuantity{Amount: 35, Unit: "ml"},
		},
		{
			name: "With 適量",
			input: []*IngredientQuantity{
				{Amount: 2, Unit: "個"},
				{Amount: 0, Unit: "適量"},
			},
			expected: &IngredientQuantity{Amount: 0, Unit: "適量"},
		},
		{
			name:     "Empty input",
			input:    []*IngredientQuantity{},
			expected: &IngredientQuantity{Amount: 0, Unit: "個"},
		},
		{
			name: "Single quantity",
			input: []*IngredientQuantity{
				{Amount: 3, Unit: "本"},
			},
			expected: &IngredientQuantity{Amount: 3, Unit: "本"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := aggregator.AggregateQuantities(test.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Amount != test.expected.Amount || result.Unit != test.expected.Unit {
				t.Errorf("Expected %+v, got %+v", test.expected, result)
			}
		})
	}
}

func TestConvertToDisplayUnit(t *testing.T) {
	aggregator := NewIngredientAggregator()

	tests := []struct {
		input    *IngredientQuantity
		expected *IngredientQuantity
	}{
		{&IngredientQuantity{Amount: 1500, Unit: "g"}, &IngredientQuantity{Amount: 1.5, Unit: "kg"}},
		{&IngredientQuantity{Amount: 500, Unit: "g"}, &IngredientQuantity{Amount: 500, Unit: "g"}},
		{&IngredientQuantity{Amount: 2000, Unit: "ml"}, &IngredientQuantity{Amount: 2, Unit: "l"}},
		{&IngredientQuantity{Amount: 500, Unit: "ml"}, &IngredientQuantity{Amount: 500, Unit: "ml"}},
		{&IngredientQuantity{Amount: 0, Unit: "適量"}, &IngredientQuantity{Amount: 0, Unit: "適量"}},
	}

	for _, test := range tests {
		result := aggregator.ConvertToDisplayUnit(test.input)
		if result.Amount != test.expected.Amount || result.Unit != test.expected.Unit {
			t.Errorf("For input %+v, expected %+v, got %+v",
				test.input, test.expected, result)
		}
	}
}

func TestFormatQuantity(t *testing.T) {
	aggregator := NewIngredientAggregator()

	tests := []struct {
		input    *IngredientQuantity
		expected string
	}{
		{&IngredientQuantity{Amount: 2, Unit: "個"}, "2個"},
		{&IngredientQuantity{Amount: 1.5, Unit: "kg"}, "1.5kg"},
		{&IngredientQuantity{Amount: 100, Unit: "g"}, "100g"},
		{&IngredientQuantity{Amount: 0, Unit: "適量"}, "適量"},
	}

	for _, test := range tests {
		result := aggregator.FormatQuantity(test.input)
		if result != test.expected {
			t.Errorf("For input %+v, expected '%s', got '%s'",
				test.input, test.expected, result)
		}
	}
}
