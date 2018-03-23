#In-memory concurrent key-value storagе (in-memory конкурентное key-value хранилище)

##Expected functionality (Ожидаемые функции)
 - TTL для ключа - фиксированное время, задается при старте сервиса.

 
##Expected API (API для работы с хранилищем)
 - Create
 - Get
 - Update
 - Remove
 - List of keys
 
 ##Implemented
 
  - package `proc` containing key-value storage controlled via channels
  - wrapper listening `tcp` on specified port in `main` package
  
  ###Build
  
  `go build -o bin/main main.go`
  
  ###Run
  
  `bin/main [--key-ttl=10000000000 --listen-addr='localhost:12345']`
  
   - take [https://github.com/morrah77/tcp-dialer](https://github.com/morrah77/tcp-dialer) to play with storage `cd .. && git clone https://github.com/morrah77/tcp-dialer && cd tcp-dialer`)
   
   - build it `go build -o bin/main main.go`
   
   - and run `bin/main`
   
   - Now type commands like
   
    -`Get foo bar`
    
    -`Create foo bar`
    
    -`Update foo meow`
    
    -`Remove foo`
    
    -`List`
    
   and see result
  
  ###Test
  
  `go test ./proc/`
  