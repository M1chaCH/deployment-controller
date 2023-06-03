# deployment-controller
I use this project to differentiate between deployments that should be publicly visible and secured ones. 

Url to my server: [michu.tech](https://michu.tech)

## details
I have one server at my home and I want to use it to deploy multiple projects. I'll use docker for this. Now, some of these projects can be visible and useable for everyone. But others I plan to secure behind a login. Because I want to deploy certain projects that contain sensitive data or have access to devices that should not be publicly accessed (like [room-automation](https://github.com/M1chaCH/room-automation)). 

This project will include a docker compose config, an nginx config for the routing and securing of everything, a helidon microservice for authorisation and a microservice for authentication. 
