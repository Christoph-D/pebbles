FROM alpine:3.23

RUN apk add --no-cache bash curl gcompat git go less openssh-client-default ripgrep tzdata make
RUN ln -sf /usr/share/zoneinfo/hostlocaltime /etc/localtime

ARG USERNAME=pebbles
RUN addgroup --gid 1000 $USERNAME && adduser --uid 1000 --ingroup $USERNAME --disabled-password $USERNAME
RUN mkdir -p /home/$USERNAME && chown $USERNAME:$USERNAME /home/$USERNAME

USER $USERNAME
WORKDIR /home/$USERNAME

RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/home/$USERNAME/persistent-config/.bash_history" \
  && echo "$SNIPPET" >> "/home/$USERNAME/.bashrc"

ENV PATH="$PATH:/usr/local/go/bin:/home/$USERNAME/go/bin/"

ENV USE_BUILTIN_RIPGREP=0
ENV CLAUDE_CONFIG_DIR=/home/$USERNAME/persistent-config
RUN curl -fsSL https://claude.ai/install.sh | bash
ENV PATH="/home/$USERNAME/.local/bin:$PATH"

ENV OPENCODE_CONFIG=/home/$USERNAME/persistent-config/opencode.json
RUN curl -fsSL https://opencode.ai/install | bash
ENV PATH="/home/$USERNAME/.opencode/bin:$PATH"

ENV EDITOR='code --wait'
WORKDIR /workspaces/pebbles
