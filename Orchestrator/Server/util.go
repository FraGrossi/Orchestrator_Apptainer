package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"bytes"
	"net/http"

	"math/rand/v2"
	"strings"

	"encoding/json"
	"io/ioutil"

	"time"

	"github.com/fatih/color"
)

type Pairs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type VolumeRequest struct {
	ContainerName     string `json:"container_name"`
	HostLocation      string `json:"host_location"`
	ContainerLocation string `json:"container_location"`
}

type ContainerConfig struct {
	Hostname     string            `json:"container_name"`
	Image        string            `json:"container_image"`
	ImageDistr   string            `json:"image_distributor"`
	ImageOS      string            `json:"image_os"`
	ImageVersion string            `json:"image_version"`
	Environment  map[string]string `json:"enviroment"`
	User         string            `json:"user"`
	WorkingDir   string            `json:"working_dir"`
	Labels       map[string]string `json:"labels"`
	Cmd          []string          `json:"container_command"`
	StartBind    string            `json:"volume_name"`
	Binds        map[string]string `json:"binds"`
	GroupAdd     string            `json:"group_add"`
	//network_mode string TODO
}

type ContainerOption struct {
	ContainerName string `json:"container_name"`
	Active        bool   `json:"isActive"`
	Ip            string `json:"ip"`
	//platform -> (arm/arm64)Non presente in Apptainer, aggirabile nella definizione del bootstrap
}

type ApptainerOrchestrator struct {
	filePath        string
	ContainerConfig map[string]ContainerConfig `json:"container_config"`
	ContainerOption map[string]ContainerOption `json:"container_option"`
	ContainerData   map[string]string          `json:"container_data"`
	//TODO possibile errore dove creando un cont ed eliminandolo SOLO dall'orchestrator
	//per poi crearne uno nuovo -> crash per immagini identiche nei dati salvati
}

func debugMessage(message string) {
	if debug_messages {
		print(time.Now().Format("15:04:05") + " ")
		debug_print := color.New(color.BgWhite, color.FgBlack, color.Bold).PrintFunc()
		debug_print("  DEBUG  ")
		fmt.Println(" | " + message)
	}
}

func errorMessage(message string) {
	fmt.Print(time.Now().Format("15:04:05") + " ")
	error_print := color.New(color.BgRed, color.Bold, color.FgWhite).PrintfFunc()
	error_print("  ERROR  ")
	fmt.Println(" | " + message)
}

func warningMessage(message string) {
	print(time.Now().Format("15:04:05") + " ")
	warning_print := color.New(color.BgYellow, color.FgWhite, color.Bold).PrintfFunc()
	warning_print(" WARNING ")
	fmt.Println(" | " + message)
}

func successMessage(message string) {
	print(time.Now().Format("15:04:05") + " ")
	success_print := color.New(color.BgGreen, color.FgWhite, color.Bold).PrintFunc()
	success_print(" SUCCESS ")
	fmt.Println(" | " + message)
}

func infoMessage(message string) {
	print(time.Now().Format("15:04:05") + " ")
	info_print := color.New(color.BgBlue, color.Bold, color.FgWhite).PrintfFunc()
	info_print("  INFO   ")
	fmt.Println(" | " + message)
}

func flagCheck() bool {
	help := flag.Bool("help", false, "Mostra guida all'utilizzo del comando")
	debug := flag.Bool("debug", false, "Abilita messaggi di debug")

	flag.Parse()

	if *help {
		red_print := color.New(color.FgRed).PrintfFunc()
		red_print("Uso:\n")
		fmt.Println("  ./server [--debug]")
		red_print("Opzioni:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *debug {
		debugMessage("| DEBUG MODE ON |")
		fmt.Println()
		return true
	} else {
		return false
	}
}

func deleteContainerData(path string) bool {
	cmd := exec.Command("sudo", "rm", "-rf", path)

	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func instanceStart(config ContainerConfig) ErrorResponse {
	arg_bind := []string{}

	for hostPath, containerPath := range config.Binds {
		arg_bind = append(arg_bind, "--bind "+hostPath+":"+containerPath)
	}

	//setto arg di Bind
	allArgs := append([]string{"instance", "start", "--contain", "--writable", "--net", "--network=bridge"}, arg_bind...)

	//setto arg di Image e Nome_Istanza
	allArgs = append(allArgs, config.Image, config.Hostname)

	cmd := exec.Command("/usr/local/bin/apptainer-wrapper.sh", allArgs...)

	err := cmd.Run()

	if err != nil {
		errorMessage("Impossibile aprire istanza")
		return ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Impossibile aprire istanza",
		}
	}

	successMessage("Istanza " + config.Hostname + " avviata.")

	return ErrorResponse{}
}

func instanceStop(container_name string) ErrorResponse {
	cmd := exec.Command("/usr/local/bin/apptainer-wrapper.sh", "instance", "stop", container_name)

	err := cmd.Run()

	if err != nil {
		errorMessage("Impossibile fermare istanza <" + container_name + ">")
		return ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Impossibile fermare istanza <" + container_name + ">",
		}
	}

	return ErrorResponse{}
}

