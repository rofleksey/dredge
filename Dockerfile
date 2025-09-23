FROM golang:1.25-alpine AS apiBuilder
WORKDIR /opt
RUN apk update && apk add --no-cache make
COPY . /opt/
RUN go mod download
ARG GIT_TAG=?
RUN make build GIT_TAG=${GIT_TAG}

FROM node:22 AS frontBuilder
WORKDIR /opt/frontend
COPY ./frontend/ .
RUN npm i && npm run build

FROM alpine
ARG ENVIRONMENT=production
WORKDIR /opt
RUN apk update && apk add --no-cache curl ca-certificates
COPY --from=apiBuilder /opt/dredge /opt/dredge
COPY --from=frontBuilder /opt/frontend/dist /opt/frontend/dist
EXPOSE 8080
RUN ulimit -n 100000
CMD [ "./dredge" ]
