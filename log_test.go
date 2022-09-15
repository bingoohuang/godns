package main

import (
	"bufio"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConsole(t *testing.T) {
	l := NewLogger()
	l.SetLogger("console", nil)
	l.SetLevel(LevelInfo)

	l.Debug("debug")
	l.Info("info")
	l.Notice("notiece")
	l.Warn("warn")
	l.Error("error")
}

func TestFile(t *testing.T) {
	l := NewLogger()
	l.SetLogger("file", map[string]interface{}{"file": "test.log"})
	l.SetLevel(LevelInfo)

	l.Debug("debug")
	l.Info("info")
	l.Notice("notice")
	l.Warn("warn")
	l.Error("error")

	time.Sleep(time.Second)

	f, err := os.Open("test.log")
	if err != nil {
		t.Fatal(err)
	}
	b := bufio.NewReader(f)
	lineNum := 0
	for {
		line, _, err := b.ReadLine()
		if err != nil {
			break
		}
		if len(line) > 0 {
			lineNum++
		}
	}

	Convey("Test Log File Handler", t, func() {
		Convey("file line nums should be 4", func() {
			So(lineNum, ShouldEqual, 4)
		})
	})

	os.Remove("test.log")
}