func checkDuplicate(config ContainerConfig) ContainerConfig {

	config_modified := config

	tmp_hostname := config.Hostname

	if _, exists := Orchestrator.ContainerConfig[tmp_hostname]; exists {
		warningMessage("Istanza <" + config.Hostname + "> già registrata")
		for {
			tmp_hostname = config.Hostname + fmt.Sprint(rand.IntN(100000))
			if _, exists := Orchestrator.ContainerConfig[tmp_hostname]; !exists {
				break
			}
		}

		config_modified.Hostname = tmp_hostname
		config_modified.Labels["notebook-name"] = tmp_hostname

		warningMessage("Istanza rinominata <" + config_modified.Hostname + ">")
	}

	return config_modified
}

func checkBindJson(w http.ResponseWriter, volume_data VolumeRequest, usage string) bool {

	//Controllo la presenza dei dati necessari all'operazione di bind
	if volume_data.ContainerName == "" {
		errorMessage("container_name non definito nelal richiesta")
		errorResponseJSON(w, http.StatusBadRequest, usage+"\ncontainer_name not defined in the request")
		return true
	}
	if volume_data.HostLocation == "" {
		errorMessage("host_location non definito")
		errorResponseJSON(w, http.StatusBadRequest, usage+"\ncontainer_location not defined in the request")
		return true
	}
	if volume_data.ContainerLocation == "" {
		errorMessage("container_location non definito")
		errorResponseJSON(w, http.StatusBadRequest, usage+"\ncontainer_location not defined in the request")
		return true
	}

	//Controllo che le directory indicate eseistano e non siano dei file
	return false
}

func checkPathBind(w http.ResponseWriter, volume_data VolumeRequest, usage string) bool {
	cmd := exec.Command("test", "-d", volume_data.HostLocation)

	err := cmd.Run()

	if err != nil {
		errorMessage("host_location non è una directory valida")
		errorResponseJSON(w, http.StatusBadRequest, usage+"\nhost_location is not a valid source")
		return true
	}

	cmd = exec.Command("/usr/local/bin/apptainer-wrapper.sh", "exec", Orchestrator.ContainerConfig[volume_data.ContainerName].Image, "test", "-d", volume_data.ContainerLocation)

	err = cmd.Run()

	if err != nil {
		errorMessage("container_location non è una directory valida")
		errorResponseJSON(w, http.StatusBadRequest, usage+"\ncontainer_location is not a valid bind point")
		return true
	}

	return false
}

