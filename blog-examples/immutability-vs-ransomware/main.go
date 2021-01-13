package main

import (
	"context"
	"github.com/codenotary/immudb/pkg/api/schema"
	immuclient "github.com/codenotary/immudb/pkg/client"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// creates example data if not existent
func SetUp() {
	if _, err := os.Stat("protected-data"); os.IsNotExist(err) {
		err = os.Mkdir("protected-data", 0755)
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create("protected-data/data.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err2 := f.WriteString("No ransom for hackers!\n")

		if err2 != nil {
			log.Fatal(err2)
		}

		log.Println("Data created!")
	}
	log.Println("Data exists!")
}

func main() {
	// invoking data creation
	SetUp()

	client, err := immuclient.NewImmuClient(immuclient.DefaultOptions())
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	// login with default username and password
	lr, err := client.Login(ctx, []byte(`immudb`), []byte(`immudb`))
	if err != nil {
		log.Fatal(err)
	}

	// immudb provides multidatabase capabilities.
	// token is used not only for authentication, but also to route calls to the correct database
	md := metadata.Pairs("authorization", lr.Token)
	ctx = metadata.NewOutgoingContext(context.Background(), md)
	/*
		// creating new database
		err = client.CreateDatabase(ctx, &schema.Database{
			Databasename: "backupdb",
		})
		if err != nil {
			log.Fatal(err)
		}
	*/
	// switch to database
	resp, err := client.UseDatabase(ctx, &schema.Database{
		Databasename: "backupdb",
	})

	md = metadata.Pairs("authorization", resp.Token)
	ctx = metadata.NewOutgoingContext(context.Background(), md)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// filter temporary file names
				res := strings.Contains(event.Name, ".goutputstream")
				if res == false {
					// get current directory
					dir, err := os.Getwd()
					if err != nil {
						log.Fatal(err)
					}
					log.Println(dir)
					// read content of file changed file in directory
					file, err := ioutil.ReadFile(event.Name)
					if err != nil {
						log.Println(err)
					}
					// create key for immudb
					key := []byte(dir +"/"+ event.Name)
					// save content of file to value for immudb
					value := file

					// set key-value pair in immudb
					tx, err2 := client.VerifiedSet(ctx, key, value)
					log.Printf("Set and verified key '%s' with value '%s' at tx %d\n", key, value, tx.Id)
					if err2 != nil {
						log.Fatal(err2)
					}

					// lookup history of file by key
					req := &schema.HistoryRequest{
						Key: []byte(key),
					}

					entries, err := client.History(ctx, req)
					if err != nil {
						log.Fatal(err)
					}
					log.Println("History-Entries:\n", len(entries.GetEntries()), string(req.Key))

				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Op, event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("protected-data")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
