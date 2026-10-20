// user-service owns user accounts and login, and is the system's only public
// entry point: a REST/JSON gateway that checks the JWT and forwards requests
// to the other services over gRPC. See README.md and docs/.
package main

import "github.com/fnxr21/user-service/cmd"

func main() {
	cmd.Run()
}
