package generator

import "strings"

func parameters(parameter string) map[string]string {
	parameters := make(map[string]string)
	for _, p := range strings.Split(parameter, ",") {
		if i := strings.Index(p, "="); i < 0 {
			parameters[p] = ""
		} else {
			parameters[p[0:i]] = p[i+1:]
		}
	}
	return parameters
}
