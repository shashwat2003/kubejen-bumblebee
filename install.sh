#!/usr/bin/env bash

set -e

if [ -z "$BUMBLEBEE_ROOT" ]; then
    if [ -z "$HOME" ]; then
        printf "$0: %s\n" \
        "Either \$BUMBLEBEE_ROOT or \$HOME must be set to determine the install location." \
        >&2
        exit 1
    fi
    export BUMBLEBEE_ROOT="${HOME}/.kubejen/bumblebee"
fi

colorize() {
    if [ -t 1 ]; then printf "\e[%sm%s\e[m" "$1" "$2"
    else echo -n "$2"
    fi
}

# Checks for `.bumblebee` file, and suggests to remove it for installing
if [ -d "${BUMBLEBEE_ROOT}" ]; then
    { echo
        colorize 1 "WARNING"
        echo ": Can not proceed with installation. Kindly remove the '${BUMBLEBEE_ROOT}' directory first."
        echo
    } >&2
    exit 1
fi

failed_checkout() {
    echo "Failed to git clone $1"
    exit -1
}

checkout() {
    [ -d "$2" ] || git -c advice.detachedHead=0 clone --branch "$3" --depth 1 "$1" "$2" || failed_checkout "$1"
}

if ! command -v git 1>/dev/null 2>&1; then
    echo "bumblebee: Git is not installed, can't continue." >&2
    exit 1
fi

if ! command -v go 1>/dev/null 2>&1; then
    echo "bumblebee: go is not installed, can't continue." >&2
    exit 1
fi

if ! command -v docker 1>/dev/null 2>&1; then
    echo "bumblebee: docker is not installed, can't continue." >&2
    exit 1
fi

# Do not check ssh authentication if NO_SSH is present
if ! [ -n "${NO_SSH}" ]; then
    if ! command -v ssh 1>/dev/null 2>&1; then
        echo "bumblebee: configuration NO_SSH found but ssh is not installed, can't continue." >&2
        exit 1
    fi
    echo "Testing ssh authentication.."
    # `ssh -T git@github.com' returns 1 on success
    # See https://docs.github.com/en/authentication/connecting-to-github-with-ssh/testing-your-ssh-connection
    ssh -T git@github.com 1>/dev/null 2>&1 || EXIT_CODE=$?
    if [[ ${EXIT_CODE} != 1 ]]; then
        echo "bumblebee: github ssh authentication failed."
        echo
        echo "In order to use the ssh connection option, you need to have an ssh key set up."
        echo "Please generate an ssh key by using ssh-keygen, or follow the instructions at the following URL for more information:"
        echo
        echo "> https://docs.github.com/en/repositories/creating-and-managing-repositories/troubleshooting-cloning-errors#check-your-ssh-access"
        echo
        echo "Once you have an ssh key set up, try running the command again."
        exit 1
    fi
fi

if [ -n "${NO_SSH}" ]; then
    GITHUB="https://github.com/"
else
    GITHUB="git@github.com:"
fi

BUMBLEBEE_GIT_URL="shashwat2003/kubejen-bumblebee"

CONTAINER_NAME="bumblebee-build"

checkout "${GITHUB}${BUMBLEBEE_GIT_URL}.git" "${BUMBLEBEE_ROOT}" "${BUMBLEBEE_GIT_TAG:-main}"

{ echo
    colorize 3 "Starting Build.."
    echo
} >&2

if docker container ls -a | grep $CONTAINER_NAME 1>/dev/null
then
    docker rm $CONTAINER_NAME 1>/dev/null
fi

docker run -v $BUMBLEBEE_ROOT:/bumblebee --name bumblebee-build -e GOOS=$(go env GOHOSTOS) -e GOARCH=$(go env GOHOSTARCH) $(docker build -q $BUMBLEBEE_ROOT)

{ echo
    colorize 5 "Build Success.."
    echo
} >&2

try_profile() {
    if [ -z "${1-}" ] || [ ! -f "${1}" ]; then
        return 1
    fi
    echo "${1}"
}

detect_profile() {
    if [ "${PROFILE-}" = '/dev/null' ]; then
        # the user has specifically requested NOT to have nvm touch their profile
        return
    fi
    
    if [ -n "${PROFILE}" ] && [ -f "${PROFILE}" ]; then
        echo "${PROFILE}"
        return
    fi
    
    local DETECTED_PROFILE
    DETECTED_PROFILE=''
    
    if [ "${SHELL#*bash}" != "$SHELL" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            DETECTED_PROFILE="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
            DETECTED_PROFILE="$HOME/.bash_profile"
        fi
        elif [ "${SHELL#*zsh}" != "$SHELL" ]; then
        if [ -f "$HOME/.zshrc" ]; then
            DETECTED_PROFILE="$HOME/.zshrc"
            elif [ -f "$HOME/.zprofile" ]; then
            DETECTED_PROFILE="$HOME/.zprofile"
        fi
    fi
    
    if [ -z "$DETECTED_PROFILE" ]; then
        for EACH_PROFILE in ".profile" ".bashrc" ".bash_profile" ".zprofile" ".zshrc"
        do
            if DETECTED_PROFILE="$(try_profile "${HOME}/${EACH_PROFILE}")"; then
                break
            fi
        done
    fi
    
    if [ -n "$DETECTED_PROFILE" ]; then
        echo "$DETECTED_PROFILE"
    fi
}

if ! command -v bumblebee 1>/dev/null; then
    PROFILE="$(detect_profile)"
    SOURCE_STR="\\n# bumblebee\\nexport BUMBLEBEE_ROOT=\"${BUMBLEBEE_ROOT}\"\\nexport PATH=\"${BUMBLEBEE_ROOT}/bin:\$PATH\""
    if [ -z "${PROFILE-}" ] ; then
        TRIED_PROFILE=''
        if [ -n "${PROFILE}" ]; then
            TRIED_PROFILE="${PROFILE} (as defined in \$PROFILE), "
        fi
        echo "=> Profile not found. Tried ${TRIED_PROFILE-}~/.bashrc, ~/.bash_profile, ~/.zprofile, ~/.zshrc, and ~/.profile."
        echo "=> Create one of them and run this script again"
        echo "   OR"
        echo "=> Append the following lines to the correct file yourself:"
        command printf "${SOURCE_STR}"
        echo
    else
        if ! grep "BUMBLEBEE_ROOT" "$PROFILE" 1>/dev/null; then
            echo "=> Appending source string to $PROFILE"
            command printf "${SOURCE_STR}" >> "$PROFILE"
            echo "Bumblebee successfully installed!"
        else
            echo "=> source string already in ${PROFILE}"
        fi
    fi
else
    { echo
        colorize 1 "WARNING"
        echo ": seems you already have 'bumblebee' to the load path."
        echo
    } >&2
fi