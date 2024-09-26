FROM gcr.io/distroless/static-debian11:nonroot

COPY "./controller" /usr/local/bin/apptrail

ENTRYPOINT ["apptrail"]
