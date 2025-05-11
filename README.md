# Распределённый вычислитель арифметических выражений

### Описание
Проект содержит в себе многопользовательский режим (регистрация и вход), хранение выражений в SQLite, общение вычислителя и сервера вычислений реализовано с помощью GRPC, проект покрыт модульными тестами.



## Запуск 
Пропишите go mod tidy находясь в корневой папке
#### Через git Bash
С помощью
``` bash
 git clone https://github.com/MrM2025/Project-3/tree/main/Sprint_2/calc_go
 ```
сделайте клон проекта.

Запустите Оркестратор:
#### Важно - проверьте, что вы находитесь в папке calc_go.

``` bash
go run cmd/Orchestrator_start/main.go
```

### Агент запускать не нужно(он запускается автоматически). 

# Для отправки curl используйте Postman

Выражение для вычисления должно передаваться в JSON-формате, в единственном поле "expression", если поле отсутствует - сервер вернет ошибку 422, "Empty expression"; если в запросе будут поля, отличные от "expression" - сервер вернет ошибку 400, "Bad request" также как и при отсуствии JSON'а в теле запроса;

Должны быть установлены Go и Git.

## Пример запроса с использованием curl(Рекомендую использовать Postman)



Для git bash:

```bash
Регистрация: 
    curl --location 'localhost:8080/api/v1/register' --header 'Content-Type: application/json' --data '{"login": "User", "Password": "123"}'
```

Ожидаемый ответ: 
{
    "status": "Successful sign up"
}

```bash
Вход:
    curl --location 'localhost:8080/api/v1/login' --header 'Content-Type: application/json' --data '{"login": "User", "Password": "123"}'
```
Ожидаемый ответ: 
{
    "status": "Successful sign in",
    "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY5OTU5NTMsImlhdCI6MTc0Njk5NTY1MywibmFtZSI6IlVzZXIiLCJuYmYiOjE3NDY5OTU2NTN9.psRkEawQV0wz-T3XtxEqlfRCJeJ_9TT1LSv7vuj2FgA"
}

``` bash
Передача сервису выражение на вычисление:
#!!! Важно: в поле jwt, нужно вставить токен, который был
#!!! выдан при входе, иначе ничего не получится
# (срок жизни сессии - 5 минут)
    curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{ "expression": "1-1+1", "login": "User", "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY5OTQ3MjksImlhdCI6MTc0Njk5NDQyOSwibmFtZSI6IlVzZXIiLCJuYmYiOjE3NDY5OTQ0Mjl9.ILkn2O7HA-UFIPYZ8ed4Ab08vHx-vF8Wf29IKRHTjkE"}'
```
Ожидаемый ответ:
{
    "id": "1"
}
``` bash
#!!! Важно: в поле jwt, нужно вставить токен, который был
#!!! выдан при входе, иначе ничего не получится
# (срок жизни сессии - 5 минут)
Просмотр выражения по его ID:
    curl --location 'localhost:8080/api/v1/expression/id' --header 'Content-Type: application/json' --data '{ "id": "1", "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY5OTQ3MjksImlhdCI6MTc0Njk5NDQyOSwibmFtZSI6IlVzZXIiLCJuYmYiOjE3NDY5OTQ0Mjl9.ILkn2O7HA-UFIPYZ8ed4Ab08vHx-vF8Wf29IKRHTjkE" }'
```
Ожидаемый ответ: 
{
    "expression": [
        {
            "id": "1",
            "expression": "1-1+1",
            "login": "User",
            "status": "completed",
            "result": 1
        }
    ]
}
``` bash
#!!! Важно: в поле jwt, нужно вставить токен, который был
#!!! выдан при входе, иначе ничего не получится
# (срок жизни сессии - 5 минут)
Передача сервису выражение на вычисление:
    curl --location 'localhost:8080/api/v1/expressions' --header 'Content-Type: application/json' --data '{ "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDY5OTQ3MjksImlhdCI6MTc0Njk5NDQyOSwibmFtZSI6IlVzZXIiLCJuYmYiOjE3NDY5OTQ0Mjl9.ILkn2O7HA-UFIPYZ8ed4Ab08vHx-vF8Wf29IKRHTjkE"}'
```
Ожидаемый ответ: 
{
    "expression": [
        {
            "id": "1",
            "expression": "1-1+1",
            "login": "User",
            "status": "completed",
            "result": 1
        }
    ]
}

#

Postman:

https://identity.getpostman.com/signup?deviceId=c30fc039-7460-4f58-8cb9-b74256c4186c  

^

|

Регистрация

https://www.postman.com/downloads/

^

|

Ссылка на скачивание приложения.

#
Мануал №1 - https://timeweb.com/ru/community/articles/kak-polzovatsya-postman

Мануал №2 - https://blog.skillfactory.ru/glossary/postman/

Мануал №3 - https://gb.ru/blog/kak-testirovat-api-postman/

## Примеры использования (cmd Windows)

### * Важно: при отображении readme в HTLM, экранирующие слэши не отображаются, поэтому копировать команды лучше из raw-формата, либо самостоятельно экранировать ковычки в json'е слэшом слева, иначе получите ошибку!

Верно заданный запрос, Status: 200
```
curl -i -X POST -H "Content-Type:application/json" -d "{\"expression\": \"20-(9+1)\"}" http://localhost:8080/api/v1/calculate
```
Результат:
{
    "expressions": [
        {
            "id": "1",
            "expression": "20-(9+1)",
            "status": "completed",
            "result": 10
        }
    ]
}


Запрос с пустым выражением, Status: 422, Error: empty expression
```
curl -i -X POST -H "Content-Type:application/json" -d "{\"expression\": \"\"}" http://localhost:8080/api/v1/calculate
```
Ответ:
{
    "error": "empty expression"
}

Запрос с делением на 0, Status: 422, Error: division by zero
```
curl -i -X POST -H "Content-Type:application/json" -d "{\"expression\": \"1/0\"}" http://localhost:8080/api/v1/calculate
```
Ответ: {"error":"division by zero"} 
(Появляется после того, как Агент попытаеться вычислить выражение)

Запрос неверным выражением, Status : 422, Error: invalid expression
```
curl -i -X POST -H "Content-Type:application/json" -d "{\"expression\": \"1++*2\"}" http://localhost:8080/api/v1/calculate
```
Ответ:
{
    "error":"incorrect expression | wrong sequence \"operation sign-\u003eoperation sign\": chars 1, 2 | wrong sequence \"operation sign-\u003eoperation sign\": chars 2, 3 "
}

## Тесты
Для тестирования перейдите в файл agent_calc_test.go и используйте команду go test или(для вывода дополнительной информации) go test -v

Для запусков всех тестов разом воспользуйтесь - go test ./...

