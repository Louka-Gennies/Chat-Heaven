# Étape 1 : Construction de l'application Go
FROM golang:1.22-alpine AS builder

# Installer SQLite pour la compilation
RUN apk add --no-cache gcc musl-dev

# Définir le répertoire de travail à l'intérieur du conteneur
WORKDIR /app

# Copier les fichiers go.mod et go.sum et installer les dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source dans le conteneur
COPY . .

# Construire l'application
RUN go build -o main .

# Étape 2 : Création de l'image finale
FROM alpine:latest

# Installer SQLite pour l'exécution
RUN apk add --no-cache sqlite

# Créer un utilisateur non root
RUN adduser -D myuser

# Définir le répertoire de travail à l'intérieur du conteneur
WORKDIR /home/myuser

# Copier l'application construite depuis l'étape 1
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/chatHeaven.db .
COPY --from=builder /app/templates ./templates

# Changer la propriété des fichiers à l'utilisateur non root
RUN chown -R myuser:myuser .

# Changer l'utilisateur pour éviter d'exécuter l'application en tant que root
USER myuser

# Exposer le port sur lequel l'application sera disponible
EXPOSE 8080

# Commande pour exécuter l'application
CMD ["./main"]
