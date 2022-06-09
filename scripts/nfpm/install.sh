#!/bin/sh

cleanInstall() {
    if ! getent passwd grendel > /dev/null; then
        printf "\033[32m Creating grendel system user & group\033[0m\n"
        groupadd -r grendel
        useradd -r -g grendel -d /var/lib/grendel -s /sbin/nologin \
                -c 'Grendel server' grendel
    fi

    mkdir -p /var/lib/grendel/images /var/lib/grendel/repo /var/lib/grendel/templates
    chown grendel:grendel /var/lib/grendel /var/lib/grendel/images /var/lib/grendel/repo /var/lib/grendel/templates
    chmod 755 /var/lib/grendel

    if [ -f "/etc/grendel/grendel.toml" ]; then
        chown grendel:grendel /etc/grendel/grendel.toml
    fi

    if [ -x "/usr/bin/deb-systemd-helper" ]; then
        printf "\033[32m Installing grendel.service systemd using deb\033[0m\n"
        deb-systemd-helper purge grendel.service >/dev/null
        deb-systemd-helper unmask grendel.service >/dev/null
    elif [ -x "/usr/bin/systemctl" ]; then
        printf "\033[32m Enable grendel.service in systemd\033[0m\n"
        systemctl daemon-reload ||:
        systemctl unmask grendel.service ||:
        systemctl preset grendel.service ||:
        systemctl enable grendel.service ||:
        systemctl restart grendel.service ||:
    fi
}

upgrade() {
    printf "\033[32m Upgrading grendel\033[0m\n"
}

# Step 2, check if this is a clean install or an upgrade
action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
  # Alpine linux does not pass args, and deb passes $1=configure
  action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
    # deb passes $1=configure $2=<current version>
    action="upgrade"
fi

case "$action" in
  "1" | "install")
    cleanInstall
    ;;
  "2" | "upgrade")
    printf "\033[32m Post Install of an upgrade\033[0m\n"
    upgrade
    ;;
  *)
    # $1 == version being installed
    printf "\033[32m Alpine\033[0m"
    cleanInstall
    ;;
esac

exit 0
