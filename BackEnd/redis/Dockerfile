FROM redis:latest

WORKDIR /data

COPY .env .

COPY . .


RUN chmod +x entrypoint.sh

RUN ./entrypoint.sh



