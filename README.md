# Orchestrator Gestione Apptainer

## Cosa è questo progetto

Questo progetto utilizza un server che ha come compito quello di fare da intermediario tra un ipotetico client, che vuole utilizzare apptainer, ed un orchestrator che va a controllare le funzionalità di Apptainer stesso. Il server espone delle API di un orchestrator Apptainer su cui il client, trammite richietse GET/POST al server, potrà andare a controllare e gestire la creazione di immagini, l'avvio di istanze su determinate immagine, fare bind di file tra macchina host e containers etc etc.

L'obbiettivo ultimo è quello poi di implementare questo server/orchestrator come un deamon per l'infrastruttura NextPyteche che gestisca le varie funzionalità di Apptainer in base alle richieste degli utenti.

## Struttura Progetto

Il progetto può essere installato usando l'eseguibile fornito che creerà nella cartella `/opt` una nuova directory contenente i file e le strutture necessarie al funzionamento del server e dell'Orchestrator. Il progetto si struttura poi in:

```
/opt/
└── apptainer_container/  # Cartella del progetto
    ├── orchestrator/     # Cartella con file per server ed Orchestrator
    |   ├── FULL_RESET    # Eseguibile per ripristinare l'Orchestrator
    |   ├── save          # File contente informazioni sui vari container
    |   └── server        # Eseguibile del server 
    └── container:img     # Immagini dei container
```

## Funzionalità dell'Orchestrator

Di per se il server ha una struttura abbastanza semplice. Partendo dalla root del server si possono chiamare i vari servizi ad esso annesso con delle richieste GET/POST inserendo poi nel body, in formato JSON, le specifiche dei comandi che vogliamo eseguire.

```go
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
```

- `/start`: Funzione per la creazione di immagini apptainer
  
- `/stop`: Funzione per fermare temporaneamente delle istanze apptainer
  
- `/restart`: Funzione per far riprendere l'esecuizione di istanze precedentemente fermate
  
- `/getcont`: Funzione che ritorna le specifiche di un detereminato container
  
- `/getallcont`: Funzione che ritorna le specifiche di tutti i container creati e registrati nell'orchestrator del server
  
- `/deletecont`: Funzione che rimuove dall'Orchestrator un Container precedentemente creato e registarto
  
- `/deletecontdata`: Funzione che rimuove dall'Orchestrator l'immagine ed i dati relativi ad un Container precedente rimosso dall'Orchestrator
  
- `/createvolume`: Funzione che crea bind dinamici tra directory della macchina host e del container
  
- `/deletevolume`: Funzione che rimuove bind dinamici precedentemente instaurati
  

## Come usare il server

Il server viene fornito come un eseguibile. Aprendolo apre il server sulla porta 3333 del localhost mettendosi in ascolto di eventuali richieste. Se all'avvio si specifica il tag `--debug` verranno mostrate in tempo reale informazioni aggiuntive sulle azioni del server. Il server ritornerà al chiamante delle funzioni delle informazioni sempre in JSON che idicheranno all'utente l'esito delle azioni richieste, specificando in caso di errore se il problema sia dovuto ad una richiesta incompleta o errata oppure ad un problema interno all'Orchestrator.
