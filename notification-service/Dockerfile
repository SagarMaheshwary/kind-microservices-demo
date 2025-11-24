FROM node:22 AS builder

WORKDIR /app

COPY ./package*.json .

RUN npm i

COPY . /app

RUN npm run build

FROM node:22-alpine AS production

WORKDIR /app

COPY --from=builder /app/dist /app/dist
COPY --from=builder /app/node_modules /app/node_modules
COPY ./package*.json /app/

EXPOSE 4000
CMD [ "npm", "run", "start:prod" ]

FROM node:22-alpine AS development

WORKDIR /app

COPY --from=builder /app /app

#install curl for docker compose healthcheck
RUN apk add curl 

EXPOSE 4000
CMD [ "npm", "run", "start:dev" ]
