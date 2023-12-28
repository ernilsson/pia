package command

func must(wd string, err error) string {
	if err != nil {
		panic(err)
	}
	return wd
}
