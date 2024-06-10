package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Task ...
type Task struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	Note         string   `json:"note"`
	Applications []string `json:"applications"`
}

var tasks = map[string]Task{
	"1": {
		ID:          "1",
		Description: "Сделать финальное задание темы REST API",
		Note:        "Если сегодня сделаю, то завтра будет свободный день. Ура!",
		Applications: []string{
			"VS Code",
			"Terminal",
			"git",
		},
	},
	"2": {
		ID:          "2",
		Description: "Протестировать финальное задание с помощью Postmen",
		Note:        "Лучше это делать в процессе разработки, каждый раз, когда запускаешь сервер и проверяешь хендлер",
		Applications: []string{
			"VS Code",
			"Terminal",
			"git",
			"Postman",
		},
	},
}

func getLastId(tasks map[string]Task) int {
	var lastId int

	for _, el := range tasks {

		id, _ := strconv.Atoi(el.ID)
		if id > lastId {
			lastId = id
		}
	}

	return lastId
}

// Ниже напишите обработчики для каждого эндпоинта
func getTasks(w http.ResponseWriter, req *http.Request) {

	// сериализуем данные из мапы tasks
	resp, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		fmt.Println(err)
	}
}

func addTask(w http.ResponseWriter, req *http.Request) {
	var task Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// проверяем что id не пуст, если пустой, то определяем максимальный и прибавляем 1
	if task.ID == "" {
		newId := getLastId(tasks)
		task.ID = strconv.Itoa(newId + 1)
	}

	// если Applications пусто, то добавляем user agent
	if len(task.Applications) > 0 {
		task.Applications = append(task.Applications, req.UserAgent())
	}

	// проверка что такого элемента нет в мапе
	_, ok := tasks[task.ID]
	if !ok {
		tasks[task.ID] = task
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// после добавления проверим, появилась ли задача с таким id в мапе
	_, ok = tasks[task.ID]
	if !ok {
		http.Error(w, "Артист не найден", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func getTask(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	// смотрим, есть ли задача с таким id в мапе
	task, ok := tasks[id]
	if !ok {
		http.Error(w, "Артист не найден", http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func deleteTask(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	// смотрим, есть ли задача с таким id в мапе
	_, ok := tasks[id]
	if !ok {
		http.Error(w, "Артист не найден", http.StatusBadRequest)
		return
	}

	// удаляем значение из мапы
	delete(tasks, id)

	// смотрим, осталась ли задача с таким id в мапе (это в задании требование такое)
	_, ok = tasks[id]
	if ok {
		http.Error(w, "Артист не найден", http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func main() {
	r := chi.NewRouter()

	// здесь регистрируйте ваши обработчики
	r.Get("/tasks", getTasks)
	r.Post("/tasks", addTask)
	r.Get("/tasks/{id}", getTask)
	r.Delete("/tasks/{id}", deleteTask)

	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}
