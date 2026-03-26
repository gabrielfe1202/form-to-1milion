# 🚀 form-to-1milion

API de alta performance para captura e processamento de dados de formulários, com foco em escalabilidade, mensageria e testes completos.

---

## 📌 Sobre o projeto

O **form-to-1milion** é um serviço backend desenvolvido em Go com o objetivo de:

- Receber dados de formulários via API
- Validar e processar informações de usuários
- Garantir alta performance e escalabilidade
- Integrar com mensageria (RabbitMQ)
- Preparar a aplicação para alto volume (1 milhão+ requests)

---

## 🧠 Arquitetura

O projeto segue princípios de:

- Clean Architecture
- Separation of Concerns
- Testabilidade (unit + integration)
- Baixo acoplamento

---

## ⚙️ Tecnologias utilizadas

- Go (Golang)
- PostgreSQL
- RabbitMQ
- Testes (unitários e integração)
- k6 (testes de carga)
- Docker / Docker Compose

---

## 🔌 Endpoints

### 📍 Criar usuário

```
POST /api/user
```

#### Body:
```json
{
  "name": "name example",
  "email": "example@email.com",
  "phone": "11999999999",
  "document": "12345678900"
}
```

---

## 🧪 Testes

```bash
go test ./...
```

---

## 🚀 Como rodar o projeto

```bash
make up-and-buld
```

---
