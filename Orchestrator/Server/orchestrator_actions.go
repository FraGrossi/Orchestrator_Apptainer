package main

import (
	"fmt"
	"net/http"
)

func start_notebook_container(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> start_notebook_container")

	const usage = `Usage: POST request
	Body {container_name, container_image, image_distributor, image_os, image_version, container_command, volume_name}
	REQUIRED: container_name, container_image, image_distributor, image_os, image_version`

	switch r.Method {
	case http.MethodPost:
		pathImg := "/opt/apptainer_container/"

		container_config, errResp := newContainerConfig(r)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		if container_config.Hostname == "" || container_config.Image == "" || container_config.ImageDistr == "" ||
			container_config.ImageOS == "" || container_config.ImageVersion == "" {

			errorMessage("Richiesta incompleta. Creazione container annullata")
			errorResponseJSON(w, http.StatusBadRequest, usage)
			break
		}

		container_config = checkDuplicate(container_config)

		container_config.Image, errResp = imageNameCombiner(pathImg, container_config.Hostname, container_config.Image)
		debugMessage(container_config.Image)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		container_options := newContainerOption(container_config)

		debugMessage(fmt.Sprint(container_config))
		debugMessage(fmt.Sprint(container_options))

		definition_file, errResp := definitionBuilder(container_config)
		debugMessage(definition_file)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		//-------------------------------------------------------

		//creazione immagine
		errResp = imageBuilder(definition_file, container_config.Image)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		//avvio container
		errResp = instanceStart(container_config)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		container_options.Ip = getContainerIP(container_options.ContainerName)

		Orchestrator.ContainerConfig[container_config.Hostname] = container_config
		Orchestrator.ContainerOption[container_config.Hostname] = container_options
		Orchestrator.ContainerData[container_config.Image] = container_config.Hostname
		successMessage("Istanza " + Orchestrator.ContainerConfig[container_config.Hostname].Hostname + " aggiunta all'orchestrator. IP assegnato  [" + Orchestrator.ContainerOption[container_config.Hostname].Ip + "]")
		StartNotebookContainerJSON(w, container_config.Hostname)
	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, usage)
	}
}

func get_notebook_container(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> get_notebook_container")
	switch r.Method {
	case http.MethodGet:
		container_name, errResp := containerNameJSON(r)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		if _, exists := Orchestrator.ContainerConfig[container_name]; !exists {
			errorMessage("Istanza <" + container_name + "> non registrata nell'orchestrator")
			errorResponseJSON(w, http.StatusBadRequest, "Istanza <"+container_name+"> non registrata nell'orchestrator")
			return
		}

		infoMessage("[Container " + container_name + "]")
		infoMessage(">" + fmt.Sprint(Orchestrator.ContainerConfig[container_name]))

		if Orchestrator.ContainerOption[container_name].Active {
			infoMessage(">Active:	Yes")
			infoMessage(">Ip:	" + Orchestrator.ContainerOption[container_name].Ip)
		} else {
			infoMessage(">Active: No")
		}

		GetNotebookContainerJSON(w, container_name)
	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, "Usage: GET request\nBody {container_name}")
	}
}

func get_all_notebook_containers_paginated(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> get_all_notebook_containers_paginated")

	usage := "Usage: GET request\nBody {}"

	switch r.Method {
	case http.MethodGet:

		infoMessage("----[Complete Container List]----")

		for i, item := range Orchestrator.ContainerOption {
			infoMessage("<Istance " + i + ">")
			infoMessage("Hostname:	" + item.ContainerName)
			if item.Active {
				infoMessage(">Active:	Yes")
				infoMessage(">Ip:	" + item.Ip)
			} else {
				infoMessage(">Active:	No")
			}
		}

		infoMessage("------[Container Data List]------")
		for location, hostname := range Orchestrator.ContainerData {
			infoMessage("<Istance " + hostname + ">")
			infoMessage("Location:	" + location)
		}

		GetAllNotebookContainersPaginatedJSON(w)
	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, usage)
	}
}

func stop_notebook_container(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> stop_notebook_container")

	usage := "Usage: POST request\nBody {container_name}"

	switch r.Method {
	case http.MethodPost:

		container_name, errResp := containerNameJSON(r)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		if _, exists := Orchestrator.ContainerConfig[container_name]; !exists {
			errorMessage("Istanza <" + container_name + "> non registrata nell'orchestrator")
			errorResponseJSON(w, http.StatusNotFound, "Istanza <"+container_name+"> non registrata nell'orchestrator")
			return
		}

		container_option := Orchestrator.ContainerOption[container_name]

		if !container_option.Active {
			warningMessage("Container <" + container_name + "> already stopped")
			errorResponseJSON(w, http.StatusConflict, "Container <"+container_name+"> already stopped")
			return
		}

		errResp = instanceStop(container_option.ContainerName)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		container_option.Ip = ""
		container_option.Active = false

		Orchestrator.ContainerOption[container_name] = container_option

		successMessage("Container <" + container_name + "> stopped")

		StopNotebookContainerJSON(w, container_name)
	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, usage)
	}

}

