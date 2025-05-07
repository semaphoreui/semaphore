package conv

func ConvertFloatToIntIfPossible(v interface{}) (int64, bool) {

	switch v.(type) {
	case float64:
		f := v.(float64)
		i := int64(f)
		if float64(i) == f {
			return i, true
		}
	case float32:
		f := v.(float32)
		i := int64(f)
		if float32(i) == f {
			return i, true
		}
	}

	return 0, false
}
