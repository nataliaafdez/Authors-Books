package main

import (
	"authorsbooks/authors/internal/controller"
	"authorsbooks/authors/internal/handler"
	"authorsbooks/authors/internal/pkg/clients"
	"authorsbooks/authors/internal/repository/memory"
	"bytes"
	"encoding/json" //Convertir entre estrcuturas de go y json
	"fmt"           //los mensajes en consola
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Esta función es para por si quiero configurar el servicio sin tener que cambiar tTODO el código, para poder poner en el docker en lugar de 8080 no se, 9090 o algo así
func getenv(k, nat string) string { // lee una variable de entorno
	if v := os.Getenv(k); v != "" {
		return v
	}
	return nat //si la variable de entorno no existe entonces devuelve un valor por defecto
}

func main() { //crea los repositorios, clientes y controller, configura las rutas http, registra el servicio en el consul y empieza el servicodr http
	addr := getenv("ADDR", ":8080")                               //donde va a estar el servicio 8080
	booksURL := getenv("BOOKS_URL", "http://books:8081")          //aqui es donde va a estar el otro microservicio que es books
	metadataURL := getenv("METADATA_URL", "http://metadata:8082") //aquí es donde va a estar el microservicio de metadata

	repo := memory.NewAuthorRepo()                       //Creo un repo de autores en memoria
	booksClient := clients.NewBooksClient(booksURL)      //cliente http para que pueda hablar con el service de libros
	metaClient := clients.NewMetadataClient(metadataURL) //lo mismo pero con metadata

	ctrl := controller.NewAuthorController(repo, metaClient, booksClient) //como lo de controller o sea que hace el repo + clientes
	hd := handler.New(ctrl)                                               //crea el hanlder que va a usar el controller

	mux := http.NewServeMux() //es como un router, tipo como el mapa de Go, o sea recibe las peticiones http y decide que handler funcion o controller debe ejecutarse según el path de la peticion
	hd.Routes(mux)            //registra las rutas del servicio de autores en el mux

	//para ver si el servicio está vivo, preguntar si esto está bien hecho ?!
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})

	//

	//CONSUL
	consulAddr := getenv("CONSUL_ADDR", "localhost:8500") // Es donde va a estar el consul
	serviceName := getenv("SERVICE_HOST", "authors")
	serviceID := serviceName + "-" + strings.TrimPrefix(addr, ":") //ID único del servicio

	_, p, err := net.SplitHostPort(addr) //separa al host y al puerto, como el ejemplo que pregunté
	if err != nil {
		fmt.Println("[authors] invalid ADDR:", err)
		os.Exit(1)
	}
	port, _ := strconv.Atoi(p)                                              //para convertir el puerto a número
	healthURL := fmt.Sprintf("http://host.docker.internal:%d/health", port) //es para que el consul pueda consultar lo de que si está vivo

	go func() { //rutina de go: corre en paralelo con el resto del programa, para que el proceso de registrar el servicio en consul no bloquee al servidor web principal
		for { //infiniro, intenta intenta hasta que consul esté listo
			if err := consulRegister(consulAddr, serviceName, serviceID, "host.docker.internal", port, healthURL); err != nil {
				fmt.Println("[authors] consul register error:", err)
				time.Sleep(2 * time.Second) //esperar dos segundos si falla y reintentar
				continue
			}
			fmt.Println("[authors] registered in consul:", serviceID) //si se registra bien entonces lo avisa y sale del bucle
			break
		}
	}()

	// desregistrar al salir
	go func() { //otra rutina para escuchar las señales del SO y desregistrarse ordendado
		ch := make(chan os.Signal, 1)                    // aqui se reciven las señales del sistema op
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM) //es como decirle a go que cuando llegue una interrupcoón mandarla al canal ch
		<-ch                                             // se queda esperando hasta que llegue una de esas señales o se pausa hasta que alguien quiera cerrar el programa
		_ = consulDeregister(consulAddr, serviceID)      //antes de morir se desregistra del consul para que no piense que el servicio esta vivo
		//acuerdate que _= es que ignora si el error falla
		os.Exit(0) //limpia cuando termicna
	}()

	fmt.Println("[authors] listening on", addr, "books:", booksURL, "metadata:", metadataURL) //log inicial con las direcciones
	if err := http.ListenAndServe(addr, mux); err != nil {                                    //inicia el servidor nttp en addr usando mux
		fmt.Println("[authors] http error:", err) //error por si el servidor está DEAD o si se cae
	}
}

// registar el servicio en consul para que otros microservicios lo encuentren automático tipo si books necesita author entonces lo puede decibrir preguntando en consul
func consulRegister(consulAddr, name, id, address string, port int, healthURL string) error { //es para registrar el servicio en el COnsul
	body := map[string]any{ //esto arma el JSON que el consul espera
		"Name":    name, //nombre del servicio
		"ID":      id,   //id del servicio
		"Address": address,
		"Port":    port,
		"Tags":    []string{"authorsbooks"}, //esto es opcional, maybe despues QUITAR MAYBE, pero es para agruparlo
		"Check": map[string]any{
			"HTTP":     healthURL, // Consul hará GET a esta URL
			"Interval": "10s",     // cada 10s que es cada cuanto va a hacer el chequeo
			"Timeout":  "2s",
			// Si la salud está crítica mucho tiempo, desregistra:
			"DeregisterCriticalServiceAfter": "1m", //si esta 1 min en critico entonces el consul lo quita
		},
	}
	b, _ := json.Marshal(body)
	//convierte el mapa que se hizo en go a JSON
	req, _ := http.NewRequest(http.MethodPut, "http://"+consulAddr+"/v1/agent/service/register", bytes.NewReader(b)) //hace un request put al Consul
	req.Header.Set("Content-Type", "application/json")                                                               //el cuerpo es JSON
	resp, err := http.DefaultClient.Do(req)                                                                          //envía la petición con el cliente HTTP por defecto
	if err != nil {
		return err
	}
	defer resp.Body.Close() //se asegura de cerrar el body de la respuesta, si el status no es ecito entonces devuelve un error con el status
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	return nil
}

func consulDeregister(consulAddr, id string) error { //si el contener muere entonces consul ya no lo mostrara como disponible
	//esto envia una peticion a consul para borrar el servicio con ese ID
	req, _ := http.NewRequest(http.MethodPut, "http://"+consulAddr+"/v1/agent/service/deregister/"+id, nil) //request put a la ruta de desregistro
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
