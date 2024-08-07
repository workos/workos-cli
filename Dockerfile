FROM scratch
COPY ./bin/workos /bin/workos
ENTRYPOINT ["/bin/workos"]