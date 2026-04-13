package dto

func PtrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func PtrUint(p *uint) uint {
	if p == nil {
		return 0
	}
	return *p
}
