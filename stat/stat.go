package stat

type Stat struct {
	Successful map[string]*Next
	Failed     map[string]*Next
}

type Next struct {
	Node  string
	Count int
}
