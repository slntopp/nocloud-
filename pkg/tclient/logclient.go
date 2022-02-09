package tclient

import (
	"bufio"
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

//Getting container logs
func NewLogger(logger *zap.Logger, follow bool, ts_start uint64) (io.ReadCloser, error) {
	// Developed on base Docker Engine SDKs
	// https://docs.docker.com/engine/api/sdk/examples/
	log := logger.Named("NewLogger")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error("fail newlogger NewClientWithOpts", zap.Error(err))
		return nil, err
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Error("fail newlogger ContainerList", zap.Error(err))
		return nil, err
	}

	container_name := `client`
	for _, container := range containers {
		if strings.Contains(container.Image, container_name) {

			options := types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				// Since:     , //Show logs since timestamp (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)
				Timestamps: true,
				// Follow:     follow,
				// Tail:       "15", //--tail , -n	all	Number of lines to show from the end of the logs,
				// Details: true,The docker logs --details command will add on extra attributes, such as environment variables and labels, provided to --log-opt when creating the container.
			}

			if !follow {
				options.Since = time.Unix(int64(ts_start), 0).Format(time.RFC3339)
			} else {
				options.Follow = true
				options.Tail = "15"

			}
			out, err := cli.ContainerLogs(ctx, container.ID, options)
			if err != nil {
				log.Error("fail newlogger ContainerLogs", zap.Error(err))
				return nil, err
			}

			return out, nil

		}
	}

	log.Error("fail newlogger ContainerLogs", zap.Error(errors.New("container not found"+container_name)))
	return nil, err

}

//Getting and sending to server container logs
func LogConnection(ctx context.Context, logger *zap.Logger, stream pb.SocketConnectionService_LogConnectionClient) error {
	log := logger.Named("LogConnection")

	errgr, ctx := errgroup.WithContext(ctx)

	rc := make(chan io.ReadCloser)
	reader_open := false
	var out io.ReadCloser
	var err_logs error

	errgr.Go(func() (err_ret error) {

		for {
			in, err := stream.Recv()
			err_ret = err
			if err == io.EOF {
				log.Info("Connection closed", zap.Error(err))
				break
			}
			if err != nil {
				log.Error("Failed to receive a note LogConnection", zap.Error(err))
				break
			}

			//SInterrupt docker log stream if open
			if reader_open {
				err_ret = out.Close()
				if err_ret != nil {
					log.Info("Connection to docker logs closed", zap.Error(err_ret))
					break
				}
			}

			if in.Stop {
				log.Debug("Stop log stream",
					zap.Bool("reader_open", reader_open),
					zap.Bool("Follow", in.Follow),
					zap.Skip())

				continue
			}

			out, err_logs = NewLogger(log, in.Follow, in.TsStart)
			if err_logs != nil {
				err_ret = err_logs
				log.Info("Connection to docker logs closed", zap.Error(err_ret))
				break
			}

			reader_open = true
			rc <- out
		}

		return err_ret
	})

	errgr.Go(func() (err_ret error) {
		err_ret = nil
		// defer out.Close() runtime error!
		for {
			select {
			case <-ctx.Done():
				if reader_open {
					out.Close()
				}
				return ctx.Err()
			case out := <-rc:

				go func() {

					reader, writer := io.Pipe()

					go func() {

						scanner := bufio.NewScanner(reader)
						for scanner.Scan() {

							var lcr pb.LogConnectionRequest

							if err_logs != nil {
								lcr = pb.LogConnectionRequest{
									Log: scanner.Text(),
									Err: err_logs.Error(),
								}
							} else {
								lcr = pb.LogConnectionRequest{
									Log: scanner.Text(),
								}
							}

							if err := stream.Send(&lcr); err != nil {
								log.Error("Failed to send a note through LogService", zap.Error(err))
								err_ret = err
								out.Close() //close docker web stream
							}
						}
						log.Debug("Scanner is closed", zap.Bool("reader_open", reader_open), zap.Skip())
					}()

					stdcopy.StdCopy(writer, writer, out)
					writer.Close() //close scanner stream
					reader_open = false
					log.Debug("io.ReadCloser, StdCopy is closed")

				}()

			}
		}
	})

	if err := errgr.Wait(); err != nil {
		log.Error("fail LogService steram connection:", zap.Error(err))
		return err
	}

	return nil
}
