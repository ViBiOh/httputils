FROM scratch

COPY passwd /etc/passwd
USER 10000

ARG TARGETOS
ARG TARGETARCH

COPY release/wait_${TARGETOS}_${TARGETARCH} /wait
