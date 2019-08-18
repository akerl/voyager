package multi

import (
	"os"
	"strings"
	"time"

	"github.com/akerl/voyager/v2/travel"

	"github.com/akerl/speculate/v2/creds"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

// Processor defines the settings for parallel processing
type Processor struct {
	Grapher      travel.Grapher
	Options      travel.TraverseOptions
	Args         []string
	RoleNames    []string
	ProfileNames []string
}

func (p Processor) ExecString(cmd string) (map[string]creds.ExecResult, error) {
	cmdSlice := strings.Split(cmd, " ")
	return p.Exec(cmdSlice)
}

func (p Processor) Exec(cmd []string) (map[string]creds.ExecResult, error) {
	paths, err := p.Grapher.ResolveAll(p.Args, p.RoleNames, p.ProfileNames)
	if err != nil {
		return map[string]creds.ExecResult{}, err
	}

	inputCh := make(chan workerInput, len(paths))
	outputCh := make(chan workerOutput, len(paths))
	refreshCh := make(chan time.Time)

	for i := 1; i <= 10; i++ {
		go execWorker(i, inputCh, outputCh)
	}

	for _, item := range paths {
		inputCh <- workerInput{
			Path:    item,
			Options: p.Options,
			Command: cmd,
		}
	}
	close(inputCh)

	progress := mpb.New(
		mpb.WithOutput(os.Stderr),
		mpb.WithManualRefresh(refreshCh),
	)
	bar := progress.AddBar(
		int64(len(paths)),
		mpb.AppendDecorators(
			decor.Percentage(),
		),
		mpb.PrependDecorators(
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
	)

	output := map[string]creds.ExecResult{}
	for i := 1; i <= len(paths); i++ {
		result := <-outputCh
		output[result.AccountId] = result.ExecResult
		bar.Increment()
		refreshCh <- time.Now()
	}
	progress.Wait()

	return output, nil
}

type workerInput struct {
	Path    travel.Path
	Options travel.TraverseOptions
	Command []string
}

type workerOutput struct {
	AccountId  string
	ExecResult creds.ExecResult
}

func execWorker(id int, inputCh <-chan workerInput, outputCh chan<- workerOutput) {
	for item := range inputCh {
		c, err := item.Path.TraverseWithOptions(item.Options)
		if err != nil {
			outputCh <- workerOutput{ExecResult: creds.ExecResult{Error: err}}
			continue
		}
		accountId, err := c.AccountID()
		if err != nil {
			outputCh <- workerOutput{ExecResult: creds.ExecResult{Error: err}}
			continue
		}
		outputCh <- workerOutput{
			AccountId:  accountId,
			ExecResult: c.Exec(item.Command),
		}
	}
}
