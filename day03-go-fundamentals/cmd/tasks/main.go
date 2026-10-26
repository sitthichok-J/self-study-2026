// CLI: tasks [-file tasks.json] add <title> | list | done <id> | rm <id>
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/store"
	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/tasks"
)

func main() {
	file := flag.String("file", "tasks.json", "path ของไฟล์เก็บงาน")
	flag.Parse()
	svc := tasks.New(store.File{Path: *file})
	// log/แสดง error ที่ชั้นบนสุดครั้งเดียว
	if err := run(svc, flag.Args(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(svc *tasks.Service, args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: tasks add <title> | list | done <id> | rm <id>")
	}
	switch args[0] {
	case "add":
		t, err := svc.Add(strings.Join(args[1:], " "))
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "added #%d %s\n", t.ID, t.Title)
	case "list":
		list, err := svc.List()
		if err != nil {
			return err
		}
		for _, t := range list {
			mark := " "
			if t.Done {
				mark = "x"
			}
			fmt.Fprintf(out, "[%s] #%d %s\n", mark, t.ID, t.Title)
		}
	case "done", "rm":
		if len(args) < 2 {
			return fmt.Errorf("%s: need id", args[0])
		}
		id, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("%s: bad id %q: %w", args[0], args[1], err)
		}
		if args[0] == "done" {
			err = svc.Done(id)
		} else {
			err = svc.Remove(id)
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s #%d ok\n", args[0], id)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}
