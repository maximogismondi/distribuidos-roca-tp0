# TP0: Docker + Comunicaciones + Concurrencia

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

## Instrucciones de uso

El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

### Servidor

Se trata de un "echo server", en donde los mensajes recibidos por el cliente se responden inmediatamente y sin alterar.

Se ejecutan en bucle las siguientes etapas:

1. Servidor acepta una nueva conexión.
2. Servidor recibe mensaje del cliente y procede a responder el mismo.
3. Servidor desconecta al cliente.
4. Servidor retorna al paso 1.

### Cliente

 se conecta reiteradas veces al servidor y envía mensajes de la siguiente forma:

1. Cliente se conecta al servidor.
2. Cliente genera mensaje incremental.
3. Cliente envía mensaje al servidor y espera mensaje de respuesta.
4. Servidor responde al mensaje.
5. Servidor desconecta al cliente.
6. Cliente verifica si aún debe enviar un mensaje y si es así, vuelve al paso 2.

### Ejemplo

Al ejecutar el comando `make docker-compose-up`  y luego  `make docker-compose-logs`, se observan los siguientes logs:

```log
client1  | 2024-08-21 22:11:15 INFO     action: config | result: success | client_id: 1 | server_address: server:12345 | loop_amount: 5 | loop_period: 5s | log_level: DEBUG
client1  | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:14 DEBUG    action: config | result: success | port: 12345 | listen_backlog: 5 | logging_level: DEBUG
server   | 2024-08-21 22:11:14 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°3
client1  | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°3
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°5
client1  | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°5
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:40 INFO     action: loop_finished | result: success | client_id: 1
client1 exited with code 0
```

## Parte 1: Introducción a Docker

En esta primera parte del trabajo práctico se plantean una serie de ejercicios que sirven para introducir las herramientas básicas de Docker que se utilizarán a lo largo de la materia. El entendimiento de las mismas será crucial para el desarrollo de los próximos TPs.

### Ejercicio N°1

Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc.

El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

`./generar-compose.sh docker-compose-dev.yaml 5`

Considerar que en el contenido del script pueden invocar un subscript de Go o Python:

```bash
#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
python3 mi-generador.py $1 $2
```

En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

### Ejercicio N°2

Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).

