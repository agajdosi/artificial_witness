# Artificial Witness

**Artificial Witness** is an experimental art game that explores the biases and prejudices of different AI models.
Rather than analyzing training datasets or statistical outputs, this game allows players to experience AI biases firsthand through modified gameplay of *Unusual Suspects* board game.

## Contributing

You can contribute to the project by reporting issues, suggesting features, giving feedback and also sharing it on the net!

For code contributions, please check the DEVELOPING chapter.
Project is not fully prepared for comfortable external contributions, but if you are brave enough, feel free to fork and try it out!

## DEVELOPING

### Architecture

Frontend is written in SvelteKit, backend is written in Go, data lies in SQLite3 database.
- only backend calls the LLM services, secrets lies in backend SQLite3 database
- frontend communicates only with backend, no secret keys are stored in frontend
- AI model is selected at the start of the game and is used for the whole game

### Frontend server

```
cd front
npm run dev
```

### Backend server
```
cd backend
go run main.go
```

## Deployment

### Build Backend Docker Image

Run the Docker build from the project root as we need access to go.mod and go.sum files,
use --file flag to specify the Dockerfile in backend directory:

```bash
VERSION="v1.0.1"
IMAGE="agajdosi/artificial_witness"

docker build \
  -t ${IMAGE}:${VERSION} \
  -t ${IMAGE}:latest \
  --platform linux/amd64 \
  --file backend/Dockerfile .

docker push ${IMAGE}:${VERSION}
docker push ${IMAGE}:latest
```

## Acknowledgments
A huge thanks to **SvelteKit** for making this possible! 🎉
