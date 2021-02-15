package main

import (
	"fmt"
	"log"

	"github.com/zurekp/yadmin/client"
)

func main() {
	fmt.Println("Still works")
}

func test(url, login, password string) {
	fmt.Println("Testing: ", url)
	c, err := client.New(client.ClientOptions{
		BaseUrl:                   url,
		AdminLogin:                login,
		AdminPassword:             password,
		SkipCertificateValidation: true})
	failOnError(err)

	s, err := c.Status()
	failOnError(err)

	fmt.Printf("%+v\n", s)
}

func failOnError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
