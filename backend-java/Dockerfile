FROM openjdk:8

WORKDIR /home/gradle/src
COPY gradlew build.gradle ./
COPY gradle ./gradle
RUN ["./gradlew", "--no-daemon", "build"]
COPY . .
CMD ["./gradlew", "--no-daemon", "run"]
