FROM nginx:stable as proxy

ADD nginx.conf /etc/nginx/nginx.conf
ADD mime.types /etc/nginx/mime.types

#host
ADD pages/host.html /usr/share/nginx/pages/host/index.html
ADD pages/login.html /usr/share/nginx/pages/host/login/index.html
#portfolio
ADD pages/portfolio.html /usr/share/nginx/pages/portfolio/index.html
#room automation
ADD pages/room-automation.html /usr/share/nginx/pages/room-automation/index.html
#teachu
ADD pages/teachu.html /usr/share/nginx/pages/teachu/index.html
# error page
ADD pages/error.html /var/www/html/nginx_error/index.html
ADD pages/favicon.ico /var/www/html/nginx_error/favicon.ico

EXPOSE 80

# for extremely detailed logs
#CMD nginx-debug -g 'daemon off;'