package gumble_ffmpeg

import (
	"io"
	"os/exec"
)

type Source interface {
	// must include the -i <filename>
	arguments() []string
	start(*exec.Cmd)
	done()
}

// sourceFile

type sourceFile string

// SourceFile is standard file source.
func SourceFile(filename string) Source {
	return sourceFile(filename)
}

func (s sourceFile) arguments() []string {
	return []string{"-i", string(s)}
}

func (sourceFile) start(*exec.Cmd) {
}

func (sourceFile) done() {
}

// sourceReader

type sourceReader struct {
	r io.ReadCloser
}

// SourceReader is a ReadCloser source.
func SourceReader(r io.ReadCloser) Source {
	return &sourceReader{r}
}

func (*sourceReader) arguments() []string {
	return []string{"-i", "-"}
}

func (s *sourceReader) start(cmd *exec.Cmd) {
	cmd.Stdin = s.r
}

func (s *sourceReader) done() {
	s.r.Close()
}