### Ejercicio N°3

Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `

### Ejercicio N°4

Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

## Parte 2: Repaso de Comunicaciones

Las secciones de repaso del trabajo práctico plantean un caso de uso denominado **Lotería Nacional**. Para la resolución de las mismas deberá utilizarse como base el código fuente provisto en la primera parte, con las modificaciones agregadas en el ejercicio 4.

### Ejercicio N°5

Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente - Agencia de Quiniela

Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Servidor - Lotería Nacional

Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación

Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:

* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).

### Ejercicio N°6

Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_).
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del _batch_ fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB.

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

### Ejercicio N°7

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

## Parte 3: Repaso de Concurrencia

En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

### Ejercicio N°8

Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

## Condiciones de Entrega

Se espera que los alumnos realicen un _fork_ del presente repositorio para el desarrollo de los ejercicios y que aprovechen el esqueleto provisto tanto (o tan poco) como consideren necesario.

Cada ejercicio deberá resolverse en una rama independiente con nombres siguiendo el formato `ej${Nro de ejercicio}`. Se permite agregar commits en cualquier órden, así como crear una rama a partir de otra, pero al momento de la entrega deberán existir 8 ramas llamadas: ej1, ej2, ..., ej7, ej8.
 (hint: verificar listado de ramas y últimos commits con `git ls-remote`)

Se espera que se redacte una sección del README en donde se indique cómo ejecutar cada ejercicio y se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado (Parte 2) y los mecanismos de sincronización utilizados (Parte 3).

Se proveen [pruebas automáticas](https://github.com/7574-sistemas-distribuidos/tp0-tests) de caja negra. Se exige que la resolución de los ejercicios pase tales pruebas, o en su defecto que las discrepancias sean justificadas y discutidas con los docentes antes del día de la entrega. El incumplimiento de las pruebas es condición de desaprobación, pero su cumplimiento no es suficiente para la aprobación. Respetar las entradas de log planteadas en los ejercicios, pues son las que se chequean en cada uno de los tests.

La corrección personal tendrá en cuenta la calidad del código entregado y casos de error posibles, se manifiesten o no durante la ejecución del trabajo práctico. Se pide a los alumnos leer atentamente y **tener en cuenta** los criterios de corrección informados  [en el campus](https://campusgrado.fi.uba.ar/mod/page/view.php?id=73393).

## Entrega

### Datos

Alumno:

* **Nombre completo:** Máximo Gismondi
* **Padrón:** 110119
* **Usuario GitHub:** maximogismondi
* **Correo:** <magismondi@fi.uba.ar>

Materia:

* **Facultad:** Factultad de Ingeniería de la Universidad de Buenos Aires
* **Materia:** Sistemas Distribuidos I
* **Catetra:** Roca
* **Cuatrimestre:** 1er Cuatrimestre 2025

Entrega:

* **Fecha de entrega:** 27/03/2025

### Aclaraciones inciales

La solución fue desarollada iterativamente para cada uno de los ejercicios planteados. Las distintas etapas se pueden ver en las distintas ramas `ej1`, `ej2`, ..., `ej8` del repositorio. Aún así, la solución en la última rama (y en la rama `main`) es aquella que más trabajada, probada y prolija se encuentra. Es por eso que si bien los tests automáticos pasan en todas las ramas, la solución final es la que se encuentra en la rama `main` y es aquella que presentaría como solución final.

### Ejecución de los ejercicios

#### Ejercicio 1

Se creó un script `generar-compose.sh` que permite generar un archivo de Docker Compose con la cantidad de clientes especificada. Para ello se utilizó un script en Python que genera el archivo de Docker Compose de forma dinámica. Está desarrollado de tal forma que se puedan agregar más configuraciones en el futuro de manera sencilla.

Para ejecutar el script `generar-compose.sh` se debe correr el siguiente comando:

```bash
./generar-compose.sh <docker-compose-file> <cantidad-de-clientes>
```

Luego de ejecutar el comando, se generará un archivo `docker-compose-file` con la definición de Docker Compose con la cantidad de clientes especificada.

Este archivo de docker compose se puede utilizar para inicializar el ambiente de desarrollo con la cantidad de clientes especificada con el comando:

```bash
docker-compose -f <docker-compose-file> up
```

#### Ejercicio 2

Para este segundo ejercicio, se agregó un archivo `.dockerignore` para evitar que los archivos de configuración sean copiados a la imagen de Docker. Además, se refinaron los archivos `Dockerfile` para que solamente se copien los archivos pertinentes para la construcción de la imagen.

Luego, para habilitar la configuración por fuera de la imagen, se utilizaron bind mounts a los contenedores de docker, referenciando localmente los archivos de configuración. Hacer esto con cada uno de los contenedores es bastante engorroso, por lo que se modificó el script `generar-compose.sh` para que automáticamente se hagan los bind mounts automáticamente en el docker-compose generado.

Todos los clientes comparten el archivo `config.yaml` y el servidor utiliza el archivo `config.ini`.

Esto permite que al construir las imágenes aprovechando la cache Docker y al mismo tiempo poder modificar los archivos de configuración sin necesidad de reconstruir las imágenes.

#### Ejercicio 3

En este caso, se creó un script `validar-echo-server.sh` que permite verificar el correcto funcionamiento del echo-server utilizando el comando `netcat` para interactuar con el mismo.

Para ejecutar el script `validar-echo-server.sh` se debe correr el siguiente comando:

```bash
./validar-echo-server.sh
```

Este envía un mensaje `Hello, World!` al servidor y espera recibir el mismo mensaje de vuelta. Si la validación es exitosa, se imprime `action: test_echo_server | result: success`, de lo contrario se imprime `action: test_echo_server | result: fail`.

#### Ejercicio 4

Para este ejercicio, se modificaron tanto el cliente como el servidor para que terminen de forma _graceful_ al recibir la señal SIGTERM. Esto implica que todos los _file descriptors_, _sockets_, _threads_ y _procesos_ deben cerrarse correctamente antes de que el thread principal de la aplicación muera.

En el servidor escrito en Python, se utilizó la librería `signal` para capturar la señal SIGTERM y se implementó un handler que cierra el socket del servidor. Además se le agrego un timeout al socket para que no quede esperando indefinidamente y que pueda reaccionar a la señal SIGTERM y cerrar y liberar los recursos.

En los clientes escritos en Go, se utilizó la librería `os/signal` para capturar la señal SIGTERM y se implementó un sistema de canales y goroutines para que el cliente pueda reaccionar a la señal SIGTERM y cerrar y liberar los recursos.

#### Ejercicio 5

Ya metiendonos en el caso de la Lotería Nacional, se modificaron tanto el cliente como el servidor para que puedan manejar las apuestas de las agencias de quiniela.

Primero definimos un protocolo el cual irá mutando a medida que se avance en los ejercicios. En este caso, el protocolo es bastante simple, se envía un mensaje con los campos de la apuesta separados por un separador (en este caso `+`).

Además se crearon wrappers para los sockets para evitar los fenómenos de _short read_ y _short write_ ya que se aseguran de leer hasta el delimitador de la comunicación establecido (en este caso `\n`).

Hasta ahora el protocolo funciona como un stop-and-wait, es decir, el cliente envía un mensaje y espera la respuesta del servidor antes de enviar el siguiente mensaje. Esto facilita mucho el uso de sockets y la implementación del protocolo.

El mensaje enviado por el cliente es de la forma:

```txt
AGENCY<id_cliente>++<nombre>+<apellido>+<dni>+<nacimiento>+<numero>\n
```

Por otro lado, el servidor recibe el mensaje, lo procesa y lo almacena en una estructura de datos. Luego, responde al cliente con un mensaje de confirmación según el resultado del almacenamiento.

OK:

```txt
success\n
```

Fail:

```txt
failure\n
```

Los datos enviados por la agencia de quiniela son configurados por variables de entorno en el cliente. Es por ello que se modificó el script `generar-compose.sh` para que configure de forma automática las variables de entorno de los clientes.

#### Ejercicio 6

Para este ejercicio, se modificaron los clientes para que envíen más de un apuesta a la vez, se envían en batches de tamaño configurable.

Para nutrir el cliente con las apuestas, se leen archivos `.csv` que contienen las apuestas de las agencias de quiniela. Estos archivos son inyectados en los contenedores de Docker de la misma forma que los archivos de configuración, utilizando bind mounts.

El servidor recibe los mensajes de los clientes y los procesa en batches. Si todos los mensajes del batch fueron procesados correctamente, el servidor responde con un mensaje de confirmación, de lo contrario responde con un mensaje de error.

Los mensajes de batch enviados por el cliente son de la forma:

```txt
AGENCY+<id_cliente>+<nombre>+<apellido>+<dni>+<nacimiento>+<numero>*...*AGENCY+<id_cliente>+<nombre>+<apellido>+<dni>+<nacimiento>+<numero>\n
```

Donde `*` es el separador de apuestas y `\n` es el delimitador de comunicación.

Si todo está OK, el servidor escribirá en el log:

```txt
action: apuesta_recibida | result: success | cantidad: <cantidad_de_apuestas>
```

En caso de error, el servidor escribirá en el log:

```txt
action: apuesta_recibida | result: fail | cantidad: <cantidad_de_apuestas>
```

El servidor responderá al cliente con un mensaje de confirmación o error según el resultado del procesamiento del batch.

```txt
success\n
```

```txt
failure\n
```

Una vez el cliente envía todos los batches, notifica al servidor que terminó de enviar las apuestas.

```txt
finish\n
```

#### Ejercicio 7

En este caso, una vez completado el envío de todas las apuestas, el cliente debe esperar a que todos los clientes terminen de enviar las apuestas y luego consultar la lista de ganadores del sorteo correspondientes a su agencia.

Para ello, se le agregó al protocolo de comunicación un mensaje inicial que permite al servidor identificar a la agencia de quiniela que está enviando el mensaje.

```txt
AGENCY <id_cliente>\n
```

Y luego, se determinaran 3 tipos de mensajes de la agencia de quiniela:

* `new_bet`: mensaje de batch de apuestas.
* `finish`: mensaje de finalización de envío de apuestas.
* `request`: mensaje de solicitud de ganadores.

En el caso del servidor existen 4 posible mensajes:

* `success`: mensaje de confirmación de almacenamiento de apuestas.
* `failure`: mensaje de error en el almacenamiento de apuestas.
* `winners`: mensaje de respuesta a la solicitud de ganadores.
* `not_ready`: mensaje de que aún no se puede responder a la solicitud de ganadores.

Una vez enviadas todas las apuestas, el cliente consultará por la lista de ganadores del sorteo correspondientes a su agencia. Si no está listo para responder, el servidor responderá con un mensaje `not_ready` y el cliente se desconectará y volverá a intentar más tarde.

Si el servidor está listo para responder, responderá con un mensaje `winners` y la lista de ganadores de la agencia de quiniela.

```txt
winners,<dni_ganador_1>,<dni_ganador_2>,...,<dni_ganador_n>\n
```

En este caso se usa el delimitador `,` para separar los DNIs de los ganadores y `\n` para delimitar la comunicación.

El cliente imprimirá por log:

```txt
action: consulta_ganadores | result: success | cant_ganadores: <cantidad_de_ganadores>
```

La política de reintentos del clientes se basa en un _backoff_ exponencial, es decir, si el servidor responde con `not_ready`, el cliente esperará un tiempo exponencialmente creciente antes de volver a intentar la consulta. Particularmente se espera `2^i` segundos, donde `i` es el número de intento.

Del lado del servidor, una vez que todas las agencias de quiniela han enviado sus apuestas, el servidor realizará el sorteo y dejará un log de que el sorteo fue exitoso.

```txt
action: sorteo | result: success
```

A partir de ese momente le respoderá a cada cliente con la lista de ganadores de su agencia y no con un `not_ready`.

Para que el servidor sepa exacatmente cuantas agencias de quiniela hay (para esperarlas a todas) se configuró una variable de entorno en el script `generar-compose.sh` que indica la cantidad de agencias de quiniela en función de la cantidad de clientes.

#### Ejercicio 8

Por último se deberá modificar el servidor para que permita aceptar conexiones y procesar mensajes de forma concurrente. Para ello se utilizó la librería `threading` de Python para crear un thread por cada conexión entrante.

Cabe acalarar que Python tiene un _Global Interpreter Lock_ (GIL) que impide que dos threads ejecuten código de Python al mismo tiempo. Aún así, se puede utilizar threads para realizar operaciones de I/O y así permitir que el servidor pueda aceptar conexiones y procesar mensajes de forma concurrente sin perder tiempo de CPU en operaciones de I/O.

En este caso, se creará un thread por cada conexión entrante por lo que el servidor podrá aceptar conexiones y procesar mensajes de forma concurrente.

En este caso contamos con un recurso compartido, que es la estructura de datos que almacena las apuestas. Para evitar problemas de concurrencia, se utilizó un `Lock` para proteger el acceso a la estructura de datos compartida. Además, se protegieron ciertas estructuras sensibles del servidor como la cantidad de clientes atendidos y los resultados del sorteo.

Como pasamos de tener 1 a N sockets abiertos de forma simultanea, se guardó una referencia a cada socket + thread en una lista para poder cerrar los sockets y esperar a que los threads terminen de forma _graceful_ al recibir la señal SIGTERM.

### Resultado de los ejercicios

Como el ejercicio se fue realizando de forma incremental, algunas características del sistema fueron mutando a lo largo de los ejercicios, se realizó una serie de refactorizaciones para mejorar la calidad del código, la modularidad y para estandarizar el código y el protocolo.

#### Protocolo de comunicación

El protocolo final de comunicación cuenta de los siguientes mensajes:

##### Cliente -> Servidor

* `agency:<id_cliente>\n`: mensaje de identificación de la agencia de quiniela.
* `bet_batch:<apuesta_1>+<apuesta_2>+...+<apuesta_n>\n`: mensaje de batch de apuestas.
  * `<apuesta_i>`: `<nombre>+<apellido>+<dni>+<nacimiento>+<numero>`
* `finish\n`: mensaje de finalización de envío de apuestas.
* `request_results\n`: mensaje de solicitud de ganadores.

##### Servidor -> Cliente

* `success\n`: mensaje de confirmación de almacenamiento de apuestas.
* `failure\n`: mensaje de error en el almacenamiento de apuestas.
* `winners:<dni_ganador_1>,<dni_ganador_2>,...,<dni_ganador_n>\n`: mensaje de respuesta a la solicitud de ganadores.
* `not_ready\n`: mensaje de que aún no se puede responder a la solicitud de ganadores.

##### Carácteres especiales

* `:`: separador de la cabecera del mensaje.
* `*`: separador de apuestas en el batch.
* `+`: separador de campos de la apuesta.
* `,`: separador de DNIs de los ganadores.
* `\n`: delimitador de comunicación.

#### Mecanismos de sincronización

Para proteger el acceso a la estructura de datos compartida se utilizó un `Lock` de Python. Este `Lock` se adquiere antes de acceder a la estructura de datos de las apuestas y se libera una vez que se termina de acceder a la estructura de datos. Además se utilizó un `Lock` para proteger el acceso a la cantidad de clientes atendidos y a los resultados del sorteo.

#### Graceful shutdown

Para el _graceful shutdown_ se utilizó un sistema de canales y goroutines en Go y un handler de señales en Python. En ambos casos se utilizó un sistema de canales para comunicar la señal SIGTERM al thread principal y que este pueda cerrar y liberar los recursos de forma _graceful_.

En el caso del servidor en Python, se guardó una referencia a par de socket + thread en una lista para poder cerrar los sockets y esperar a que los threads terminen de forma _graceful_ al recibir la señal SIGTERM.

Además, por cada nueva conexión (o timeout) se verificará el estado de los threads del servidor y se cerrarán aquellos que hayan terminado para liberar los recursos lo antes posible.

#### Logs

Se dejaron logs informativos en el servidor y en los clientes para poder seguir el flujo de la aplicación y poder identificar posibles problemas. Los logs siguen el formato:

```txt
action: <nombre_de_la_accion> | result: <resultado_de_la_accion> | <otras_variables>
```

Se trata de dejar en el nivel de log INFO información relevante para el seguimiento de la aplicación y en el nivel de log DEBUG información más detallada para el seguimiento de la aplicación.

#### Lecturas y Escrituras

Para simplificar las lecturas y escrituras de los sockets se crearon wrappers que se encargan de leer y escribir hasta el delimitador de comunicación establecido. Esto permite que el servidor y el cliente puedan leer y escribir mensajes de forma sencilla y sin preocuparse por los fenómenos de _short read_ y _short write_ ya que estos se encuentran handleados con reintentos en los casos de escritura y buffers con el contenido "sobranten" en los casos de lectura.

Tanto las escritura como lecturas se realizan con un buffer de tamaño fijo de 1024 bytes y cuentan con un timeout para evitar que la aplicación quede esperando indefinidamente.

#### Reintentos

En el caso de que el servidor no esté listo para responder a la solicitud de ganadores, el cliente esperará un tiempo exponencialmente creciente antes de volver a intentar la consulta. Particularmente se espera `2^i` segundos, donde `i` es el número de intento. Esto se conoce como _backoff_ exponencial y permite evitar busy-waiting y congestionar el servidor con consultas innecesarias.

Además se agregó un límite de reintentos para evitar que el cliente esté reintentando la consulta de forma indefinida. Este limite es de 64 segundos, es decir, el cliente intentará la consulta 6 veces antes de desistir.

#### Comandos de ejecución

Para ejecutar los ejercicios se deben seguir los siguientes pasos:

1. Generar el archivo de Docker Compose con la cantidad de clientes deseada:

    ```bash
    ./generar-compose.sh <docker-compose-file> <cantidad-de-clientes>
    ```

    Cabe aclarar que el script `generar-compose.sh` se encarga de configurar las variables de entorno de los clientes y de inyectar los archivos de configuración y los archivos de apuestas en los contenedores de Docker. Los archivos de apuestas se encuentran en la carpeta `.data` del proyecto y deben tener el formato `agency-{N}.csv`.

2. Inicializar el ambiente de desarrollo con el archivo de Docker Compose generado:

    ```bash
    docker-compose -f <docker-compose-file> up [-d]
    ```

3. Si se ejecutan de forma `detach` se pueden ver los logs de los contenedores con el comando:

    ```bash
    docker-compose -f <docker-compose-file> logs -f
    ```

    Para finalizar la ejecución de los contenedores se puede utilizar el comando:

    ```bash
    docker-compose -f <docker-compose-file> down
    ```

Los archivos de configuración con los que se puede interactuar son `config.yaml` para los clientes y `config.ini` para el servidor.
