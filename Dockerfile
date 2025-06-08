FROM scratch
ARG TARGETARCH

COPY ./dist/repoxy-linux-${TARGETARCH} /bin/repoxy
ENTRYPOINT ["/bin/repoxy"]
LABEL org.opencontainers.image.source="https://github.com/davidjspooner/repoxy"