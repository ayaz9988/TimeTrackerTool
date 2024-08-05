package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/user"
	"sync"
	"time"
)

type Task struct {
	ID          int           `json:"id"`
	Title       string        `json:"title"`
	Content     string        `json:"content"`
	CreatedAt   time.Time     `json:"createdAt"`
	ElapsedTime time.Duration `json:"elapsedTime"`
	IsComplete  bool          `json:"isComplete"`
	Running     bool          `json:"alreadyStarted"`
	Timer       *Timer
}

type Tasks struct {
	tasks []Task
	mutex sync.RWMutex
}

func (t *Tasks) Append(task Task) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tasks = append(t.tasks, task)
}

func (t *Tasks) GetAll() []Task {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.tasks
}

func (t *Tasks) Delete(id int) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for i, task := range t.tasks {
		if task.ID == id {
			t.tasks = append(t.tasks[:i], t.tasks[i+1:]...)
			return nil
		}
	}
	return errors.New("task not found")
}

var dataFile string

func SetDataFile(filenames ...string) error {
	var filename string
	if len(filenames) == 0 {
		filename = ".tasks.json"
	} else {
		filename = filenames[0]
	}
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %w", err)
	}
	dataFile = fmt.Sprintf("%s/%s", usr.HomeDir, filename)
	return nil
}

func CreateTask(tasks *Tasks, title, content string) {
	task := Task{
		ID:         len(tasks.tasks) + 1,
		Title:      title,
		Content:    content,
		CreatedAt:  time.Now(),
		Timer:      NewTimer(),
		IsComplete: false,
		Running:    false,
	}
	tasks.Append(task)
}

func ListTasks(tasks *Tasks) <-chan Task {
	ch := make(chan Task)
	go func() {
		for _, task := range tasks.GetAll() {
			ch <- task
		}
		close(ch)
	}()
	return ch
}

func UpdateTask(task *Task, title, content string) {
	task.Title = title
	task.Content = content
	task.CreatedAt = time.Now()
}

func DeleteTask(tasks *Tasks, taskID int) (*Tasks, error) {
	newTasks := &Tasks{}
	for _, task := range tasks.GetAll() {
		if task.ID != taskID {
			newTasks.Append(task)
		}
	}
	if len(newTasks.GetAll()) == len(tasks.GetAll()) {
		return &Tasks{}, errors.New("the task does not exist")
	}
	return newTasks, nil
}

func TimeOfCreation(task Task) time.Time {
	return task.CreatedAt
}

func CompleteTask(task *Task) {
	task.IsComplete = true
}

func GetTime(task Task) string {
	return task.ElapsedTime.String()
}

func SaveTask(tasks *Tasks) error {
	data, err := json.Marshal(tasks.tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks to JSON: %w", err)
	}
	file, err := os.Create(dataFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func LoadTask() (*Tasks, error) {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Tasks{}, nil // File doesn't exist, return empty tasks
		}
		return nil, fmt.Errorf("failed to read tasks from file: %w", err)
	}
	if len(data) == 0 {
		return &Tasks{}, nil // File is empty, return empty tasks
	}
	var tasks Tasks
	err = json.Unmarshal(data, &tasks.tasks) // Unmarshal into tasks.tasks
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks from JSON: %w", err)
	}
	return &tasks, nil
}
