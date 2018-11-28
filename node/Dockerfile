FROM docker.vpclub.cn/hidevopsio/centos7:1.0
WORKDIR  ${HOME}/vpclub/
COPY config ${HOME}/vpclub/config
COPY hinode ${HOME}/vpclub/hinode
RUN  chmod +x ${HOME}/vpclub/hinode
ENV MAVEN_MIRROR_URL http://nexus.vpclub.cn/repository/maven-public/
ENV MAVEN_CLEAR_REPO false
ENV DOCKER_API_VERSION 1.24

EXPOSE 7575
EXPOSE 8080

USER 0

CMD ["./hinode"]
