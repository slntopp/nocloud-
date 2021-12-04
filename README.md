### Клиент-серверный файловый сервис на базе сокетов или grpc
Решает задачу чтения параметров клиента по запросу с сервера. 


#### Порядок работы:
- Запустить в отдельных окнах терминала в указанной последовательности \
для сокетов
```bash
go run cmd/soketserver/main.go
go run cmd/soketclient/main.go
```
или для grpc
```bash
go run cmd/grpcserver/main.go
go run cmd/grpcclient/main.go
```
- На терминале сервера появляется строка приглашения message to client **m2c >**.
- После ввода любой строки (набор строки плюс Enter) на экране клиента появляется отклик
```bash
Hello from server! hhh
```
- На экране сервера появляется отклик клиента
```bash
Hello from client to server! hhh
```
- Выход Ctrl-C