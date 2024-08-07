# deployment-controller
This project is the top layer to my webserver. It features a reverse proxy that routes all the request via an auth request to the expected projects, a backend that handles the auth requests and a small UI for a small overview and a login screen.

## features
- [x] `n` projects behind one domain, with subdomains or locations
- [x] optionally require login to access project
- [x] actually secure login (? :))
- [x] store who access the pages
- [ ] visualize who accessed the pages
- [ ] logging, that makes performance and quality reviews easy
- [ ] send informative mails to the admin and the users

## tech stack
- nginx
- postgres
- go + gin
- sveltekit
- elastic, kibana, logstash

## dev
When developing, it helps to have a reverse proxy setup, so that everything can be tested. To make this process easy, I have created the dev-proxy dir.

### proxy
```bash
docker compose up -d -f ./dev-proxy/docker-compose.yml
```
**subdomains**  
To use subdomains in localhost *(on mac)* I had to modify the `/etc/hosts` file. I added lines like this:
```
127.0.0.1 michu-tech-dev.net
127.0.0.1 host.michu-tech-dev.net
127.0.0.1 host.backend.michu-tech-dev.net
127.0.0.1 teachu.michu-tech-dev.net
127.0.0.1 room-automation.michu-tech-dev.net
```

### k6
[k6](https://k6.io/) is a test application that helps with testing truly parallel requests.  
This needs to be installed on the developers machine, otherwise the tests in `./backend-k6-test/script.js` won't run.

### db
The file `./db/init.sql` creates the DB schema.  
The file `./dev-proxy/test-data.sql` inserts some test pages, so that you can test.  
**A host page must exist. (even the host page has an auth request)**

### ekl
...

### backend
The backend can usually be started in the IDE or with `go run main.go`.  
In production the app will run in a docker container. To test the container run the following.  
**The docker container uses the config-docker.yml config!**
```bash
docker build --tag deployment_controller_dev_backend ./backend
```
```bash
docker run -p 8080:8080 --name dp_crtl_be deployment_controller_dev_backend
```

### frontend
The frontend can be started and built with the following command.  
Make sure that a .env file exists with the `PUBLIC_BACKEND_URL` config. 
```bash
# run the dev builds
npm run dev

# compile for production
npm run build
```