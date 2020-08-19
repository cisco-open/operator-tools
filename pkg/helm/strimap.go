package helm

type Strimap = map[string]interface{}

type StrimapBuilder Strimap

func (s StrimapBuilder) Getin(strs ...string) Strimap {
	if s == nil || len(strs) == 0 {
		return nil
	}

	submap, ok := s[strs[0]].(Strimap)
	if !ok {
		return nil
	}

	if len(strs[1:]) > 0 {
		return StrimapBuilder(submap).Getin(strs[1:]...)
	}
	return submap
}

func MergeMaps(a, b Strimap) Strimap {
	out := make(Strimap, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(Strimap); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(Strimap); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}