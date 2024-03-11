FROM amd64/ubuntu:22.04

# Создание пользователя
RUN groupadd -g 501 mygroup && \
    useradd -u 501 -g 501 -m mikhailyudin

# Установка необходимых пакетов
RUN apt-get update && apt-get install -y  \
    build-essential \
    git \
    unzip \
    wget \
    coreutils \
    && rm -rf /var/lib/apt/lists/*

# Переключение на пользователя root для установки Go
USER root

# Установка Go
RUN wget https://go.dev/dl/go1.21.8.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.8.linux-amd64.tar.gz && \
    rm go1.21.8.linux-amd64.tar.gz

# Настройка переменных окружения для Go
ENV GOROOT /usr/local/go

# Переключение на пользователя
USER mikhailyudin
ENV GOPATH /home/mikhailyudin/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH

# Установка рабочего каталога
WORKDIR /opt/musl-cross-make

CMD ["tail", "-f", "/dev/null"]