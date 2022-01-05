package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

const containerImage = "ghcr.io/algoattacker/judgeimage-clang:main"

func main() {
	app := fiber.New(fiber.Config{
		Prefork: true,
		AppName: "AlgoAttacker/JudgeServer",
	})

	app.Use(logger.New())

	cli, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		panic(err)
	}

	app.Get("/problems/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		exist := exists("./problems/" + id)

		if !exist {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"content": "Problem not found",
			})
		}

		content, err := ioutil.ReadFile("./problems/" + id + "/PROBLEM.md")

		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"content": "Problem not found",
			})
		}

		return c.Status(200).JSON(fiber.Map{
			"success": true,
			"content": string(content),
		})
	})

	app.Post("/problems/:id", func(c *fiber.Ctx) error {
		body := c.Body()
		id := c.Params("id")
		ctx := context.Background()
		exist := exists("./problems/" + id)

		if !exist {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"content": "Problem not found",
			})
		}

		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"content": "Problem not found",
			})
		}

		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: containerImage,
		}, nil, nil, nil, "")

		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "container creation failed",
			})
		}

		err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "container start failed",
			})
		}

		judgeFilesArchive, err := createTarIncludesFolder("./problems/" + id + "/judge")
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "judge files archive failed",
			})
		}

		err = cli.CopyToContainer(ctx, resp.ID, "/judge", judgeFilesArchive, types.CopyToContainerOptions{})
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "judge files copy failed",
			})
		}

		sourceFileArchive, err := createTarIncludesSource("source.c", body)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "source file archive failed",
			})
		}

		err = cli.CopyToContainer(ctx, resp.ID, "/", sourceFileArchive, types.CopyToContainerOptions{})
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "source file copy failed",
			})
		}

		judgeSecret := randomString(32)

		resp2, err := cli.ContainerExecCreate(ctx, resp.ID, types.ExecConfig{
			Cmd:          []string{"/solver.sh", judgeSecret, "3"},
			AttachStdout: true,
			AttachStderr: true,
		})
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "exec create failed",
			})
		}

		resp3, err := cli.ContainerExecAttach(ctx, resp2.ID, types.ExecStartCheck{})
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "exec attach failed",
			})
		}

		stdout := new(bytes.Buffer)

		_, err = stdcopy.StdCopy(stdout, stdout, resp3.Reader)
		if err != nil {
			return c.Status(502).JSON(fiber.Map{
				"success": false,
				"content": "exec attach failed",
			})
		}

		c.Status(200).JSON(fiber.Map{
			"success": true,
			"content": stdout.String(),
			"secret":  judgeSecret,
		})

		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
			Force: true,
		})

		return nil
	})

	if port, ok := os.LookupEnv("PORT"); ok {
		app.Listen(port)
	} else {
		app.Listen(":8080")
	}
}
