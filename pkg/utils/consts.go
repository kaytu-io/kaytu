package utils

import "fmt"

var (
	ContactUsEmail   = "hello@kaytu.io"
	ContactUsMessage = fmt.Sprintf("You have reached the limit for this user and organization.\n"+
		"Contact us and request for increase in limits:\n"+
		"%s", ContactUsEmail)
)
