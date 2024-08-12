package main

import (
	"log"

	"github.com/EinStack/glide/cmd"
)

//	@title			Glide
//	@version		0.0.1
//	@description	API documentation for Glide, an open-source lightweight high-performance model gateway

//	@contact.name	EinStack Community
//	@contact.url	https://github.com/EinStack/glide/
//  @contact.email  contact@einstack.ai

//	@license.name	Apache 2.0
//	@license.url	https://github.com/EinStack/glide/blob/develop/LICENSE

// @externalDocs.description  Documentation
// @externalDocs.url          https://glide.einstack.ai/

// @host		localhost:9099
// @BasePath	/
// @schemes	http
func main() {
	cli := cmd.NewCLI()

	if err := cli.Execute(); err != nil {
		log.Fatalf("💥Glide has finished with error: %v", err)
	}
}
