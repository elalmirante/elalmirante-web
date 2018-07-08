package main

import (
	"strings"
	"sync"

	"github.com/elalmirante/elalmirante/models"
	"github.com/elalmirante/elalmirante/providers"
)

type serverResult struct {
	Name   string
	Output string
	Error  error
}

func deployServers(servers []models.Server, ref string) []serverResult {
	results := make([]serverResult, len(servers))

	var wg sync.WaitGroup
	for i, s := range servers {
		wg.Add(1)

		go func(s models.Server, i int) {
			defer wg.Done()

			provider := providers.GetProvider(s.Provider)
			out, err := provider.Deploy(s, ref)

			htmlOut := strings.Replace(out, "\n", "<br />", -1)
			results[i] = serverResult{Name: s.Name, Output: htmlOut, Error: err}
		}(s, i)
	}

	wg.Wait()
	return results
}
