package tools

func Combination(v []string, n int) (r [][]string) {
	bits := make([]int32, len(v))
	for i := 0; i < n; i++ {
		bits[i] = 1
	}

	for {
		var d []string
		for i := range v {
			if bits[i] == 1 {
				d = append(d, v[i])
			}
		}
		r = append(r, d)

		first := -1
		for i := range bits {
			if i == 0 {
				continue
			}
			if bits[i-1] == 1 && bits[i] == 0 {
				first = i
				break
			}
		}

		if first != -1 {
			bits[first-1] = 0
			bits[first] = 1

			k := -1
			for i := 0; i < first-1; i++ {
				if bits[i] == 0 {
					k = i
					break
				}
			}

			if k != -1 {
				c := k + 1
				for ; c < first-1; c++ {
					if bits[c] == 1 {
						bits[k] = 1
						bits[c] = 0
						k++
					}
				}
			}
		} else {
			break
		}
	}

	return r
}
