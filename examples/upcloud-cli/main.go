package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"git.sr.ht/~yoink00/goflenfig"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
)

var username string
var password string

func init() {
	goflenfig.Prefix("UPCLOUD_")
	goflenfig.StringVar(&username, "username", "", "UpCloud user name")
	goflenfig.StringVar(&password, "password", "", "UpCloud password")
}

func main() {
	os.Exit(run())
}

func run() int {
	goflenfig.Parse()

	command := flag.Arg(0)

	if len(username) == 0 {
		fmt.Fprintln(os.Stderr, "Username must be specified")
		return 1
	}

	if len(password) == 0 {
		fmt.Fprintln(os.Stderr, "Password must be specified")
		return 2
	}

	fmt.Println("Creating new client")
	c := client.New(username, password)
	s := service.New(c)

	switch command {
	case "deleteservers":
		if err := deleteServers(s); err != nil {
			return 1
		}
	case "deletestorage":
		if err := deleteStorage(s); err != nil {
			return 2
		}
	default:
		fmt.Fprintln(os.Stderr, "Unknown command: ", command)
		return 99
	}

	return 0
}

func deleteServers(s *service.Service) error {
	fmt.Println("Getting servers")
	servers, err := s.GetServers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get servers: %#v\n", err)
		return err
	}

	fmt.Printf("Retrieved %d servers\n", len(servers.Servers))

	if len(servers.Servers) > 0 {
		fmt.Println("Deleting all servers")
		for _, server := range servers.Servers {
			if server.State != upcloud.ServerStateStopped {
				fmt.Printf("Server %s (%s) is not stopped. Stopping\n", server.Title, server.UUID)
				_, err := s.StopServer(&request.StopServerRequest{
					UUID:     server.UUID,
					StopType: request.ServerStopTypeHard,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Unable to stop server: %#v\n", err)
					return err
				}
			}
			fmt.Printf("Deleting %s (%s)\n", server.Title, server.UUID)
			err := s.DeleteServerAndStorages(&request.DeleteServerAndStoragesRequest{
				UUID: server.UUID,
			})

			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to delete server: %#v\n", err)
				return err
			}
			fmt.Printf("Successfully deleted %s (%s)\n", server.Title, server.UUID)
		}
	}

	return nil
}

func deleteStorage(s *service.Service) error {
	fmt.Println("Getting storage")
	storages, err := s.GetStorages(&request.GetStoragesRequest{
		Access: upcloud.StorageAccessPrivate,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get storages: %#v\n", err)
		return err
	}

	fmt.Printf("Retrieved %d storages\n", len(storages.Storages))

	if len(storages.Storages) > 0 {
		fmt.Println("Deleting all storages")
		for _, storage := range storages.Storages {
			err := errors.New("Dummy")
			for i := 0; err != nil && i < 5; i++ {
				fmt.Printf("%d: Deleting %s (%s)\n", i, storage.Title, storage.UUID)
				err = s.DeleteStorage(&request.DeleteStorageRequest{
					UUID: storage.UUID,
				})
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to delete storage: %#v (%s)\n", err, err.Error())
				return err
			}

			fmt.Printf("Successfully deleted %s (%s)\n", storage.Title, storage.UUID)
		}
	}

	return nil
}
