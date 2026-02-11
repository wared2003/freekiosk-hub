# 🚀 FreeKiosk Hub

Le cerveau central pour la gestion des tablettes FreeKiosk. Ce hub surveille l'état des batteries, gère la connectivité via Tailscale (optionnel) et expose une interface de contrôle.

## 🛠 Architecture
- **Langage** : Go (Golang)
- **Base de données** : SQLite (via `sqlx` & `glebarez/go-sqlite`)
- **Web** : Echo Framework
- **Réseau** : Tailscale `tsnet` pour l'accès distant sécurisé

## 🚀 Démarrage Rapide

1. **Configuration** :
   ```bash
   cp .env.example .env
   # Remplis ton TS_AUTHKEY si tu utilises Tailscale