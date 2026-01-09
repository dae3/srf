FROM debian:trixie

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
ARG LANG=en_AU.UTF-8

# 1. Install System Dependencies & System Python
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl git tmux wget unzip ca-certificates zsh locales openssh-client \
    vim ripgrep fzf procps zip make gcc flex bison gperf \
    libncurses5-dev libncursesw5-dev pkg-config \
    python3 python3-pip python3-setuptools python-is-python3 sudo

# 2. Install Python dependencies globally (No virtualenv)
# Using --break-system-packages because this is a dedicated container
RUN pip3 install --break-system-packages \
    pyserial future cryptography==3.4.8 pyparsing==2.3.1 pyelftools

# 3. Provision Xtensa Toolchain
RUN mkdir -p /opt/espressif && \
    curl -L https://dl.espressif.com/dl/xtensa-lx106-elf-gcc8_4_0-esp-2020r3-linux-amd64.tar.gz \
    | tar -xz -C /opt/espressif
ENV PATH="/opt/espressif/xtensa-lx106-elf/bin:${PATH}"

# 4. Setup Locales
RUN sed -i "s/^# \($LANG\)/\1/" /etc/locale.gen && locale-gen
ENV LANG=$LANG
ENV LC_ALL=$LANG

# 5. Global Environment Variables
RUN echo 'export PATH="/opt/espressif/xtensa-lx106-elf/bin:$PATH"' > /etc/profile.d/esp8266-env.sh && \
    echo 'export CHECK_PY_DEPS=0' >> /etc/profile.d/esp8266-env.sh && \
    echo 'export IDF_PATH="/workspaces/hardware/sdk"' >> /etc/profile.d/esp8266-env.sh
RUN echo 'source /etc/profile.d/esp8266-env.sh' >> /etc/zsh/zprofile

# 6. Setup User
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME -s /usr/bin/zsh \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

USER $USERNAME
WORKDIR /workspaces/hardware
