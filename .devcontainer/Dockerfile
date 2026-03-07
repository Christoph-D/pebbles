FROM alpine:3.23

RUN apk add --no-cache bash curl gcompat git go less openssh-client-default ripgrep tzdata make
RUN ln -sf /usr/share/zoneinfo/hostlocaltime /etc/localtime

ARG USERNAME=pebbles
RUN addgroup --gid 1000 $USERNAME && adduser --uid 1000 --ingroup $USERNAME --disabled-password $USERNAME
RUN mkdir -p /home/$USERNAME/persistent-config && chown $USERNAME:$USERNAME /home/$USERNAME /home/$USERNAME/persistent-config

USER $USERNAME
WORKDIR /home/$USERNAME

RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/home/$USERNAME/persistent-config/.bash_history" \
  && echo "$SNIPPET" >> "/home/$USERNAME/.bashrc"

ENV PATH="$PATH:/usr/local/go/bin:/home/$USERNAME/go/bin/"

WORKDIR /workspaces/pebbles