func restart_notebook_container(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> restart_notebook_container")

	usage := "Usage: POST request\nBody {container_name}"

	switch r.Method {
	case http.MethodPost:

		container_name, errResp := containerNameJSON(r)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		if _, exists := Orchestrator.ContainerConfig[container_name]; !exists {
			errorMessage("Istanza <" + container_name + "> non registrata nell'orchestrator")
			errorResponseJSON(w, http.StatusNotFound, "Istanza <"+container_name+"> non registrata nell'orchestrator")
			return
		}

		var container_option = Orchestrator.ContainerOption[container_name]

		if container_option.Active {
			warningMessage("Container " + container_name + " avviato")
			errorResponseJSON(w, http.StatusConflict, "Container <"+container_name+"> already running")
			return
		}

		errResp = instanceStart(Orchestrator.ContainerConfig[container_name])

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		container_option.Ip = getContainerIP(container_option.ContainerName)
		container_option.Active = true

		Orchestrator.ContainerOption[container_name] = container_option

		successMessage("Container " + container_name + " now running")
		RestartNotebookContainerJSON(w, container_name)

	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, usage)
	}
}

func delete_notebook_container(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> delete_notebook_container")

	usage := "Usage: POST request\nBody {container_name}"

	switch r.Method {
	case http.MethodPost:

		container_name, errResp := containerNameJSON(r)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		if _, exists := Orchestrator.ContainerConfig[container_name]; !exists {
			errorMessage("Istanza <" + container_name + "> non registrata nell'orchestrator")
			errorResponseJSON(w, http.StatusNotFound, "Istanza <"+container_name+"> non registrata nell'orchestrator")
			return
		}

		if isContainerActive(container_name) {
			errorMessage("Impossibile rimuovere una istanza attiva dall'orchestrator")
			errorResponseJSON(w, http.StatusUnprocessableEntity, "Impossibile rimuovere una istanza attiva dall'orchestrator")
			return
		}

		delete(Orchestrator.ContainerOption, container_name)
		delete(Orchestrator.ContainerConfig, container_name)

		warningMessage("Consigliato eliminare i dati del <" + container_name + "> prima della creazione di un'altro container")
		successMessage("Eliminato il container <" + container_name + "> dall'orchestrator")

		DeleteNotebookContainerJSON(w)
	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, usage)
	}
}

func create_notebook_volume(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> create_notebook_volume")

	usage := `Usage: POST request\nBody {container_name, host_location, container_location}`

	switch r.Method {
	case http.MethodPost:
		volume_data, errResp := containerVolumeJSON(r)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		//Controllo la presenza dei dati necessari nella richiesta per l'operzione
		if checkBindJson(w, volume_data, usage) {
			return
		}

		if checkPathBind(w, volume_data, usage) {
			return
		}

		//controllo che il container indicato ESISTA
		if _, exists := Orchestrator.ContainerConfig[volume_data.ContainerName]; !exists {
			errorMessage("Istanza <" + volume_data.ContainerName + "> non registrata nell'orchestrator")
			errorResponseJSON(w, http.StatusNotFound, "Container ["+volume_data.ContainerName+"] does not exist")
			return
		}

		//controllo che il bind non esista già
		if dHost, exists := Orchestrator.ContainerConfig[volume_data.ContainerName].Binds[volume_data.HostLocation]; exists {
			errorMessage("Bind della cartella Host:\"" + volume_data.HostLocation + "\" già presente nella directory Container:\"" + dHost + "\"")
			errorResponseJSON(w, http.StatusConflict, "Bind della cartella Host:\""+volume_data.HostLocation+"\" già presente nella directory Container:\""+dHost+"\"")
			return
		}

		//vedo se il container è inattivo o no
		if Orchestrator.ContainerOption[volume_data.ContainerName].Active {
			warningMessage("[Warning] Istanza attiva per il container " + volume_data.ContainerName)
			warningMessage("[Warning] Modifiche visibili dopo il riavvio")
		}

		Orchestrator.ContainerConfig[volume_data.ContainerName].Binds[volume_data.HostLocation] = volume_data.ContainerLocation

		successMessage("Bind tra Host:\"" + volume_data.HostLocation + "\" e Container:\"" + volume_data.ContainerLocation + "\" creato per il container <" + volume_data.ContainerName + ">")
		CreateNotebookVolumeJSON(w, volume_data.ContainerName)
	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, usage)
	}
}

