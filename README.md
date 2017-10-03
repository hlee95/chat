# challenge-eng-base

This starter kit currently supports `React` for the frontend and `Go`, `Python`, `Java`, or `Node` for the backend.

To get the project up and running:
1. Install Docker https://docs.docker.com/engine/installation/
2. In a terminal, go to the directory `challenge-eng-base-master`
3. Edit `docker-compose.yml`. Change `services: backend: build:` based on your preferred language. Options are `backend-golang`, `backend-python`, `backend-java`, or `backend-node`.
4. For a backend project
    1. `docker-compose up backend`
    2. Test that it's running http://localhost:18000/test
5. For a fullstack project
    1. `docker-compose up fullstack`
    2. Test that it's running http://localhost:13000/test

To restart the project

    docker-compose down
    docker-compose up <backend or fullstack>

To see schema changes, remove the old db volume by adding `-v` when stopping

    docker-compose down -v

To see code changes, rebuild by adding `--build` when starting

    docker-compose up --build <backend or fullstack>

If you run into issues connecting to the db on startup, try restarting (without the `-v` flag).

## Sample cURL commands

To create a new user:

    curl -i -d '{"username":"user1", "password":"super-secret"}' -H "Content-Type: application/json" -X POST localhost:18000/users

To send a message:

    curl -i -d '{"sender":"user2", "recipient":"user1", "messageType":"plaintext", "content":"Hi there!"}' -H "Content-Type: application/json" -X POST localhost:18000/messages

where `messageType` is one of `"plaintext"`, `"image_link"` or `"video_link"`.

Example of an `image_link` message:

    curl -i -d '{"sender":"user2", "recipient":"user1", "messageType":"image_link", "content":"https://www.what-dog.net/Images/faces2/scroll0015.jpg"}' -H "Content-Type: application/json" -X POST localhost:18000/messages

To fetch a conversation:

    curl -i "localhost:18000/messages?sender=user1&recipient=user2&messagesPerPage=2&pageToLoad=1"

where `messagesPerPage` and `pageToLoad` are optional and can be usd for pagination, and `pageToLoad` is 0-indexed.

