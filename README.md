# API de Notícias

Este projeto é uma API simples de notícias desenvolvida em Go usando o framework Gin e documentada com Swagger.

## Funcionalidades

A API oferece as seguintes operações:

- Listar todas as notícias
- Obter uma notícia específica por ID
- Criar uma nova notícia
- Atualizar uma notícia existente
- Deletar uma notícia

## Tecnologias Utilizadas

- Go
- Gin (framework web)
- UUID (para geração de IDs únicos)
- Swagger (para documentação da API)

## Instalação

1. Certifique-se de ter Go instalado em sua máquina.
2. Clone este repositório:
3. Instale as dependências:

## Uso

1. Execute o servidor:
2. O servidor estará rodando em `http://localhost:8080`

## Rotas da API

- `GET /news`: Lista todas as notícias
- `GET /news/:id`: Obtém uma notícia específica por ID
- `POST /news`: Cria uma nova notícia
- `PUT /news/:id`: Atualiza uma notícia existente
- `DELETE /news/:id`: Remove uma notícia

## Documentação Swagger

A documentação da API está disponível através do Swagger UI. Para acessá-la:

1. Certifique-se de que o servidor está rodando
2. Acesse `http://localhost:8080/swagger/index.html` em seu navegador

## Estrutura do Projeto

- `main.go`: Arquivo principal contendo toda a lógica da API
- `docs/`: Diretório contendo os arquivos gerados pelo Swagger

## Gerando Documentação Swagger

Para gerar ou atualizar a documentação Swagger, execute:

## Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para abrir issues ou enviar pull requests.
