```markdown
# Tamis 🧹

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Backend-Go-brightgreen)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Frontend-Next.js-blue)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-TS-blue)](https://www.typescriptlang.org/)
[![Docker](https://img.shields.io/badge/Docker-Containers-lightblue)](https://www.docker.com/)

**Tamis** est un outil moderne et sécurisé pour nettoyer vos emails de manière simple et efficace.  
Il unifie plusieurs comptes mail (Gmail, Yahoo, comptes scolaires) et permet de filtrer et supprimer les emails indésirables ou doublons en toute sécurité.

---

## 📌 Table des matières

1. [Fonctionnalités](#-fonctionnalités)
2. [Architecture](#-architecture)
3. [Tech Stack](#-tech-stack)
4. [Installation & Setup](#-installation--setup)
5. [Développement](#-développement)
6. [Contribution](#-contribution)
7. [Roadmap](#-roadmap)
8. [Licence](#-licence)

---

## 🚀 Fonctionnalités V1

- 🔑 Connexion multi-comptes (Gmail, Yahoo, école)
- 📋 Listing et filtrage des mails (doublons, newsletters)
- 👀 Prévisualisation avant suppression
- 🛡 Suppression sécurisée avec logs
- 🎨 Interface moderne et ergonomique (Next.js + TypeScript)
- ⚡ Backend performant et modulaire en Go

---

## 🏗️ Architecture

```

Tamis/
├─ client/                # Frontend Next.js + TS
│  ├─ src/
│  │   ├─ components/     # Composants UI réutilisables
│  │   ├─ app/          # Pages Next.js
│  │   ├─ services/       # API calls
│  │   └─ types/          # Types TS
├─ server/                # Backend Go
│  ├─ cmd/                # main.go
│  ├─ internal/
│  │   ├─ api/            # Handlers HTTP
│  │   ├─ services/       # Logique métier (IMAP, filtres)
│  │   ├─ repository/     # Accès DB
│  │   ├─ models/         # Structs Mail / Account / Logs
│  │   └─ utils/          # Helpers (logger, config)
├─ docker-compose.yml
└─ .env

````

> Modulaire, scalable et maintenable dès la V1.  

---

## 🛠️ Tech Stack

| Partie        | Technologie               |
|---------------|--------------------------|
| Frontend      | Next.js + TypeScript + TailwindCSS |
| Backend       | Go + IMAP/OAuth2         |
| Base de données | PostgreSQL (Docker)      |
| Conteneurisation | Docker                   |

---

## ⚡ Installation & Setup

1. **Cloner le projet**  

```bash
git clone https://github.com/<votre-utilisateur>/Tamis.git
cd Tamis
````

2. **Copier et configurer `.env`**

```bash
cp .env.example .env
# Remplir les variables DB et tokens
```

3. **Lancer Docker**

```bash
docker-compose up --build
```

* Backend : `http://localhost:5000`
* Frontend : `http://localhost:3000`

---

## 🧩 Développement

### Backend Go

```bash
cd server
go run cmd/main.go
```

* Structure modulaire avec `api`, `services`, `repository`, `models`, `utils`
* Endpoints :

  * `POST /login` → connecter un compte mail
  * `GET /emails` → lister mails filtrés
  * `POST /delete` → suppression sécurisée

### Frontend TS / Next.js

```bash
cd client
npm install
npm run dev
```

* Pages : `/login`, `/dashboard`
* Components : `MailList`, `MailItem`, `DeleteButton`, `ProviderSelector`

---

## 🤝 Contribution

Les contributions sont les bienvenues !

```bash
# Fork
git checkout -b feature/nom-de-la-fonctionnalité
git commit -m "feat: description courte"
git push origin feature/nom-de-la-fonctionnalité
# Ouvrir un PR
```

**Règles :**

* Code clair et commenté
* Respect conventions Go (backend) et TS (frontend)
* Tests unitaires si possible
* `main` = stable, branches features pour dev

---

## 📈 Roadmap

**V1 :**

* Connexion multi-comptes
* Liste et filtrage des mails
* Suppression sécurisée + logs
* Dashboard ergonomique

**V2+ :**

* Scheduler automatique
* Classification intelligente des mails
* Support OAuth complet pour tous les fournisseurs
* Multi-device responsive

---

## 📄 Licence

MIT License – voir [LICENSE](LICENSE)

---

> Tamis est conçu pour être **sécurisé, simple et efficace**, tout en restant un projet open-source moderne et collaboratif pour développeurs.