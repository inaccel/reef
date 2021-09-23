FROM scratch
COPY reef /bin/reef
ENTRYPOINT ["reef"]
