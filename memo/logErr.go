package memo

import (
	"fmt"
	"log"
)

func handleCriticalErr(err error, echoMsg string){
	if err != nil{
		log.Fatalln(echoMsg, err)
	}
}

func handleInfoErr(err error, echoMsg string) error{
	if err != nil{
		log.Println(echoMsg, err)
		return fmt.Errorf(echoMsg)
	}
	return nil
}
