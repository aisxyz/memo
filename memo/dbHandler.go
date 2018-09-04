package memo

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dbName = "memo?parseTime=true&loc=Local"
	createDbSql = `CREATE DATABASE IF NOT EXISTS memo;`
	createTodoTableSql = `
	CREATE TABLE IF NOT EXISTS todolist(
		id INT(3) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		theme VARCHAR(50) NOT NULL,
		email VARCHAR(30) NOT NULL,
		remindTime DATETIME(0) NOT NULL,
		lastRemind DATETIME(0),
		calendarType ENUM('lunar', 'solar') DEFAULT 'solar',
		content VARCHAR(500) NOT NULL
	)ENGINE=InnoDB;
	`
)

var db *sql.DB

func init(){
	var err error
	var (
		user = os.Getenv("DbUser")
		pwd = os.Getenv("DbPwd")
	)
	dsn := fmt.Sprintf("%s:%s@/", user, pwd)
	db, err = sql.Open("mysql", dsn + dbName)
	if db.Ping() != nil || err != nil{
		db, _ = sql.Open("mysql", dsn)
		_, err = db.Exec(createDbSql)
		handleCriticalErr(err, "Create database error:")
		db.Close()
		db, err = sql.Open("mysql", dsn + dbName)
		handleCriticalErr(err, "Connect database error:")
	}
	_, err = db.Exec(createTodoTableSql)
	handleCriticalErr(err, "Create todolist table error:")
}

type DbOp int

const (
	DbInsert DbOp = iota
	DbUpdate
	DbDelete
)

func queryTodoItems() (items TodoList){
	sql := "SELECT * FROM todolist;"
	rows, err := db.Query(sql)
	if err != nil{
		return
	}
	defer rows.Close()
	for rows.Next(){
		var item TodoItem
		err := rows.Scan(&item.Id, &item.Theme, &item.Email, &item.RemindTime, &item.LastRemind, &item.CalendarType, &item.Content)
		if err == nil{
			items = append(items, item)
		}
	}
	if err = rows.Err(); err != nil{
		handleInfoErr(err, "QueryTodoItems():")
	}
	return items
}

func syncDb(op DbOp, pItem *TodoItem) error{
	if pItem.CalendarType == ""{
		pItem.CalendarType = "solar"
	}
	switch op{
	case DbInsert:
		return insertTodoItem(pItem)
	case DbUpdate:
		return updateTodoItem(pItem)
	case DbDelete:
		return deleteTodoItem(pItem)
	}
	return fmt.Errorf("Unsupported operation on database!")
}

func insertTodoItem(pItem *TodoItem) error{
	sql := "INSERT INTO todolist VALUES (NULL, ?, ?, ?, ?, ?, ?);"
	r, err := db.Exec(sql, pItem.Theme, pItem.Email, pItem.RemindTime,
					 pItem.LastRemind, pItem.CalendarType, pItem.Content)
	err = handleInfoErr(err, "Failed to add todo item!")
	if err == nil{
		id, e := r.LastInsertId()
		if e == nil{
			pItem.Id = int(id)
		}
		err = e
	}
	return err
}

func deleteTodoItem(item *TodoItem) error{
	sql := "DELETE FROM todolist WHERE id = ? LIMIT 1;"
	_, err := db.Exec(sql, item.Id)
	return handleInfoErr(err, "Failed to delete todo item!")
}

func updateTodoItem(item *TodoItem) error{
	sql := `
	UPDATE todolist
	SET theme=?, email=?, remindTime=?, lastRemind=?, calendarType=?, content=?
	WHERE id = ? LIMIT 1;
	`
	_, err := db.Exec(sql, item.Theme, item.Email, item.RemindTime, item.LastRemind, item.CalendarType, item.Content, item.Id)
	return handleInfoErr(err, "Failed to update todo item!")
}
