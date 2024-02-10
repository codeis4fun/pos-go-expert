## Descrição
Desenvolver uma ferramenta de stress test que funcione via CLI e retorne um report com os resultados.

## Requisitos
- [x] A ferramenta deve ser capaz de enviar requisições HTTP através do método GET.
- [x] A ferramenta deve ser capaz de aceitar parâmetros como: URL, número de requests, número de go routines.
- [x] A ferramenta deve ser capaz garantir as requests sejam distribuidas entre as go routines.
- [x] A ferramenta deve ser capaz de retornar um report com o tempo total, quantidade de requests realizadas, quantidade de status com sucesso e demais status HTTP.
- [x] A ferramenta deve ser capaz ser executada via docker.

## Bibliotecas
- Context: Utilizado para cancelar as go routines se por ventura o usuário cancelar a execução do programa.
- Flag: Utilizado para capturar os parâmetros passados via CLI.
- Net/http: Utilizado para realizar as requisições HTTP.
- Time: Utilizado para calcular o tempo total de execução do programa.
- Sync.WaitGroup: Utilizado para garantir que todas as go routines sejam finalizadas antes de retornar o report.
- Sync.Atomic: Utilizado para garantir que as variáveis de contagem sejam acessadas de forma segura pelas go routines.

## Como executar
sem docker
```bash
go run main.go -url=https://www.google.com -requests=100 -goroutines=10
```
com docker
```bash
docker build -t stress-test .
docker run stress-test -url=https://www.google.com -requests=100 -goroutines=10
```

## Exemplo de report
```bash
Report:
All requests finished in 7.737614003s
Total requests: 100
Status code 200: 100
```

## Autor
- [Nícholas Carballo](https://www.linkedin.com/in/nicholascarballo/)
