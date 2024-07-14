package backend

type Timer struct {
	started   time.Time
	paused    bool
	stopped   bool
	cancel    chan bool
	timerChan chan time.Duration
}


func NewTimer() *Timer {
	return &Timer{
		cancel:    make(chan bool),
		timerChan: make(chan time.Duration),
	}
}

func (t *Timer) Start() {
	t.started = time.Now()
	t.paused = false
	t.stopped = false
	go t.runTimer()
}

func (t *Timer) Pause() {
	t.paused = true
}

func (t *Timer) Resume() {
	t.paused = false
}

func (t *Timer) Cancel() {
	t.stopped = true
	t.cancel <- true
}

func (t *Timer) runTimer() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !t.paused && !t.stopped {
				elapsed := time.Since(t.started)
				t.timerChan <- elapsed
			}
		case <-t.cancel:
			return
		}
	}
}

func StartTimer(t Task, stopChan chan bool) {
	t.Timer.Start()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case elapsed := <-t.Timer.timerChan:
				t.ElapsedTime = elapsed
			case <-stopChan:
				t.Timer.Cancel()
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()
}

func (t *Timer) GetElapsedTime() time.Duration {
	return <-t.timerChan
}

func StopTimer(t Task) {
	t.Timer.Cancel()
}
