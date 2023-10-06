package benchtest

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type Resource struct {
	pool             *dockertest.Pool
	databaseResource *dockertest.Resource
	webappResource   *dockertest.Resource
	network          *dockertest.Network
}

func (r *Resource) WebappIPAddress() string {
	webappIp := r.webappResource.GetIPInNetwork(r.network)
	return fmt.Sprintf("http://%s:12345", webappIp)
}

func runContainer(pool *dockertest.Pool, opts *dockertest.RunOptions, retryFunc func(*dockertest.Resource) error) (*dockertest.Resource, error) {
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		return nil, err
	}
	if err := pool.Retry(func() error {
		return retryFunc(resource)
	}); err != nil {
		return nil, err
	}

	return resource, nil
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cwdParts := strings.Split(cwd, "/")

	var projectRootIdx int
	for idx, part := range cwdParts {
		if part == "isucon13" {
			projectRootIdx = idx
			break
		}
	}

	var sliceEnd int
	if projectRootIdx+1 <= len(cwdParts) {
		sliceEnd = projectRootIdx + 1
	} else {
		sliceEnd = projectRootIdx
	}

	return strings.Join(cwdParts[:sliceEnd], "/"), nil
}

func Setup(packageName string) (*Resource, error) {
	baseDir, err := findProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("find projectroot error: %w", err)
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("create pool error: %w", err)
	}
	pool.MaxWait = 30 * time.Second
	if err := pool.Client.Ping(); err != nil {
		return nil, fmt.Errorf("pool ping error: %w", err)
	}

	network, err := pool.CreateNetwork(fmt.Sprintf("isupipe-%s", packageName))
	if err != nil {
		return nil, fmt.Errorf("create network error: %w", err)
	}

	// run database
	log.Println("[dockertest] running database container ...")
	databaseResource, err := runContainer(pool, &dockertest.RunOptions{
		Repository: "mysql/mysql-server",
		Tag:        "8.0.31",
		Name:       fmt.Sprintf("mysql-%s", packageName),
		Env: []string{
			"MYSQL_ROOT_HOST=%",
			"MYSQL_ROOT_PASSWORD=root",
		},
		Mounts: []string{
			strings.Join([]string{
				baseDir + "/webapp/sql/initdb.d",
				"/docker-entrypoint-initdb.d",
			}, ":"),
		},
		Networks: []*dockertest.Network{
			network,
		},
	}, func(resource *dockertest.Resource) error {
		mysql.SetLogger(new(nopLogger))
		db, err := sql.Open("mysql", fmt.Sprintf("isucon:isucon@(localhost:%s)/isupipe", resource.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		if err := db.Ping(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("run database container error: %w", err)
	}
	databaseIp := databaseResource.GetIPInNetwork(network)

	// run webapp
	log.Println("[dockertest] running webapp container ...")
	webappResource, err := runContainer(pool, &dockertest.RunOptions{
		Repository: "isupipe",
		Tag:        "latest",
		Name:       fmt.Sprintf("isupipe-%s", packageName),
		Env: []string{
			fmt.Sprintf("ISUCON13_MYSQL_DIALCONFIG_ADDRESS=%s", databaseIp),
			// fmt.Sprintf("ISUCON13_POWERDNS_HOST=%s", powerDNSIp),
			"ISUCON13_POWERDNS_APIKEY=isudns",
			"ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS=127.0.0.1",
			"ISUCON13_POWERDNS_DISABLED=true",
		},
		Networks: []*dockertest.Network{
			network,
		},
	}, func(resource *dockertest.Resource) error {
		addr := net.JoinHostPort("localhost", resource.GetPort("12345/tcp"))
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/initialize", addr), nil)
		if err != nil {
			return err
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("initialize endpoint responses unexpected status code (%d)", resp.StatusCode)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("run webapp container error: %w", err)
	}

	log.Println("[dockertest] setup completed successfully.")
	return &Resource{
		pool:             pool,
		webappResource:   webappResource,
		databaseResource: databaseResource,
		network:          network,
	}, nil
}

func Teardown(r *Resource) error {
	/*
		if err := r.pool.Purge(r.databaseResource); err != nil {
			return err
		}
		if err := r.pool.Purge(r.webappResource); err != nil {
			return err
		}
		if err := r.network.Close(); err != nil {
			return err
		}
	*/

	return nil
}