func orchestratorLoad() {

	infoMessage("[Loading Orchestrator Save]")

	// Verifica la presenza del file di salvataggio dell'orchestrator

	_, err := os.Stat(Orchestrator.filePath)
	if err != nil && os.IsNotExist(err) {
		errorMessage("Orchestrator File not found")
		errorMessage("Wipe di dati e istanze passate...")

		debugMessage(filepath.Join(filepath.Dir(Orchestrator.filePath), "FULL_RESET"))
		exec.Command("sudo", filepath.Join(filepath.Dir(Orchestrator.filePath), "FULL_RESET"), Orchestrator.filePath).Run()

		errorMessage("Wipe Completato")
		warningMessage("Operazione di Loading saltata")
		return
	}

	jsonData, err := os.ReadFile(Orchestrator.filePath)
	if err != nil {
		errorMessage("Apertura file orchestrator fallita")
		return
	}

	err = json.Unmarshal(jsonData, &Orchestrator)
	if err != nil {
		errorMessage("Lettura dati JSON dell'orchestrator fallita")
		return
	}

	infoMessage("<Check container files presence>")

	for container_name, config := range Orchestrator.ContainerConfig {

		//Controllo che ogni container nell'orchestrator abbia i propri file presenti

		if !isContainerImagePresent(config.Image) {
			errorMessage("Immagine del container <" + container_name + "> non presente. Ricostruzione...")

			def_file, errResp := definitionBuilder(config)
			if errResp.Status != 0 {
				print("hello")
			}
			imageBuilder(def_file, config.Image)

			warningMessage("Immagine container <" + container_name + "> ricostruita.")

		}
	}

	infoMessage("<Check instances status>")

	for container_name, option := range Orchestrator.ContainerOption {
		if option.Active {

			//Controllo se i Container attivi nell'orchestrator sono effettivamente attivi

			if !isContainerActive(container_name) {
				warningMessage("Container <" + container_name + "> non attivo. Avvio...")
				instanceStart(Orchestrator.ContainerConfig[container_name])
				tmp := Orchestrator.ContainerOption[container_name]
				tmp.Ip = getContainerIP(container_name)
				Orchestrator.ContainerOption[container_name] = tmp
			}
		}
	}

	infoMessage("[Orchestrator Ready]\n\n\n")
}

func orchestratorSave() {
	jsonData, err := json.MarshalIndent(Orchestrator, "", "  ")

	if err != nil {
		errorMessage("Impossibile preparare i dati dell'orchestrator per la scrittura")
		return
	}

	// Scrivere i dati su un file
	err = ioutil.WriteFile(Orchestrator.filePath, jsonData, 0644)
	if err != nil {
		errorMessage("Errore durante la scrittura dei dati orchestrator sul file")
		return
	}
}

func isContainerActive(instance_name string) bool {
	cmd := exec.Command("/usr/local/bin/apptainer-wrapper.sh", "exec", "instance://"+instance_name, "ls")

	err := cmd.Run()

	if err != nil {
		return false
	}

	return true
}

func isContainerImagePresent(image string) bool {
	cmd := exec.Command("/usr/local/bin/apptainer-wrapper.sh", "inspect", image)

	err := cmd.Run()

	if err != nil {
		return false
	}

	return true
}

func exit4error() {
	orchestratorSave()
	os.Exit(1)
}

func newContainerConfig(r *http.Request) (ContainerConfig, ErrorResponse) {
	var container_config = ContainerConfig{
		Environment: map[string]string{
			"NB_USER":    "nextpyter",
			"NB_GID":     "33",
			"NB_UMASK":   "002",
			"CHOWN_HOME": "yes",
		},
		User:       "root",
		WorkingDir: "/home/nextpyter",
		GroupAdd:   "user",
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		errorMessage("Errore nella lettura del corpo della richiesta per la creazione della configurazione")
		return ContainerConfig{}, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Errore nella lettura del corpo della richiesta per la creazione della configurazione",
		}
	}

	defer r.Body.Close()

	err = json.Unmarshal(body, &container_config)

	if err != nil {
		errorMessage("Lettura JSON del container_config fallita")
		return ContainerConfig{}, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Lettura JSON del container_config fallita",
		}
	}

	container_config.Labels = map[string]string{
		"taint":         "nextpyter-notebook",
		"notebook-name": container_config.Hostname,
	}

	//TODO trovare un modo di eliminare StartBind
	if container_config.StartBind != "" {
		container_config.Binds = map[string]string{
			container_config.StartBind: "/home/nextpyter",
		}
	} else {
		container_config.Binds = make(map[string]string)
	}

	return container_config, ErrorResponse{}
}

func newContainerOption(containerConfig ContainerConfig) ContainerOption {
	return ContainerOption{
		ContainerName: containerConfig.Hostname,
		Active:        true,
	}
}

