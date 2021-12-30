FROM onosproject/golang-build:v1 as build

ENV ONOS_ADAPTER_ROOT=/go/src/github.com/onosproject/gnmi-netconf-adapter
ENV CGO_ENABLED=0

RUN mkdir -p $ONOS_ADAPTER_ROOT/

COPY . $ONOS_ADAPTER_ROOT

RUN cd $ONOS_ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/gnmi_target ./cmd/gnmi_target


FROM alpine:3.11
RUN apk add bash openssl curl libc6-compat
ENV GNMI_PORT=10161
ENV GNMI_INSECURE_PORT=11161
ENV HOME=/home/gnmi-netconf-adapter
RUN mkdir $HOME
WORKDIR $HOME

COPY --from=build /go/bin/gn* /usr/local/bin/

COPY configs/target_configs target_configs
COPY tools/scripts scripts
COPY pkg/certs certs

RUN chmod +x ./scripts/run_targets.sh
CMD ["./scripts/run_targets.sh"]