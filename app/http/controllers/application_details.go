package controllers

import (
	"strconv"
	"strings"
)

var validEmploymentTypes = map[string]struct{}{
	"full_time":  {},
	"part_time":  {},
	"internship": {},
	"freelance":  {},
}

type applicationDetails struct {
	EmploymentType *string
	SalaryMin      *int64
	SalaryMax      *int64
}

func validateApplicationDetails(employmentType, salaryMin, salaryMax string) (applicationDetails, map[string]string) {
	details := applicationDetails{}
	errors := make(map[string]string)

	employmentType = strings.TrimSpace(employmentType)
	if employmentType != "" {
		if _, ok := validEmploymentTypes[employmentType]; !ok {
			errors["employment_type"] = "Employment type tidak valid"
		} else {
			details.EmploymentType = &employmentType
		}
	}

	salaryMin = strings.TrimSpace(salaryMin)
	salaryMax = strings.TrimSpace(salaryMax)
	if (salaryMin == "") != (salaryMax == "") {
		errors["salary_min"] = "Minimum dan maksimum salary harus diisi bersama"
		errors["salary_max"] = "Minimum dan maksimum salary harus diisi bersama"
		return details, errors
	}

	if salaryMin == "" {
		return details, errors
	}

	parsedMin, err := strconv.ParseInt(salaryMin, 10, 64)
	if err != nil || parsedMin < 0 {
		errors["salary_min"] = "Minimum salary harus berupa angka positif"
	}
	parsedMax, err := strconv.ParseInt(salaryMax, 10, 64)
	if err != nil || parsedMax < 0 {
		errors["salary_max"] = "Maksimum salary harus berupa angka positif"
	}
	if len(errors) > 0 {
		return details, errors
	}
	if parsedMax < parsedMin {
		errors["salary_max"] = "Maksimum salary tidak boleh lebih kecil dari minimum salary"
		return details, errors
	}

	details.SalaryMin = &parsedMin
	details.SalaryMax = &parsedMax
	return details, errors
}
