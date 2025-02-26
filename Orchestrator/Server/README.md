# Server ed Orchestrator

Questa cartella contiene i file source dell'Orchestrator e del server.

- **main.go**
  
  Contiene lo start-up del server e le chiamate alle funzioni dell'orchestrator relative ad una richiesta di un utente.

- **orchestrator_actions.go**
  
  Contiene le funzioni dell'orchestrator che gestiscono l'operato di apptainer sulla macchina. La logica e struttura dietro queste funzioni sono state riprese dagli orchestrator di Docker per l'applicativo NextPyter.

- **util.go**
  
  Contiene funzioni di supporto al lavoro di `orchestrator_actions.go`. Implementato per una miglior leggibilità del codice dell'orchestrator e per introdurre modularità all'interno del progetto.

- **response.go**
  
  Ogni funzione di orchestrator, sia in caso di successo che di fallimento, restituirà una risposta JSON all'utente che la ha invocata. La struttura dietro ogni risposta è definita in questo file.

## Funzioni Orchestrator

- `/start`
  
  Usando una richiesta **POST** ed inviando relative informazioni necessarie, crea una immagine di un container e ne avvia una istanza. Questo container viene poi registrato nell'orchestrator per essere reperibile e modificabile nel futuro.
  
  Tag JSON utilizzabili nella richiesta:
  
  - `container_name`: Specifica il nome da dare all'istanza del container 
  
  - `container_image`: Specifica il nome da dare all'immagine del container
  
  - `image_distributor`: Indica chi sarà il distributore dell'immagine (DockerHub, Singularity Hub, etc...)
  
  - `image_os`, `image_version`: Indica il sistema operativo e la versione da usare nella creazione dell'immagine del container
  
  - `container_command`: Facoltativo, indica quali comandi eseguire all'avvio del container
  
  - `volume_name`: Facoltativo, indica quale cartella della macchina host andrà bindata con la cartella `/home/nextpyter` nel container

- `/stop`
  
  Usando una richiesta **POST** e specificando nel body il tag JSON `container_name` si va a fermare l'istanza attiva di un container. Se il container non ha istanze attive al momento il comando ritornerà un errore.

- `/restart`
  
  Usando una richiesta **POST** e specificando nel body il tag JSON `container_name` si va ad attivare l'istanza a riposo di un container. Se il container ha già una istanza attiva al momento dell'esecuzione del comando, verrà ritornato un errore.

- `/getcont`
  
  Usando una richiesta **GET** e specificando nel body il tag JSON `container_name` si ottiene come risposta JSON tutte le informazione relative al container indicato

- `/getallcont`
  
  Usando una richiesta **GET**, si ottiene come risposta JSON tutte le informazioni relative ai container registrati nell'orchestrator e alle loro rispettive immagini

- `/deletecont`
  
  Usando una richiesta **POST** e specificando nel body il tag JSON `container_name` si va a rimuovere un container dall'orchestrator. Se il container ha una istanza attiva al momento dell'esecuzione del comando allora non saranno effettuate modifiche all'orchestrator e verrà ritornato un errore.

- `/deletecontdata`
  
  Usando una richiesta **POST** e specificando nel body il tag JSON `container_name` si va a rimuovere i dati relativi ad un container precedentemente rimosso dall'orchestrator. Se il container è ancora registrato nell'orchestrator o ha una istanza attiva al momento dell'esecuzione del comando allora sarà ritornato un errore.

- `/createvolume`
  
  Usando una richiesta **POST** e specificando nel body il tag JSON `container_name` insieme a 2 directory, una sulla macchina host `host_location`e una nel container indicato `container_location`, si andrà a creare un bind tra le 2 cartelle. Se le directory specificate non esisteranno o saranno già presenti bind con la stessa cartella della macchina host allora il comando ritornerà errore.

- `/deletevolume`
  
  Usando una richiesta **POST** e specificando nel body i tag JSON `container_name`, `host_location` e `container_location`, si andrà a rimuovere un bind precedentemente creato tra la directory host e quella del container. Se il bind non è stato trovato o le directory indicate non esistono allora verrà ritornato un errore.

