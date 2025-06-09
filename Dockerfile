FROM scratch
ARG TARGETARCH
ARG IMAGE_VERSION

COPY ./dist/repoxy-linux-${TARGETARCH} /bin/repoxy
ENTRYPOINT ["/bin/repoxy"]
LABEL org.opencontainers.image.source="https://github.com/davidjspooner/repoxy"
LABEL org.opencontainers.image.description="Repoxy is a proxy server for various repositories"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.version="${IMAGE_VERSION}"
LABEL org.opencontainers.image.architecture="${TARGETARCH}"