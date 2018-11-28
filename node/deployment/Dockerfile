FROM openshift/base-centos7

WORKDIR /vpcloue

ADD node /vpcloue/
ADD config /vpcloue/
RUN chmod +x /vpcloue/node

EXPOSE 7575
EXPOSE 8080

CMD ["./node"]