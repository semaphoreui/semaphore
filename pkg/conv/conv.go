package conv

func ConvertFloatToIntIfPossible(v any) (int64, bool) {

	switch v := v.(type) {
	case float64:
		f := v
		i := int64(f)
		if float64(i) == f {
			return i, true
		}
	case float32:
		f := v
		i := int64(f)
		if float32(i) == f {
			return i, true
		}
	}

	return 0, false
}
