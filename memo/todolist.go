package memo

import "time"

type TodoItem struct{
	Id int
	Theme string
	Email string
	Content string
	RemindTime time.Time
	CalendarType string
	LastRemind time.Time
}

type TodoList []TodoItem

/* GetTodoItems should call at start, and just only once. */
func GetTodoItems() (items TodoList){
	return queryTodoItems()
}

func (ptl *TodoList) AddTodoItem(pItem *TodoItem) error{
	pItem.LastRemind = time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	err := syncDb(DbInsert, pItem)
	if err == nil{
		*ptl = append(*ptl, *pItem)
	}
	return err
}

func (ptl *TodoList) DelTodoItem(pItem *TodoItem) error{
	err := syncDb(DbDelete, pItem)
	if err == nil{
		for i, t := range *ptl{
			if t.Id == pItem.Id{
				*ptl = append((*ptl)[:i], (*ptl)[i+1:]...)
				break
			}
		}
	}
	return err
}

func (tl TodoList) UpdateTodoItem(pItem *TodoItem) error{
	err := syncDb(DbUpdate, pItem)
	if err == nil{
		for i, t := range tl{
			if t.Id == pItem.Id{
				tl[i] = *pItem
				break
			}
		}
	}
	return err
}
