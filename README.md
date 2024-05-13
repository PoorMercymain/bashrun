# bashrun
Сервис для запуска команд (bash скриптов)

# Как запустить
Либо переименовав (или составив по образцу) .env.example в .env, и запустить командой `docker-compose up`, либо передав .env.example в качестве значения флага --env_file для `docker-compose up`. После этого сервис будет запущен на соответствующем порту (по умолчанию 8080), после того, как БД пройдет healthcheck. На рисунке ниже показаны логи, после вывода которых уже можно обращаться к сервису

![изображение](https://github.com/PoorMercymain/bashrun/assets/67076111/187bb177-bf5e-4c71-9bd7-59d931267d12)

# Swagger
Для того, чтобы получить доступ к Swagger UI, можно обратиться по `/swagger`. Там Эндпойнты описаны более подробно

![изображение](https://github.com/PoorMercymain/bashrun/assets/67076111/6b661df3-f3f7-46fb-bde0-b929fcabaaf0)

# Postman коллекция
Также в репозитории присутствует <a href="https://github.com/PoorMercymain/bashrun/blob/main/bashrun.postman_collection.json">postman коллекция</a> с примерами запросов

<p align="center"><img src="https://github.com/PoorMercymain/bashrun/assets/67076111/26700dc7-b071-4c10-a60a-54ad883560ab"></p>

# Миграции
Миграции находятся в директории <a href="https://github.com/PoorMercymain/bashrun/tree/main/migrations">migrations</a>. В них производится создание таблицы и индекса
<p align="center"><img src="https://github.com/PoorMercymain/bashrun/assets/67076111/3dad7380-6228-4354-af99-68493c222f4f"></p>

# Docker
Для развертывания сервиса, есть <a href="https://github.com/PoorMercymain/bashrun/blob/main/Dockerfile">Dockerfile</a>, используемый в <a href="https://github.com/PoorMercymain/bashrun/blob/main/docker-compose.yml">docker-compose</a> (указан соответствующий контекст). Также, в docker-compose развертывается и БД

# Используемый дистрибутив Linux
Использовал:
- Ubuntu 22.04.1 LTS (для запуска тестов)
- Alpine (для запуска сервиса в контейнере)

# Тесты
Тесты можно запустить с помощью команды `go test ./...`. Для проверки покрытия можно использовать команду `go tool cover -func cov.profile` (cov.profile также находится в репозитории, но также его можно собрать самостоятельно). Часть вывода после использования этой команды представлена на рисунке ниже

![изображение](https://github.com/PoorMercymain/bashrun/assets/67076111/c7fddc39-05b7-444d-bd9f-5eb9126d48c6)

# ТЗ
Приложение должно иметь базу данных для хранения команд ✓ (используется БД под управлением PostgreSQL)

API приложения должно содержать следующий функционал:

- Создание новой команды. Запускает переданную bash-команду, сохраняет результат выполнения в БД. ✓ (эндпойнт `POST /commands`)
- Получение списка команд ✓ (эндпойнт `GET /commands`)
- Получение одной команды ✓ (эндпойнт `GET /commands/{command_id}`)

Написать тесты. ✓

Написать инструкцию по запуску программы. ✓

# Эндпойнты
`GET /ping` - пинг БД

<p align="center"><img src="https://github.com/PoorMercymain/bashrun/assets/67076111/762f1c99-6100-4f94-bb7e-874e6004bfdc"></p>

`POST /commands` - создание и запуск команды (в отдельной горутине, с семафором в качестве ограничителя числа одновременно выполняющихся команд, его "вес" настраивается с помощью `MAX_CONCURRENT_COMMANDS` в .env файле)

<p align="center"><img src="https://github.com/PoorMercymain/bashrun/assets/67076111/919facbc-e40f-466d-8dd8-224d7a738528"></p>

`GET /commands` - получение списка команд (также в query можно указать limit и offset, по умолчанию они 15 и 0 соответственно)

<p align="center"><img src="https://github.com/PoorMercymain/bashrun/assets/67076111/fa201430-33d2-4885-bd7f-800c3a60ab94"></p>

`GET /commands/stop/{command_id}` - остановка команды (для минимизации обращений к ОС используется singleflight)

<p align="center"><img src="https://github.com/PoorMercymain/bashrun/assets/67076111/e684a3b6-1fbd-4850-93fc-30fab5c13d6b"></p>

`GET /commands/{command_id}` - получение одной команды по id

<p align="center"><img src="https://github.com/PoorMercymain/bashrun/assets/67076111/3a299a7a-65a7-4aec-b3a3-f1420239cc51"></p>

`GET /commands/output/{command_id}` - получение вывода команды (вывод обновляется по мере выполнения скрипта)

<p align="center"><img src="https://github.com/PoorMercymain/bashrun/assets/67076111/5ad169e4-8e8a-44c2-9e8b-56472391a85f"></p>
