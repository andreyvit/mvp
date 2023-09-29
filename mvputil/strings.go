package mvputil

func YesNo(v bool, yes, no string) string {
	if v {
		return yes
	} else {
		return no
	}
}
