package main

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type StartNotebookContainerResponse struct {
	Status          int    `json:"status"`
	ContainerName   string `json:"container_name"`
	ContainerActive bool   `json:"container_active"`
	ContainerIp     string `json:"container_ip"`
}

type GetNotebookContainerResponse struct {
	Status          int               `json:"status"`
	ContainerName   string            `json:"container_name"`
	ContainerActive bool              `json:"container_active"`
	ContainerIp     string            `json:"container_ip"`
	Image           string            `json:"container_image"`
	ImageDistr      string            `json:"image_distributor"`
	ImageOS         string            `json:"image_os"`
	ImageVersion    string            `json:"image_version"`
	Binds           map[string]string `json:"binds"`
}

type StopNotebookContainerResponse struct {
	Status          int    `json:"status"`
	ContainerName   string `json:"container_name"`
	ContainerActive bool   `json:"container_active"`
}

type NotebookVolumeResponse struct {
	Status        int               `json:"status"`
	ContainerName string            `json:"container_name"`
	Binds         map[string]string `json:"binds"`
}

func errorResponseJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := ErrorResponse{
		Status:  status,
		Message: message,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func StartNotebookContainerJSON(w http.ResponseWriter, container_name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	startNotebookContainer := StartNotebookContainerResponse{
		Status:          http.StatusCreated,
		ContainerName:   container_name,
		ContainerActive: Orchestrator.ContainerOption[container_name].Active,
		ContainerIp:     Orchestrator.ContainerOption[container_name].Ip,
	}

	json.NewEncoder(w).Encode(startNotebookContainer)
}

func GetNotebookContainerJSON(w http.ResponseWriter, container_name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	getNotebookContainer := GetNotebookContainerResponse{
		Status:          http.StatusOK,
		ContainerName:   container_name,
		ContainerActive: Orchestrator.ContainerOption[container_name].Active,
		ContainerIp:     Orchestrator.ContainerOption[container_name].Ip,
		Image:           Orchestrator.ContainerConfig[container_name].Image,
		ImageDistr:      Orchestrator.ContainerConfig[container_name].ImageDistr,
		ImageOS:         Orchestrator.ContainerConfig[container_name].ImageOS,
		ImageVersion:    Orchestrator.ContainerConfig[container_name].ImageVersion,
		Binds:           Orchestrator.ContainerConfig[container_name].Binds,
	}

	json.NewEncoder(w).Encode(getNotebookContainer)
}

func GetAllNotebookContainersPaginatedJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var getAllNotebookContainersPaginated []GetNotebookContainerResponse

	for container_name, options := range Orchestrator.ContainerOption {
		container := GetNotebookContainerResponse{
			Status:          http.StatusOK,
			ContainerName:   container_name,
			ContainerActive: options.Active,
			ContainerIp:     options.Ip,
			Image:           Orchestrator.ContainerConfig[container_name].Image,
			ImageDistr:      Orchestrator.ContainerConfig[container_name].ImageDistr,
			ImageOS:         Orchestrator.ContainerConfig[container_name].ImageOS,
			ImageVersion:    Orchestrator.ContainerConfig[container_name].ImageVersion,
			Binds:           Orchestrator.ContainerConfig[container_name].Binds,
		}
		getAllNotebookContainersPaginated = append(getAllNotebookContainersPaginated, container)
	}

	json.NewEncoder(w).Encode(getAllNotebookContainersPaginated)
}

func StopNotebookContainerJSON(w http.ResponseWriter, container_name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	StopNotebookContainer := StopNotebookContainerResponse{
		Status:          http.StatusOK,
		ContainerName:   container_name,
		ContainerActive: Orchestrator.ContainerOption[container_name].Active,
	}

	json.NewEncoder(w).Encode(StopNotebookContainer)
}

func RestartNotebookContainerJSON(w http.ResponseWriter, container_name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	restartNotebookContainer := StartNotebookContainerResponse{
		Status:          http.StatusOK,
		ContainerName:   container_name,
		ContainerActive: Orchestrator.ContainerOption[container_name].Active,
		ContainerIp:     Orchestrator.ContainerOption[container_name].Ip,
	}

	json.NewEncoder(w).Encode(restartNotebookContainer)
}

func DeleteNotebookContainerJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func CreateNotebookVolumeJSON(w http.ResponseWriter, container_name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	notebookBinds := NotebookVolumeResponse{
		Status:        http.StatusOK,
		ContainerName: container_name,
		Binds:         Orchestrator.ContainerConfig[container_name].Binds,
	}

	json.NewEncoder(w).Encode(notebookBinds)
}

func RemoveNotebookVolumeJSON(w http.ResponseWriter, container_name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	notebookBinds := NotebookVolumeResponse{
		Status:        http.StatusOK,
		ContainerName: container_name,
		Binds:         Orchestrator.ContainerConfig[container_name].Binds,
	}

	json.NewEncoder(w).Encode(notebookBinds)
}

func DeleteNotebookContainerDataJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

}
