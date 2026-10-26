package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/store"
	"github.com/sitthichok-J/self-study-2026/day03-go-fundamentals/internal/tasks"
)

func TestRun(t *testing.T) {
	svc := tasks.New(&store.Memory{})
	tests := []struct {
		name    string
		args    []string
		wantOut string
		wantErr string
	}{
		{"ไม่มี arg", nil, "", "usage"},
		{"add", []string{"add", "ซื้อ", "กาแฟ"}, "added #1 ซื้อ กาแฟ\n", ""},
		{"add ว่าง", []string{"add"}, "", "title is empty"},
		{"list", []string{"list"}, "[ ] #1 ซื้อ กาแฟ\n", ""},
		{"done", []string{"done", "1"}, "done #1 ok\n", ""},
		{"list หลัง done", []string{"list"}, "[x] #1 ซื้อ กาแฟ\n", ""},
		{"done ไม่มี id", []string{"done"}, "", "need id"},
		{"done id ไม่ใช่เลข", []string{"done", "abc"}, "", "bad id"},
		{"done ไม่เจอ", []string{"done", "9"}, "", "task 9 not found"},
		{"rm", []string{"rm", "1"}, "rm #1 ok\n", ""},
		{"คำสั่งแปลก", []string{"zzz"}, "", "unknown command"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := run(svc, tc.args, &buf)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected err %v", err)
			}
			if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("err=%v want contains %q", err, tc.wantErr)
			}
			if buf.String() != tc.wantOut {
				t.Fatalf("out=%q want %q", buf.String(), tc.wantOut)
			}
		})
	}
}
