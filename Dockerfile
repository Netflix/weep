FROM scratch
ENV USER=docker
EXPOSE 9090
COPY weep /
ENTRYPOINT ["/weep"]
