FROM debian:trixie

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
ARG LANG=en_AU.UTF-8
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    git \
    tmux \
    wget \
    unzip \
    ca-certificates \
    zsh \
    python3 \
    locales \
    openssh-client \
    vim \
    ripgrep \
    fzf \
    exuberant-ctags \
    procps

RUN sed -i "s/^# \($LANG\)/\1/" /etc/locale.gen && locale-gen
ENV LANG=$LANG

RUN curl -sfL https://go.dev/dl/go1.24.6.linux-amd64.tar.gz | tar -C /usr/local -zxf -
ENV PATH=/usr/local/go/bin:${PATH}
RUN GOPATH=/usr/local/go go install golang.org/x/tools/gopls@latest
RUN GOPATH=/usr/local/go go install github.com/air-verse/air@latest
RUN GOPATH=/usr/local/go go install github.com/go-delve/delve/cmd/dlv@latest

RUN apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME

RUN chown vscode /home/vscode
RUN chsh -s $(which zsh) vscode

USER vscode

RUN curl -L https://fly.io/install.sh | sh
ENV PATH=${PATH}:/home/vscode/.fly/bin

RUN curl -LsSf https://astral.sh/uv/install.sh | sh
RUN ${HOME}/.local/bin/uv tool install --force --python python3.12 aider-chat@latest
RUN ${HOME}/.local/bin/uv tool install httpie

EXPOSE 8080
