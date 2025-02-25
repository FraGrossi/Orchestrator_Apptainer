package main

//TODO > controllare https://github.com/apptainer/apptainer/tree/main/examples/plugins/cli-plugin
//				   	 https://pkg.go.dev/github.com/apptainer/apptainer/pkg/plugin#Callback

//TODO > fare in modo che aprendo una shell si ottenga come dir di partenza /home/nextpyter

//TODO > capire come ricevere in JSON il valore in coppia di dirHost/dirCont -> utile per risolvere
//		 StartBind    string   `json:"volume_name"`
//		 Binds        []Pairs  `json:"binds"`

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"errors"
	"net/http"
)

var debug_messages = true

var Orchestrator = ApptainerOrchestrator{
	filePath:        os.Getenv("ORCHESTRATOR_PATH"),
	ContainerConfig: make(map[string]ContainerConfig),
	ContainerOption: make(map[string]ContainerOption),
	ContainerData:   make(map[string]string),
}

func main() {
	debug_messages = flagCheck()

	debugMessage(Orchestrator.filePath)

	orchestratorLoad()

	mux := http.NewServeMux()
	mux.HandleFunc("/start", start_notebook_container)
	mux.HandleFunc("/stop", stop_notebook_container)
	mux.HandleFunc("/restart", restart_notebook_container)
	mux.HandleFunc("/getcont", get_notebook_container)
	mux.HandleFunc("/getallcont", get_all_notebook_containers_paginated)
	mux.HandleFunc("/deletecont", delete_notebook_container)
	mux.HandleFunc("/deletecontdata", delete_notebook_container_data)
	mux.HandleFunc("/createvolume", create_notebook_volume)
	mux.HandleFunc("/removevolume", remove_notebook_volume)

	//Salvataggio Orchestrator in caso di uscita manuale
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Println()
		infoMessage("Chiusura Server:" + fmt.Sprint(sig))

		infoMessage("Salvataggio Orchestrator in [" + Orchestrator.filePath + "]")
		orchestratorSave()

		os.Exit(0)
	}()

	err := http.ListenAndServe(":3333", mux)
	if errors.Is(err, http.ErrServerClosed) {
		errorMessage("Server closed")
	} else if err != nil {
		errorMessage("Error starting server: " + fmt.Sprint(err))
		exit4error()
	}
}
