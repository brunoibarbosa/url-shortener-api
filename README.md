# 📎 URL Shortener (Go + Redis + PostgreSQL)

Um encurtador de URLs simples, rápido e escalável, desenvolvido em **Golang 1.24.2**, com cache em **Redis**, persistência em **PostgreSQL**, e suporte a **i18n (Português/English)**.

---

## 📌 Tecnologias e Arquitetura

- **Go 1.24.2**  
  Backend escrito em Go com foco em performance e concorrência.

- **PostgreSQL 16**  
  Armazena URLs persistentes e metadados relacionados.

- **Redis (com maxmemory-policy: `volatile-lfu`)**  
  Usado como cache para consultas de shortcodes e redirecionamentos, evitando sobrecarga no banco relacional.

- **i18n - Internationalization**  
  Suporte multilíngue: mensagens de erro, respostas de API e validações disponíveis em **Inglês (en)** e **Português (pt-BR)**.

- **Geração de ShortCode com Alta Entropia + Check de Colisão**  
  Algoritmo próprio para gerar shortcodes aleatórios com alta entropia (para reduzir chances de colisões), sempre validando contra o banco antes de salvar.

---

## ⚙️ Funcionalidades principais

- Criar URLs encurtadas.
- Redirecionamento rápido de URLs.
- Cache inteligente com Redis usando **LFU eviction policy**.
- Expiração automática de URLs.
- Suporte a mensagens de erro localizadas via i18n.
- Prevenção de colisões nos códigos encurtados.
- API RESTful.

---

## 🗄️ Configuração do Redis

```bash
maxmemory-policy volatile-lfu
