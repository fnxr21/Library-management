// book-service stores books, their categories and stock, and serves them
// over gRPC. It calls author-service and category-service to check
// references. See README.md and docs/ for how it fits into the system.
package main

import "github.com/fnxr21/book-service/cmd"

func main() {
	cmd.Run()
}
