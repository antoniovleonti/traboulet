FROM ubuntu:latest

# nginx version check here "http://nginx.org/download/" for available versions
ENV NGINX_VERSION 1.25.2

ENV NGINX_PUSH_STREAM_MODULE_PATH /Sources/nginx-push-stream-module

RUN apt-get update \
    && apt-get install -y \
       wget \
       git \
       gcc \
       build-essential \
       libc6 \
       libpcre3 \
       libpcre3-dev \
       libssl-dev \
       zlib1g \
       zlib1g-dev \
       lsb-base

RUN git clone https://github.com/wandenberg/nginx-push-stream-module.git \
    && wget http://nginx.org/download/nginx-${NGINX_VERSION}.tar.gz \
    && tar xzvf nginx-${NGINX_VERSION}.tar.gz \
    && cd nginx-${NGINX_VERSION} \
    && ./configure --add-module=../nginx-push-stream-module \
    && make \
    && make install


# ADD Command copy the files from source on the host to destination on the
# container's file system Usage: ADD source destination
COPY nginx.conf /usr/local/nginx/conf/nginx.conf
COPY client /usr/share/nginx/serve-files

RUN apt-get purge -y \
    wget \
    git \
    gcc \
    && apt-get clean

EXPOSE 25566
CMD ["/usr/local/nginx/sbin/nginx", "-g", "daemon off;"]
