FROM node:20 AS build-stage

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY ./ /app/

EXPOSE 4200

CMD ["npm", "run", "start", "--", "--host", "0.0.0.0"]