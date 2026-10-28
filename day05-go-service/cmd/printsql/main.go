// พิมพ์ query ทั้งหมดของ Postgres repository (ใช้กับ scripts/check_sql.sh)
package main

import (
	"fmt"
	"sort"

	"github.com/sitthichok-J/self-study-2026/day05-go-service/internal/repository"
)

func main() {
	q := repository.Queries()
	names := make([]string, 0, len(q))
	for n := range q {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Printf("-- %s\n%s;\n", n, q[n])
	}
}
