FROM node:latest AS build-frontend

WORKDIR /frontend

ADD ./frontend .
RUN npm install
RUN npm run build

FROM nginx:latest AS proxy

ADD nginx.conf /etc/nginx/nginx.conf
ADD mime.types /etc/nginx/mime.types
ADD dev/test-automation.html /pages/automation/index.html
ADD dev/test-portfolio.html /pages/portfolio/index.html

COPY --from=build-frontend /frontend/public /etc/nginx/html

EXPOSE 80
EXPOSE 443