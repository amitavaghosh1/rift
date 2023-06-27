package display

import "fmt"

type SQSFormatter struct {
	TextFormat bool
}

func (sf *SQSFormatter) Display(queueName, message string) {
	fmt.Println(queueName)
	fmt.Println(message)
}
