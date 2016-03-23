FROM ubuntu:latest

COPY livelyweb-firmware /liveliyweb-firmware

ENTRYPOINT ["/liveliyweb-firmware"]
