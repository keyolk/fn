package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab-odx.oracle.com/odx/functions/api/models"
	"gitlab-odx.oracle.com/odx/functions/api/runner/common"
	"gitlab-odx.oracle.com/odx/functions/api/runner/task"
)

func getTask(ctx context.Context, url string) (*models.Task, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var task models.Task

	if err := json.Unmarshal(body, &task); err != nil {
		return nil, err
	}

	if task.ID == "" {
		return nil, nil
	}
	return &task, nil
}

func getCfg(t *models.Task) *task.Config {
	cfg := &task.Config{
		Image:   *t.Image,
		ID:      t.ID,
		AppName: t.AppName,
		Env:     t.EnvVars,
	}
	if t.Timeout == nil || *t.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	} else {
		cfg.Timeout = time.Duration(*t.Timeout) * time.Second
	}
	if t.IdleTimeout == nil || *t.IdleTimeout <= 0 {
		cfg.IdleTimeout = DefaultIdleTimeout
	} else {
		cfg.IdleTimeout = time.Duration(*t.IdleTimeout) * time.Second
	}

	return cfg
}

func deleteTask(url string, task *models.Task) error {
	// Unmarshal task to be sent over as a json
	body, err := json.Marshal(task)
	if err != nil {
		return err
	}

	// Send out Delete request to delete task from queue
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	c := &http.Client{}
	if resp, err := c.Do(req); err != nil {
		return err
	} else if resp.StatusCode != http.StatusAccepted {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}
	return nil
}

// RunAsyncRunner pulls tasks off a queue and processes them
func RunAsyncRunner(ctx context.Context, tasksrv string, rnr *Runner, ds models.Datastore) {
	u := tasksrvURL(tasksrv)

	startAsyncRunners(ctx, u, rnr, ds)
	<-ctx.Done()
}

func startAsyncRunners(ctx context.Context, url string, rnr *Runner, ds models.Datastore) {
	var wg sync.WaitGroup
	ctx, log := common.LoggerWithFields(ctx, logrus.Fields{"runner": "async"})
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return

		default:
			if !rnr.hasAsyncAvailableMemory() {
				log.Debug("memory full")
				time.Sleep(1 * time.Second)
				continue
			}
			task, err := getTask(ctx, url)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					log.WithError(err).Errorln("Could not fetch task, timeout.")
					continue
				}
				log.WithError(err).Error("Could not fetch task")
				time.Sleep(1 * time.Second)
				continue
			}
			if task == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			ctx, log := common.LoggerWithFields(ctx, logrus.Fields{"call_id": task.ID})
			log.Debug("Running task:", task.ID)

			wg.Add(1)

			go func() {
				defer wg.Done()
				// Process Task
				_, err := rnr.RunTrackedTask(task, ctx, getCfg(task), ds)
				if err != nil {
					log.WithError(err).Error("Cannot run task")
				}
				log.Debug("Processed task")
			}()

			// Delete task from queue
			if err := deleteTask(url, task); err != nil {
				log.WithError(err).Error("Cannot delete task")
				continue
			}

			log.Info("Task complete")
		}
	}
}

func tasksrvURL(tasksrv string) string {
	parsed, err := url.Parse(tasksrv)
	if err != nil {
		logrus.WithError(err).Fatalln("cannot parse API_URL endpoint")
	}

	if parsed.Scheme == "" {
		parsed.Scheme = "http"
	}

	if parsed.Path == "" || parsed.Path == "/" {
		parsed.Path = "/tasks"
	}

	return parsed.String()
}
