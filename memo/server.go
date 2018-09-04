package memo

import (
	"fmt"
	"html/template"
	"net/http"
)

var uMonitor *Monitor

func InitUiMonitor() *Monitor{
	uMonitor = NewMonitor()
	uMonitor.Start()
	return uMonitor
}

func Serve(){
	fmt.Println("Server is listening on port 5050...")
	// 启动静态文件服务
	h := http.FileServer(http.Dir("memo/static"))
    http.Handle("/static/", http.StripPrefix("/static/", h))
	http.HandleFunc("/", showTodos)
	http.HandleFunc("/editor", editTodo)
	http.ListenAndServe(":5050", nil)
}

func showTodos(w http.ResponseWriter, req *http.Request){
	fmt.Println(req.Proto, req.Method, req.URL)
	if req.URL.Path != "/"{
		http.NotFound(w, req)
		return
	}
	t, _ := template.ParseFiles("memo/templates/index.html")
	t.Execute(w, uMonitor.TodoItems)
}

func editTodo(w http.ResponseWriter, req *http.Request){
	// TODO
}
