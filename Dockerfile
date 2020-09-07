# Start with a full-fledged golang image, but strip it to the final image.
FROM alpine:latest

COPY ./fibsrv ./fibsrv

CMD ["./fibsrv"]