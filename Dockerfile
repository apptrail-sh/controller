FROM gcr.io/distroless/static-debian11:nonroot

COPY "./apptrail" /usr/local/bin/apptrail

ENTRYPOINT ["apptrail"]
