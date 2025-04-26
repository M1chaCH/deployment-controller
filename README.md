# deployment-controller
This project is the top layer to my webserver. It features a reverse proxy that routes all the requests via an auth request to the expected projects, a backend that handles the auth requests and a small UI for a small overview and a login screen.

Also, there is a whole elastic stack behind it with multiple agents monitoring and observing the system.
At this point it can be tracked what every single client does on my servers. (A client is a set of devices, and a device is an ip address and an agent)

## features
- [x] `n` projects behind one domain, with subdomains or locations
- [x] optionally require login to access the project
- [x] actually secure login (? :))
  - [x] SQL injection protection
  - [x] XSS protection
  - [x] secure JWT
  - [x] secure hash for password store
  - [x] MFA
  - [x] no session fixation
  - [x] secure to CSRF
  - [x] ABAC
  - [ ] protection against brute force
- [x] store who accessed the pages
- [x] attempt to track a single user across multiple devices (when on my servers)
- [x] visualize who accessed the pages (kibana)
  - [ ] implement mini dashboard on the admin page on the host site
- [x] logging, that makes performance and quality reviews easy (kibana)
- [ ] send informative mails to the admin and the users
- [x] (close to) completely stateless backend

## tech stack
- nginx
- docker
- postgres
- go + gin + logrus
- sveltekit
- elastic, kibana, filebeat, fleet, agents, apm

## dev
When developing, it helps to have a reverse proxy setup so that everything can be tested. To make this process easy, I have created the dev-env dir.

### proxy
```bash
docker compose up -d -f ./dev-env/docker-compose.yml
```
**subdomains**  
To use subdomains in localhost *(on mac)* I had to modify the `/etc/hosts` file. I added lines like this:
```
127.0.0.1 michu-tech-dev.net
127.0.0.1 host.michu-tech-dev.net
127.0.0.1 host.backend.michu-tech-dev.net
127.0.0.1 teachu.michu-tech-dev.net
127.0.0.1 room-automation.michu-tech-dev.net
127.0.0.1 kibana.michu-tech-dev.net
```

### k6
[k6](https://k6.io/) is a test application that helps with testing truly parallel requests.  
This needs to be installed on the developers machine, otherwise the tests in `./backend-k6-test/script.js` won't run.

### db
The file `./db/init.sql` creates the DB schema.  
The file `./dev-proxy/test-data.sql` inserts some test pages, so that you can test.  
**A host page must exist. (even the host page has an auth request)**

### elastic
Using Elasicsearch, Kibana, Fleet and so on to keep an overview of my system.

Here is a list of things that are in use on my elastic stack.
- [Nginx Metrics and Logs](https://www.elastic.co/docs/current/en/integrations/nginx)
- [Postgres Metrics](https://www.elastic.co/docs/current/integrations/postgresql)
- [Docker Metrics](https://www.elastic.co/docs/current/en/integrations/docker)
- [Docker Container logs](https://www.elastic.co/guide/en/beats/filebeat/current/filebeat-input-container.html)
  - since I have a lot of projects that just run in a docker container, it will be very usefull to see the all the logs without having to SSH onto the server.
- [Backend with APM](https://www.elastic.co/guide/en/apm/agent/go/master/introduction.html)
  - Requests should be logged
  - Errors should be reported
  - (maybe also log DB Queries, Postgres integration might be enough)
- [Agent on Raspberry PI](https://www.elastic.co/guide/en/fleet/current/install-fleet-managed-elastic-agent.html)
  - Get metrics and maybe logs from my raspberry pi
- Helpful Kibana dashboards
- Security or error alerts

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

#### configuration
The backend can be configured in the config.yml file in its root directory.
When running in production, this config can be overwritten by the environment variables.
The environment variables should be written in uppercase camel case.

#### filebeat / elastic formatted logs
Currently, the backend writes its logs to a JSON file. The JSON structure matches elastic formatting. This makes further analysis in kibana rather easy.
Filebeat is the tool that keeps track of the logfile from the backend and sends the new lines to elasticsearch.

### frontend
The frontend can be started and built with the following command.  
Make sure that a .env file exists with the `PUBLIC_BACKEND_URL` config. 
```bash
# run the dev builds
npm run dev

# compile for production
npm run build
```