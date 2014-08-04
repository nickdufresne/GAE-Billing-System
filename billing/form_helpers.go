package billing

import (
	"strconv"
)

func getFormFieldString(m map[string][]string, f string) string {
	if res, ok := m[f]; ok {
		if len(res) > 0 {
			return res[0]
		}
	}

	return ""
}

func getFormFieldInt(m map[string][]string, f string) int {
	res := getFormFieldString(m, f)
	if res == "" {
		return 0
	}

	val, err := strconv.Atoi(res)

	if err != nil {
		return 0
	}

	return val
}
