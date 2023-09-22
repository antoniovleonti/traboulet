FROM nginx
RUN rm /etc/nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY client /usr/share/nginx/serve-files
COPY nginx.conf /etc/nginx
