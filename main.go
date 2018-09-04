package main

import (
	"fmt"
	"github.com/wangmingzhitu/memo/memo"
)

func main(){
	monitor := memo.InitUiMonitor()
	go memo.Serve()
	for{
		select{
		case err := <-monitor.Err:
			fmt.Println("Error in main:", err)
		}
	}
}
