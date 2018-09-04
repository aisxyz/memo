package memo

import (
	"fmt"
	"sync"
	"time"
)

type Monitor struct{
	TodoItems TodoList
	ToAdd chan *TodoItem
	ToDel chan *TodoItem
	ToUpdate chan *TodoItem
	Err chan error
	cache map[int]*TodoItem
	cacheRange time.Duration
	timer *time.Timer
	nearest time.Time
	sync.Mutex
}

func NewMonitor() *Monitor{
	m := Monitor{
		TodoItems: GetTodoItems(),
		ToAdd: make(chan *TodoItem, 10),
		ToDel: make(chan *TodoItem, 10),
		ToUpdate: make(chan *TodoItem, 10),
		Err: make(chan error, 30),
		cache: make(map[int]*TodoItem),
		cacheRange: 24 * time.Hour,		// cache within one day
	}
	m.initCache()
	return &m
}

func (m *Monitor) initCache(){
	for i, item := range m.TodoItems{
		if m.isCacheable(item){
			m.cache[item.Id] = &m.TodoItems[i]	// Note: can't save as &item
		}
	}
}

func (m *Monitor) Remind(item TodoItem) error{
	/* TODO */
	fmt.Println("***", item.RemindTime, item.Email, item.Content)
	return nil
}

func (m *Monitor) Start(){
	go m.loop()			// Set up timer
	go func(){
		for{
			var err error
			select{
			case item := <-m.ToAdd:
				err = m.addTodoItem(item)
			case item := <-m.ToDel:
				err = m.delTodoItem(item)
			case item := <-m.ToUpdate:
				err = m.updateTodoItem(item)
			}
			if err != nil{
				m.Err<- err
			}
		}
	}()
}

func (m *Monitor) loop(){
	var wg sync.WaitGroup
	m.Lock()
	for _, pItem := range m.cache{
		wg.Add(1)
		go func(item TodoItem){
			defer wg.Done()
			if m.isArriveTime(item){
				if err := m.Remind(item); err != nil{
					m.Err<- err
					return		// Remind again next time
				}
				m.Lock()
				delete(m.cache, item.Id)
				if ! m.isCyclicalTodo(item){
					if err := m.TodoItems.DelTodoItem(&item); err != nil{
						m.Err<- err
					}
				}else{
					item.LastRemind = time.Now()
					m.TodoItems.UpdateTodoItem(&item)
				}
				m.Unlock()
			}
		}(*pItem)
	}
	m.Unlock()
	wg.Wait()
	/* Calculate the nearest time for next loop. */
	now := time.Now()
	nearest := now.Add(m.cacheRange)
	m.Lock()
	if len(m.cache) == 0{
		m.initCache()
	}
	for _, pItem := range m.cache{
		t := m.tuneCycleTime(*pItem)
		if t.Before(nearest){
			nearest = t
		}
	}
	m.nearest = nearest
	m.timer = time.AfterFunc( nearest.Sub(now), m.loop) // timer
	m.Unlock()
}

func (m *Monitor) addTodoItem(item *TodoItem) error{
	m.Lock()
	defer m.Unlock()
	err := m.TodoItems.AddTodoItem(item)
	if err == nil && m.isCacheable(*item){
		m.cache[item.Id] = item
		m.tuneTimer(*item)
	}
	return err
}

func (m *Monitor) delTodoItem(item *TodoItem) error{
	m.Lock()
	defer m.Unlock()
	err := m.TodoItems.DelTodoItem(item)
	if err == nil{
		delete(m.cache, item.Id)
	}
	return err
}

func (m *Monitor) updateTodoItem(item *TodoItem) error{
	m.Lock()
	defer m.Unlock()
	m.tuneTimer(*item)
	return m.TodoItems.UpdateTodoItem(item)
}

func (m *Monitor) tuneTimer(item TodoItem){
	t := m.tuneCycleTime(item)
	if t.Before(m.nearest){
		m.nearest = t
		m.timer.Stop()
		m.timer = time.AfterFunc( t.Sub(time.Now()), m.loop)
	}
}

func (m *Monitor) tuneCycleTime(item TodoItem) time.Time{
	if m.isCyclicalTodo(item){
		t := item.RemindTime
		m, d := t.Month(), t.Day()
		now := time.Now()
		ny, nm, nd := now.Year(), now.Month(), now.Day()
		if t.Year() < 2000{
			m = nm
			if m < 12{		// circle by day
				d = nd
			}
		}
		return time.Date(ny, m, d, t.Hour(), t.Minute(), t.Second(), 0, time.Local)
	}
	return item.RemindTime
}

func (m *Monitor) isCyclicalTodo(item TodoItem) bool{
	// yyyy-mm-dd => disposable
	// 2000-mm-dd => circle by year
	// 2000-00-dd => circle by month -> save as: 1999-12-dd
	// 2000-00-00 => circle by day -> save as: 1999-11-30
	return item.RemindTime.Year() <= 2000
}

func (m *Monitor) isCacheable(item TodoItem) bool{
	return time.Now().Sub(item.LastRemind) > m.cacheRange
}

func (m *Monitor) isArriveTime(item TodoItem) bool{
	return ! time.Now().Before( m.tuneCycleTime(item) )
}
