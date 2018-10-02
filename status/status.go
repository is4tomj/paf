package status

import (
	"fmt"
	"os"
)

var pe = os.Stderr.Write
var sprintf = fmt.Sprintf
func pes(str string) {
	pe([]byte(str))
}
func pesf(format string, args ...interface{}) {
	pe([]byte(sprintf(format, args)))
}

type Status struct {
	MaxLines int
}

func (s *Status)Init() {
	max := (*s).MaxLines
	for i := 0; i < max; i++ {
		pes("\n")
	}
}

func (s *Status)Set(str string, line int) {
	offset := (*s).MaxLines - line
	fstr := fmt.Sprintf("\033[%dF\033[2K%s\033[%dE", offset, str, offset)
	os.Stderr.Write([]byte(fstr))
}
