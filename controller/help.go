package controller

type Help struct {
	lines []string
}

func NewHelp() *Help {
	return &Help{}
}

func (h *Help) SetKeyMap(lines []string) {
	h.lines = lines
}

func (h *Help) Help() []string {
	return h.lines
}
