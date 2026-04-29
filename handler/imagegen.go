package handler

import (
	"context"

	"github.com/moby/moby/client"
)

func HandleImageGendown() error {
	//startet den docker container neu
	timeout := 5

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	ctx := context.Background()

	_, err = cli.ContainerRestart(ctx, "container_name", client.ContainerRestartOptions{
		Timeout: &timeout,
	})

	return err

}
