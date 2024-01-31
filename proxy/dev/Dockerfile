FROM nginx:latest AS proxy

ADD test-automation.html /pages/automation/index.html
ADD test-portfolio.html /pages/portfolio/index.html
ADD nginx-dev.conf /etc/nginx/nginx.conf
ADD mime.types /etc/nginx/mime.types

EXPOSE 80