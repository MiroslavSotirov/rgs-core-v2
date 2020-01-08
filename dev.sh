#!/bin/bash
set -e

_app=$(basename $(pwd))
_dir=$(cd "$(dirname "$0")"; pwd)

_base_version="1.13.2029"
_base_image="harbor.stargazer.com.sg/utils/rgs-v2-dev"
_version="${_base_version}.${_app}.local"
_utils_image="${_base_image}:${_version}"

_huid=$(id -u)
_hgid=$(id -g)
_dgid=$(id -g)
_husr=${USER}
_os="${OSTYPE}"
_name="${_app}-${_husr}"
_docker_config_json="${HOME}/.docker/config.json"

function _get_os() {
  case "$OSTYPE" in
    linux*)   echo "linux"   ;;
    darwin*)  echo "osx"     ;; 
    win*)     echo "win"     ;;
    msys*)    echo "msys"    ;;
    cygwin*)  echo "cygwin"  ;;
    bsd*)     echo "bsd"     ;;
    solaris*) echo "solaris" ;;
    *)        echo "unknown" ;;
  esac
}
_os=$( _get_os )
function _get_hgid() {
  if [[ "${_os}" == "osx" ]]; then
    echo "$(id -u)"
  else
    echo "$(id -g)"
  fi
}
_hgid=$( _get_hgid )
function _get_dgid() {
  if [[ "${_os}" == "osx" ]]; then
    # there is no need to mapped for docker group if its on osx. 
    echo "9999"
  else
    echo "$(getent group docker | cut -d: -f3)"
  fi
}
_dgid=$( _get_dgid )
function _set_docker_config_json() {
  if [[ "${_os}" == "osx" ]]; then
    echo "${HOME}/.docker/config.osx.json"
  else
    echo "${HOME}/.docker/config.json"
  fi
}
_docker_config_json="$( _set_docker_config_json )"
function _check() {
  if [[ "${_os}" == "osx" ]]; then
    local _docker_sock_mod="$( ls -l /var/run/docker.sock | awk '{k=0;for(i=0;i<=8;i++)k+=((substr($1,i+2,1)~/[rwx]/) *2^(8-i));if(k)printf("%0o ",k);print}' | cut -c 1-3 )"
    if [[ "${_docker_sock_mod}" != "755" ]]; then
      echo "Please ensure that permission os /var/run/docker.sock is 755"
      exit 0
    fi
  fi

  if [ -f "${_docker_config_json}" ]; then
    echo "Will make use of ${_docker_config_json} as docker /home/${_husr}/.docker/config.json"
  else 
    echo "Please ensure ${_docker_config_json} exists"
    exit 0
  fi
}
_check
function _init() {
  if [[ ! -d "${_dir}/.home.user.gradle" ]]; then
    mkdir -p ${_dir}/.home.user.gradle
  fi
}
_init
function _stop() {
  echo "stopping $1"
  bash -c "docker stop $1 || true" >> /dev/null 2>&1
  bash -c "docker rm $1 || true" >> /dev/null 2>&1
  bash -c "docker stop memcached-${_husr} || true" >> /dev/null 2>&1
  bash -c "docker rm memcached-${_husr} || true" >> /dev/null 2>&1
}
function _build() {
docker build --tag="${_utils_image}" -<<-EOF
       FROM ${_base_image}:${_base_version}
       RUN apt-get update -y && apt-get install -y libz-dev

       # generated sudoer file
       RUN echo "${_husr} ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/${_husr}
       RUN chmod 0440 /etc/sudoers.d/${_husr} 

       # generate .bashrc file
       RUN cp /root/.bashrc /tmp/bashrc
       RUN echo $'sudo chmod 777 /var/run/docker.sock\n\
                  export MCROUTER=memcached-${_husr}:11211\n\ 
                  cd /home/${_husr}/go/src/${_app}\n\
                  \n'\
           | awk 'sub(/^ */, "")' >> /tmp/bashrc

       RUN sudo echo "GO111MODULE=on" >> /etc/environment

       # add user
       RUN groupadd -g ${_dgid} dind && \
           groupadd -g ${_hgid} ${_husr} && \
           useradd -u ${_huid} -g ${_hgid} -G sudo -G docker -G dind -m ${_husr} -s /bin/bash && \
           mv /tmp/bashrc /home/${_husr}/.bashrc && \
           mkdir -p /home/${_husr}/.docker && \
           mkdir -p /home/${_husr}/go/src && \
           mkdir -p /home/${_husr}/go/pkg && \
           git clone https://github.com/fatih/vim-go.git /home/${_husr}/.vim/pack/plugins/start/vim-go && \
           chown -R ${_husr}:${_husr} /home/${_husr}
       CMD sudo su - ${_husr}
EOF
}
function _run() {
  docker run -d  --name="memcached-${_husr}" \
                 -p 11211:11211 \
                 --expose=11211 \
                 memcached \
  && \
  docker run -it --name=${_name} \
                 --hostname=${_name} \
                 --user=${_huid}:${_hgid} \
                 --add-host=${_name}:127.0.0.1 \
                 --link memcached-${_husr}:memcached-${_husr} \
                 -p 3000:3000 \
		 --expose=3000 \
                 --env MCROUTER=memcached-${_app}:11211 \
                 --volume=${_dir}:/home/${_husr}/go/src/${_app} \
                 --volume=${_dir}/.home.user.gradle:/home/${_husr}/.gradle \
                 --volume=/var/run/docker.sock:/var/run/docker.sock \
                 --volume=${_docker_config_json}:/home/${_husr}/.docker/config.json \
                 ${_utils_image}
}
function _command() {
  local _cmd="$1"

  if [[ -z "${_cmd}" ]]; then
    _cmd="run"
  fi

  case "${_cmd}" in
    bash)
      docker exec -it ${_name} bash
    ;;

    stop)
      _stop ${_name}
    ;;

    run)
      _stop ${_name} && \
      _build && \
      _run
    ;;

    *)
      echo "Usage: ./dev.sh OR ./dev.sh <cmd>"
      echo "cmd  : [ bash, stop, run ]"
      echo "       default cmd will be [run]"
    ;;
  esac
}
_command $1
