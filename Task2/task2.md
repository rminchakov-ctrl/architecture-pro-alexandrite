# Выбор и настройка мониторинга в системе

## Мотивация
Есть подозрения в неполадках отдельных частях системы. Чтобы не действовать наугад и терять время (и деньги) стоит сделать систему более наблюдаемой. Необходимо завести метрики, они покажут *состояние здоровья* отдельных компонентов системы, слабые места, которые надо будет исправить.
Это не разовая акция - в дальнейшем можно будет наблюдать потенциальную деградацию под нагрузкой, настроить предупреждения о вероятных сбоях еще до их возникновения, что может помочь сэкономить деньги на их восстановление и предупредить потерю репутации компании. Алертинг позволит заранее видеть слабые места, которые выявляются под нагрузкой (а мы же планируем расширяться) и исправлять возможные сбои еще до их возникновения.

## Выбор подхода к мониторингу
- USE: анализ ресурсов (CPU, память, диски, сеть). Отвечает на вопрос "Ресурс перегружен?"
- RED: анализ сервисов (веб-сервисы, API, микросервисы). Отвечает на вопрос "Сервис работает хорошо для пользователей?"
- Четыре золотых сигнала (FGS): универсальные сигналы для любого сервиса. Отвечают на вопрос "Каково общее состояние сервиса?"

Начать предлагается с FGS - как наиболее универсального подхода к мониторингу сервисов.
Постепенно нужно будет внедрять USE для каждого сервиса, начиная с тех, на которые укажет о проблемах FGS.

FGS показывает *что* плохо, а USE помогает найти *почему* это плохо.

## План действий

Number of dead-letter-exchange letters in RabbitMQ
Number of message in flight in RabbitMQ
Number of requests (RPS) for internet shop API
Number of requests (RPS) for CRM API
Number of requests (RPS) for MES API
Number of requests (RPS) per user for internet shop API
Number of requests (RPS) per user for CRM API
Number of requests (RPS) per user for MES API
CPU % for shop API
CPU % for CRM API
CPU % for MES API
Memory Utilisation for shop API
Memory Utilisation for CRM API
Memory Utilisation for MES API
Memory Utilisation for shop db instance
Memory Utilisation for MES db instance
Number of connections for shop db instance
Number of connections for MES db instance
Response time (latency) for shop API
Response time (latency) for CRM API
Response time (latency) for MES API
Size of S3 storage
Size of shop db instance
Size of MES db instance
Number of HTTP 200 for shop API
Number of HTTP 200 for CRM API
Number of HTTP 200 for MES API
Number of HTTP 500 for shop API
Number of HTTP 500 for CRM API
Number of HTTP 500 for MES API
Number of HTTP 500 for shop API
Number of simultanious sessions for shop API
Number of simultanious sessions for CRM API
Number of simultanious sessions for MES API
Kb tranferred (received) for shop API
Kb tranferred (received) for CRM API
Kb tranferred (received) for MES API
Kb provided (sent) for shop API
Kb provided (sent) for CRM API
Kb provided (sent) for MES API

## Показатели насыщенности
По сути - пороги алертов.
Для хранилища - проценты оставшегося сводбодного места.
Для сервисов - проценты использования CPU и пямяти.