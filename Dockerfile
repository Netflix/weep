FROM scratch
ENV USER=docker
EXPOSE 9091
COPY weep /
ENTRYPOINT ["/weep"]