func remove_notebook_volume(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> remove_notebook_volume")

	usage := "Usage: POST request\nBody {container_name, host_location, container_location}"

	switch r.Method {
	case http.MethodPost:
		volume_data, errResp := containerVolumeJSON(r)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		//Controllo la presenza dei dati necessari nella richiesta per l'operzione
		if checkBindJson(w, volume_data, usage) {
			return
		}

		//controllo che il container indicato ESISTA
		if _, exists := Orchestrator.ContainerConfig[volume_data.ContainerName]; !exists {
			errorMessage("Istanza <" + volume_data.ContainerName + "> non registrata nell'orchestrator")
			errorResponseJSON(w, http.StatusNotFound, "Container ["+volume_data.ContainerName+"] does not exist")
			return
		}

		//vedo se il container è inattivo o no
		if Orchestrator.ContainerOption[volume_data.ContainerName].Active {
			errorMessage("Impossibile rimuovere bind con instanza attiva")
			errorResponseJSON(w, http.StatusUnprocessableEntity, "Impossibile rimuovere bind con instanza attiva")
			return
		}

		//controllo che esista il bind indicato dalla richiesta
		if dHost, exists := Orchestrator.ContainerConfig[volume_data.ContainerName].Binds[volume_data.HostLocation]; exists {
			if dHost != volume_data.ContainerLocation {
				errorMessage("Bind della cartella Host:\"" + volume_data.HostLocation + "\" esistente ma non nella location Container:\"" + volume_data.ContainerLocation + "\"")
				errorResponseJSON(w, http.StatusBadRequest, "Bind della cartella Host:\""+volume_data.HostLocation+"\" esistente ma non nella location Container:\""+volume_data.ContainerLocation+"\"")
				return
			}
		} else {
			errorMessage("Bind della cartella Host:\"" + volume_data.HostLocation + "\" nel container <" + volume_data.ContainerName + "> non esistente")
			errorResponseJSON(w, http.StatusNotFound, "Bind della cartella Host:\""+volume_data.HostLocation+"\" nel container <"+volume_data.ContainerName+"> non esistente")
			return
		}

		delete(Orchestrator.ContainerConfig[volume_data.ContainerName].Binds, volume_data.HostLocation)

		successMessage("Bind tra Host:\"" + volume_data.HostLocation + "\" e Container:\"" + volume_data.ContainerLocation + "\" rimosso per il container <" + volume_data.ContainerName + ">")
		RemoveNotebookVolumeJSON(w, volume_data.ContainerName)
	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, usage)
	}
}

func delete_notebook_container_data(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	infoMessage("[REQUEST] -> delete_notebook_container_data")

	usage := "Usage: POST request\nBody {container_name}"

	switch r.Method {
	case http.MethodPost:
		delete_path, errResp := containerNameJSON(r)

		if errResp.Status != 0 {
			errorResponseJSON(w, errResp.Status, errResp.Message)
			return
		}

		if isContainerActive(delete_path) {
			errorMessage("Impossibile eliminari i dati di un container attivo")
			errorResponseJSON(w, http.StatusUnprocessableEntity, "Impossibile eliminari i dati di un container attivo")
			return
		}

		if _, exists := Orchestrator.ContainerConfig[delete_path]; exists {
			errorMessage("Impossibile eliminari i dati di un container registrato all'orchestrator")
			errorResponseJSON(w, http.StatusUnprocessableEntity, "Impossibile eliminari i dati di un container registrato all'orchestrator")
			return
		}

		for path, hostname := range Orchestrator.ContainerData {
			if hostname == delete_path {
				infoMessage("Dati trovati in " + path + " -> Eliminazione in corso...")

				if !deleteContainerData(path) {
					errorMessage("Dati non eliminati")
					errorResponseJSON(w, http.StatusInternalServerError, "An internal server error occurred while attempting to delete the data")
					return
				}

				delete(Orchestrator.ContainerData, path)

				successMessage("Dati eliminati")
				DeleteNotebookContainerDataJSON(w)
				return
			}
		}

		errorMessage("Dati di un container <" + delete_path + "> non trovati")
		errorResponseJSON(w, http.StatusNotFound, "Dati di un container <"+delete_path+"> non trovati")
	default:
		errorMessage("Richiesta " + r.Method + " non supportata")
		errorResponseJSON(w, http.StatusMethodNotAllowed, usage)
	}

}
