package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// RepeatitiveFunctionProvide provides utility functions for handlers
// Note: Most business logic has been moved to service layer
type RepeatitiveFunctionProvide struct {
}

// Generate log message with specified parameters
func (rfp *RepeatitiveFunctionProvide) log_output(
	log_type string,
	ip_address string,
	user_agent string,
	fmt_msg string,
	args ...interface{},
) {
	prefix := fmt.Sprintf("From %s, User-Agent is %s. ", ip_address, user_agent)
	finalMsg := prefix + fmt.Sprintf(fmt_msg, args...)
	if fn, ok := log_func[log_type]; ok {
		fn(finalMsg)
	} else {
		log.Println(finalMsg)
	}
}

func (rfp *RepeatitiveFunctionProvide) log_with_ctx(
	c *gin.Context,
	log_type string,
	fmt_msg string,
	args ...interface{},
) {
	rfp.log_output(
		log_type,
		c.Request.RemoteAddr,
		c.GetHeader("User-Agent"),
		fmt_msg,
		args...,
	)
}
