package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	backend "github.com/ayaz9988/TimeTrackerTool/Backend"
)

func main() {
	a := app.New()
	w := a.NewWindow("Time Tracker")

	tasks := &backend.Tasks{}
	backend.SetDataFile()

	// Load existing tasks
	if loadedTasks, err := backend.LoadTask(); err == nil {
		tasks = loadedTasks
	} else {
		log.Println("Could not load tasks:", err)
	}

	var selectedIndex int // Declare selectedIndex here
	var Running bool

	taskList := widget.NewList(
		func() int {
			return len(tasks.GetAll())
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("") // Create a new label for each item
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			task := tasks.GetAll()[i]
			elapsedTime := task.ElapsedTime.String()
			// Assert the type safely
			label := o.(*widget.Label)
			var ch rune
			if task.IsComplete {
				ch = '🗸'
			} else {
				ch = '✗'
			}
			label.SetText(fmt.Sprintf("[%c] %s    Time: %s | CreatedAt: %v", ch, task.Title, elapsedTime, task.CreatedAt.Format("2006/01/02 15:04")))
		},
	)

	// Set OnSelected event handler once
	taskList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = int(id) // Update selectedIndex
		Running = false
	}

	addButton := widget.NewButton("Add Task", func() {
		titleEntry := widget.NewEntry()
		contentEntry := widget.NewMultiLineEntry()

		dialog.ShowForm("New Task", "Add", "Cancel", []*widget.FormItem{
			widget.NewFormItem("Title", titleEntry),
			widget.NewFormItem("Content", contentEntry),
		}, func(response bool) { // Correct signature for the callback
			if response {
				title := titleEntry.Text
				content := contentEntry.Text
				backend.CreateTask(tasks, title, content)
				taskList.Refresh()
				backend.SaveTask(tasks) // Save after adding
			}
		}, w)
	})

	startTimerButton := widget.NewButton("Start Timer", func() {
		if selectedIndex >= 0 && !Running {
			task := &tasks.GetAll()[selectedIndex]
			stopChan := make(chan bool)
			Running = !Running

			go backend.StartTimer(*task, stopChan)

			//dialog.ShowInformation("Timer Started", fmt.Sprintf("Timer started for task: %s", task.Title), w)

			// Update the elapsed time in the UI
			go func() {
				for {
					select {
					case <-stopChan:
						return
					default:
						task.ElapsedTime = task.Timer.GetElapsedTime()
						taskList.Refresh()
					}
				}
			}()
		}
	})

	stopTimerButton := widget.NewButton("Stop Timer", func() {
		if selectedIndex >= 0 && !Running {
			task := &tasks.GetAll()[selectedIndex]
			backend.StopTimer(*task)
			taskList.Refresh()
			backend.SaveTask(tasks) // Save after stopping
			//dialog.ShowInformation("Timer Stopped", fmt.Sprintf("Timer stopped for task: %s", task.Title), w)
			Running = !Running
		}
	})

	updateButton := widget.NewButton("Update Task", func() {
		if selectedIndex >= 0 {
			task := tasks.GetAll()[selectedIndex]
			titleEntry := widget.NewEntry()
			contentEntry := widget.NewMultiLineEntry()
			titleEntry.SetText(task.Title)
			contentEntry.SetText(task.Content)

			dialog.ShowForm("Update Task", "Update", "Cancel", []*widget.FormItem{
				widget.NewFormItem("Title", titleEntry),
				widget.NewFormItem("Content", contentEntry),
			}, func(response bool) { // Correct signature for the callback
				if response {
					newTitle := titleEntry.Text
					newContent := contentEntry.Text
					backend.UpdateTask(task, tasks, newTitle, newContent)
					taskList.Refresh()
					backend.SaveTask(tasks) // Save after updating
				}
			}, w)
		}
	})

	completeButton := widget.NewButton("Complete Task", func() {
		if selectedIndex >= 0 {
			task := &tasks.GetAll()[selectedIndex]
			if !task.IsComplete {
				backend.CompleteTask(task, true)
			} else {
				backend.CompleteTask(task, false)
			}
			taskList.Refresh()
			backend.SaveTask(tasks) // Save after completing
		}
	})

	deleteButton := widget.NewButton("Delete Task", func() {
		if selectedIndex >= 0 {
			taskID := tasks.GetAll()[selectedIndex].ID
			if err := tasks.Delete(taskID); err != nil {
				dialog.ShowError(err, w)
			}
			taskList.Refresh()
			backend.SaveTask(tasks) // Save after deleting
		}
	})

	con1 := container.NewHBox(addButton, startTimerButton, stopTimerButton, updateButton, completeButton, deleteButton)
	con2 := container.NewScroll(taskList)
	con2.SetMinSize(fyne.NewSize(300, 200))
	content := container.NewVBox(con1, con2)
	w.SetContent(content)
	w.ShowAndRun()

	// Save tasks on exit
	defer func() {
		if err := backend.SaveTask(tasks); err != nil {
			log.Println("Could not save tasks:", err)
		}
	}()
}