func definitionBuilder(config ContainerConfig) (string, ErrorResponse) {
	var definition_file = "definition.def"
	var definition = fmt.Sprintf(`
Bootstrap: %s 
From: %s:%s 

%%post
	apt-get update && apt-get install -y curl
	apt install iputils-ping -y 
	mkdir %s

%%environment
export HOME=%s
cd %s
	`, config.ImageDistr, config.ImageOS, config.ImageVersion, config.WorkingDir, config.WorkingDir, config.WorkingDir)

	for varName, varValue := range config.Environment {
		definition += "	export " + varName + "=" + varValue + "\n"
	}

	definition += "\n%%startscript\n"

	for _, item := range config.Cmd {
		definition += "	" + item + "\n"
	}

	//TODO da aggiungere altre impostazioni

	//Osservo il definition File
	debugMessage(definition)

	file, err := os.Create(definition_file)

	if err != nil {
		errorMessage("Creazione del file di definizione non riuscita")
		return "", ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Creazione del file di definizione non riuscita",
		}
	}

	defer file.Close()

	_, err = file.WriteString(definition)

	if err != nil {
		errorMessage("Inserimento dati nel file .def non riuscito")
		return "", ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Inserimento dati nel file .def non riuscito",
		}
	}

	return definition_file, ErrorResponse{}
}

func imageBuilder(definition_file string, image string) ErrorResponse {

	debugMessage("/usr/local/bin/apptainer-wrapper.sh build --sandbox " + image + " " + definition_file)

	cmd := exec.Command("/usr/local/bin/apptainer-wrapper.sh", "build", "--sandbox", image, definition_file)

	err := cmd.Run()

	if err != nil {
		errorMessage("Creazione dell'immagine sandbox non riuscita")
		return ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Creazione dell'immagine sandbox non riuscita",
		}
	}

	successMessage("Immagine Sandbox creata.")

	return ErrorResponse{}
}

func imageNameCombiner(path string, hostname string, image string) (string, ErrorResponse) {

	if strings.ContainsAny(image, ";&|`\\/") {
		errorMessage("Prevenzione esecuzione non sicura sudo: nome immagine " + image + " non valido.\n")
		return "", ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Prevenzione esecuzione non sicura sudo: nome immagine " + image + " non valido",
		}
	}

	if strings.ContainsAny(hostname, ";&|`") {
		errorMessage("Prevenzione esecuzione non sicura sudo: hostname " + hostname + " non valido.\n")
		return "", ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Prevenzione esecuzione non sicura sudo: hostname " + hostname + " non valido",
		}
	}

	return path + hostname + ":" + image, ErrorResponse{}
}

func getContainerIP(container_name string) string {
	cmd := exec.Command("/usr/local/bin/apptainer-wrapper.sh", "instance", "list")

	var out bytes.Buffer
	cmd.Stdout = &out

	//TODO in caso di errore mettere una aggiornamento degli IP sull'orchestrator quando si fa un getallcont
	cmd.Run()

	ip := "0.0.0.0"

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, container_name) {
			fields := strings.Fields(line)
			ip = fields[2]
		}
	}

	return ip
}

func containerNameJSON(r *http.Request) (string, ErrorResponse) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		errorMessage("Errore nella lettura del corpo della richiesta")
		return "", ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Errore nella lettura del corpo della richiesta",
		}
	}

	defer r.Body.Close()

	var jsonData map[string]interface{}

	err = json.Unmarshal(body, &jsonData)

	if err != nil {
		errorMessage("Errore nel parsing del JSON nella ricerca del 'container_name'")
		return "", ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Errore nel parsing del JSON nella ricerca del 'container_name'",
		}
	}

	container_name, ok := jsonData["container_name"].(string)

	if !ok {
		errorMessage("Tag JSON 'container_name' non trovato o non è una stringa > " + fmt.Sprint(http.StatusBadRequest))
		return "", ErrorResponse{
			Status:  http.StatusBadRequest,
			Message: "Tag JSON 'container_name' non trovato o non è una stringa",
		}
	}

	return container_name, ErrorResponse{}
}

func containerVolumeJSON(r *http.Request) (VolumeRequest, ErrorResponse) {
	volume_data := VolumeRequest{}

	body, err := ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		errorMessage("Errore nella lettura del corpo della richiesta")
		return volume_data, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Internal error while reading JSON request ",
		}
	}

	err = json.Unmarshal(body, &volume_data)

	if err != nil {
		errorMessage("Errore nel parsing del JSON per la lettura dei dati di Bind")
		return volume_data, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: "Internal error while parsing JSON request ",
		}
	}

	return volume_data, ErrorResponse{}
}
