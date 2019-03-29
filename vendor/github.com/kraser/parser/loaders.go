// loaders
package parser

import (
	"container/heap"
	"priceloader"
	"sync"
)

//Загрузчик
type Loader struct {
	task    chan priceloader.LoadTask // канал для заданий
	pending int                       // кол-во оставшихся задач
	index   int                       // позиция в куче
	wg      *sync.WaitGroup           //указатель на группу ожидания
}

func (loader *Loader) work(done chan *Loader) {
	for {
		taskToDo := <-loader.task //читаем следующее задание
		loader.wg.Add(1)          //инкриминируем счетчик группы ожидания
		LoadAndParse(taskToDo)    //загружаем позиции
		loader.wg.Done()          //сигнализируем группе ожидания что закончили
		done <- loader            //показываем что завершили работу
	}
}

//Это будет наша "куча":
type Pool []*Loader

//Проверка кто меньше - в нашем случае меньше тот у кого меньше заданий:
func (p Pool) Less(i, j int) bool { return p[i].pending < p[j].pending }

//Вернем количество рабочих в пуле:
func (p Pool) Len() int { return len(p) }

//Реализуем обмен местами:
func (p Pool) Swap(i, j int) {
	if i >= 0 && i < len(p) && j >= 0 && j < len(p) {
		p[i], p[j] = p[j], p[i]
		p[i].index, p[j].index = i, j
	}
}

//Заталкивание элемента:
func (p *Pool) Push(x interface{}) {
	n := len(*p)
	loader := x.(*Loader)
	loader.index = n
	*p = append(*p, loader)
}

//И выталкивание:
func (p *Pool) Pop() interface{} {
	old := *p
	n := len(old)
	item := old[n-1]
	item.index = -1
	*p = old[0 : n-1]
	return item
}

//Регулятор нагрузки загрузчиков
type LoadController struct {
	Loaders        int
	LoaderCapacity int
	pool           Pool                      //"Куча" загрузчиков
	done           chan *Loader              //Канал уведомления о завершении для загрузчиков
	requests       chan priceloader.LoadTask //Канал для получения новых заданий
	flowctrl       chan bool                 //Канал для PMFC
	queue          int                       //Количество незавершенных заданий переданных загрузчикам
	wg             *sync.WaitGroup           //Группа ожидания для загрузчиков
}

//Инициализируем регулятор. Аргументом получаем канал по которому приходят задания
func (controller *LoadController) init(task chan priceloader.LoadTask) {
	controller.requests = make(chan priceloader.LoadTask)
	controller.flowctrl = make(chan bool)
	controller.done = make(chan *Loader)
	controller.wg = new(sync.WaitGroup)

	//Запускаем наш Flow Control:
	go func() {
		for {
			controller.requests <- <-task //получаем новое задание и пересылаем его на внутренний канал
			<-controller.flowctrl         //а потом ждем получения подтверждения
		}
	}()

	//Инициализируем кучу и создаем загрузчики:
	heap.Init(&controller.pool)
	for i := 0; i < controller.Loaders; i++ {
		load := &Loader{
			task:    make(chan priceloader.LoadTask, controller.LoaderCapacity),
			index:   0,
			pending: 0,
			wg:      controller.wg,
		}
		go load.work(controller.done)     //запускаем рабочего
		heap.Push(&controller.pool, load) //и заталкиваем его в кучу
	}
}

//Рабочая функция регулятора получает аргументом канал уведомлений от главного цикла
func (controller *LoadController) balance(quit chan bool) {
	lastjobs := false //Флаг завершения, поднимаем когда кончились задания
	for {
		select { //В цикле ожидаем коммуникации по каналам:

		case <-quit: //пришло указание на остановку работы
			controller.wg.Wait() //ждем завершения текущих загрузок...
			quit <- true         //..и отправляем сигнал что закончили

		case task := <-controller.requests: //Получено новое задание (от flow controller)
			if task.Message != ENDMESSAGE { //Проверяем - а не кодовая ли это фраза?
				controller.dispatch(task) // если нет, то отправляем загрузчикам
			} else {
				lastjobs = true //иначе поднимаем флаг завершения
			}

		case load := <-controller.done: //пришло уведомление, что загрузчик закончил загрузку
			controller.completed(load) //обновляем его данные
			if lastjobs {
				if load.pending == 0 { //если у загрузчика кончились задания..
					heap.Remove(&controller.pool, load.index) //то удаляем его из кучи
				}
				if len(controller.pool) == 0 { //а если куча стала пуста
					//значит все загрузчики закончили свои очереди
					quit <- true //и можно отправлять сигнал подтверждения готовности к останову
				}
			}
		}
	}
}

// Функция отправки задания
func (controller *LoadController) dispatch(task priceloader.LoadTask) {
	load := heap.Pop(&controller.pool).(*Loader) //Берем из кучи самого незагруженный загрузчик..
	load.task <- task                            //..и отправляем ему задание.
	load.pending++                               //Добавляем ему "весу"..
	heap.Push(&controller.pool, load)            //..и отправляем назад в кучу
	if controller.queue++; controller.queue < controller.Loaders*controller.LoaderCapacity {
		controller.flowctrl <- true
	}
}

//Обработка завершения задания
func (controller *LoadController) completed(load *Loader) {
	load.pending--
	heap.Remove(&controller.pool, load.index)
	heap.Push(&controller.pool, load)
	if controller.queue--; controller.queue == controller.Loaders*controller.LoaderCapacity-1 {
		controller.flowctrl <- true
	}
}
