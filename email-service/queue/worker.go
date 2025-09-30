package queue

import (
	"email-service/utils"
	"time"

	"github.com/sirupsen/logrus"
)

type Worker struct {
	ID       int
	JobQueue chan EmailJob
	Quit     chan bool
}

func NewWorker(id int) *Worker {
	return &Worker{
		ID:       id,
		JobQueue: make(chan EmailJob),
		Quit:     make(chan bool),
	}
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case job := <-w.JobQueue:
				w.processJob(job)
			case <-w.Quit:
				logrus.Infof("Worker %d stopping", w.ID)
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.Quit <- true
}

func (w *Worker) processJob(job EmailJob) {
	logrus.Infof("Worker %d processing job %s", w.ID, job.ID)

	// Rate limiting check
	if !CheckRateLimit(job.To, 10, time.Hour) {
		logrus.Warnf("Rate limit exceeded for %s", job.To)
		RequeueEmail(job)
		return
	}

	var err error
	data := utils.EmailData{
		Name:       getString(job.Data, "name"),
		Email:      job.To,
		OTP:        getString(job.Data, "otp"),
		AppURL:     getString(job.Data, "app_url"),
		ProfileURL: getString(job.Data, "profile_url"),
	}

	switch job.Type {
	case "otp":
		err = utils.SendOTPEmail(data.Email, data.Name, data.OTP)
	case "welcome":
		err = utils.SendWelcomeEmail(data.Email, data.Name)
	case "profile_reminder":
		err = utils.SendProfileReminderEmail(data.Email, data.Name)
	default:
		err = utils.SendEmail(job.To, job.Subject, job.TemplateName, data)
	}

	if err != nil {
		logrus.Errorf("Worker %d failed to process job %s: %v", w.ID, job.ID, err)
		RequeueEmail(job)
	} else {
		logrus.Infof("Worker %d completed job %s", w.ID, job.ID)
	}
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

type WorkerPool struct {
	Workers    []*Worker
	JobQueue   chan EmailJob
	NumWorkers int
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		Workers:    make([]*Worker, numWorkers),
		JobQueue:   make(chan EmailJob, 100),
		NumWorkers: numWorkers,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.NumWorkers; i++ {
		worker := NewWorker(i + 1)
		wp.Workers[i] = worker
		worker.Start()
	}

	// Job dispatcher
	go func() {
		for {
			job, err := DequeueEmail()
			if err != nil {
				logrus.Error("Failed to dequeue email:", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if job != nil {
				wp.JobQueue <- *job
			}
		}
	}()

	// Job distributor
	go func() {
		for job := range wp.JobQueue {
			for _, worker := range wp.Workers {
				select {
				case worker.JobQueue <- job:
					goto nextJob
				default:
					continue
				}
			}
			// If all workers are busy, requeue the job
			RequeueEmail(job)
		nextJob:
		}
	}()

	logrus.Infof("Worker pool started with %d workers", wp.NumWorkers)
}

func (wp *WorkerPool) Stop() {
	for _, worker := range wp.Workers {
		worker.Stop()
	}
	close(wp.JobQueue)
}