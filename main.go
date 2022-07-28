package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	cron "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

type job struct {
	schedule string
	command  string
	args     []string
}

func execute(job job) {

	fmt.Printf("schedule: [%s] executing: %s %s\n", job.schedule, job.command, strings.Join(job.args, " "))

	cmd := exec.Command(job.command, job.args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("fatal: %s\n", err.Error())
	}
}

func parse(args []string) []job {
	var jobs []job
	var commands [][]string
	var lastSeparator int
	for i, arg := range args {
		if arg == "--" {
			commands = append(commands, args[lastSeparator:i])
			lastSeparator = i + 1
		} else if i == len(args)-1 {
			commands = append(commands, args[lastSeparator:])
		}
	}

	for _, command := range commands {
		if len(command) <= 1 {
			panic("invalid schedule and command: " + strings.Join(command, " "))
		}
		jobs = append(jobs, job{
			schedule: command[0],
			command:  command[1],
			args:     command[2:],
		})
	}

	return jobs
}

func create(jobs []job) (cr *cron.Cron, wgr *sync.WaitGroup) {

	wg := &sync.WaitGroup{}

	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))

	for _, job := range jobs {
		if err := createCronJob(c, wg, job); err != nil {
			panic(err.Error())
		}
	}

	return c, wg
}

func createCronJob(c *cron.Cron, wg *sync.WaitGroup, job job) error {
	_, err := c.AddFunc(job.schedule, func() {
		wg.Add(1)
		execute(job)
		wg.Done()
	})

	if err != nil {
		return fmt.Errorf("invalid schedule [%s]: %w", job.schedule, err)
	}

	return nil
}

func start(c *cron.Cron) {
	c.Start()
}

func stop(c *cron.Cron, wg *sync.WaitGroup) {
	println("Stopping")
	c.Stop()
	println("Waiting")
	wg.Wait()
	println("Exiting")
	os.Exit(0)
}

func main() {

	rootCmd := &cobra.Command{
		Use:   "cron -- [schedule1] [command1] [args1...] -- [schedule2] [command2] [args2...]",
		Short: "example: cron -- \"* * * * * *\" echo 'hello'",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			jobs := parse(args)
			c, wg := create(jobs)

			go start(c)

			ch := make(chan os.Signal)
			signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			println(<-ch)

			stop(c, wg)
		},
	}

	err := rootCmd.Execute()
	if err != nil {
		panic(err.Error())
	}
}
