package controllers

import "testing"

func TestValidateApplicationDetails(t *testing.T) {
	tests := []struct {
		name           string
		employmentType string
		salaryMin      string
		salaryMax      string
		wantErrors     bool
	}{
		{name: "all optional", wantErrors: false},
		{name: "valid details", employmentType: "internship", salaryMin: "4000000", salaryMax: "10000000", wantErrors: false},
		{name: "invalid employment type", employmentType: "contract", wantErrors: true},
		{name: "missing maximum", salaryMin: "4000000", wantErrors: true},
		{name: "negative salary", salaryMin: "-1", salaryMax: "100", wantErrors: true},
		{name: "reversed salary", salaryMin: "10000000", salaryMax: "4000000", wantErrors: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details, errors := validateApplicationDetails(tt.employmentType, tt.salaryMin, tt.salaryMax)
			if (len(errors) > 0) != tt.wantErrors {
				t.Fatalf("validateApplicationDetails() errors = %v, wantErrors %v", errors, tt.wantErrors)
			}
			if !tt.wantErrors && tt.salaryMin != "" && (details.SalaryMin == nil || details.SalaryMax == nil) {
				t.Fatal("valid salary range was not parsed")
			}
		})
	}
}
