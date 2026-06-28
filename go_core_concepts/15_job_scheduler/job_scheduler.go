package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Job struct {
	Name     string
	Interval time.Duration
	Task     func()
}

type Scheduler struct {
	jobs   []Job
	wg     sync.WaitGroup
	cancel context.CancelFunc
	ctx    context.Context
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Scheduler) AddJob(name string, interval time.Duration, task func()) {
	s.jobs = append(s.jobs, Job{
		Name:     name,
		Interval: interval,
		Task:     task,
	})
}

func (s *Scheduler) Start() {
	fmt.Println("Scheduler started...")
	for _, job := range s.jobs {
		s.wg.Add(1)
		go s.run(job)
	}
}

func (s *Scheduler) run(job Job) {
	defer s.wg.Done()
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	fmt.Printf("[%s] Job scheduled every %v\n", job.Name, job.Interval)

	for {
		select {
		case <-s.ctx.Done():
			fmt.Printf("[%s] Job stopped.\n", job.Name)
			return
		case t := <-ticker.C:
			fmt.Printf("[%s] Running at %s\n", job.Name, t.Format("15:04:05"))
			job.Task()
		}
	}
}

func (s *Scheduler) Stop() {
	fmt.Println("\nStopping scheduler...")
	s.cancel()
	s.wg.Wait()
	fmt.Println("All jobs stopped.")
}

func main() {
	scheduler := NewScheduler()

	scheduler.AddJob("CleanupJob", 2*time.Second, func() {
		fmt.Println("  → Cleaning up temp files...")
	})

	scheduler.AddJob("HeartbeatJob", 3*time.Second, func() {
		fmt.Println("  → Sending heartbeat ping...")
	})

	scheduler.AddJob("ReportJob", 5*time.Second, func() {
		fmt.Println("  → Generating report...")
	})

	scheduler.Start()

	// Run for 10 seconds then stop
	time.Sleep(10 * time.Second)
	scheduler.Stop()
}
